package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

type InMemoryCache struct {
	CacheBase
	storage         map[string]*entry
	sizeLimit       int
	mu              sync.RWMutex
	cleanupStopChan <-chan struct{}
}

type InMemoryCacheConfig struct {
	CacheConfig
	SizeLimit       int
	CleanupInterval time.Duration
}

func NewInMemoryCache(config InMemoryCacheConfig) *InMemoryCache {
	inMemoryCache := &InMemoryCache{
		CacheBase: CacheBase{
			ttl:     config.Ttl,
			keyHash: config.KeyHash,
		},
		storage:   make(map[string]*entry),
		sizeLimit: config.SizeLimit,
	}

	inMemoryCache.startCleanupIfNeeded(config.CleanupInterval)

	return inMemoryCache
}

type entry struct {
	exp time.Time
	val string
}

func (imc InMemoryCache) cacheKey(fnname string, fn interface{}, args ...interface{}) (string, error) {
	key, err := cacheKeyJson(fnname, fn, args...)
	if err != nil {
		return "", fmt.Errorf("Cache: could not create json cache key.")
	}
	return imc.encodeKey(key), nil
}

func (imc InMemoryCache) GetKeys() []string {
	return getSortedMapKeys(imc.storage)
}

func (imc InMemoryCache) Get(_ context.Context, key string) (string, bool, error) {
	imc.mu.RLock()
	e, ok := imc.storage[key]
	imc.mu.RUnlock()

	if !ok || (e.exp.IsZero() == false && time.Now().After(e.exp)) {
		if ok { // stale entry – purge lazily
			imc.mu.Lock()
			delete(imc.storage, key)
			imc.mu.Unlock()
		}
		return "", false, nil
	}
	return e.val, true, nil
}

func removeOldest(m map[string]*entry) {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, e := range m {
		if first || e.exp.Before(oldestTime) {
			oldestKey = key
			oldestTime = e.exp
			first = false
		}
	}

	if !first {
		delete(m, oldestKey)
	}
}

func (imc InMemoryCache) Set(_ context.Context, key string, value string) error {
	imc.mu.Lock()
	defer imc.mu.Unlock()

	if !imc.hasSpace() {
		removeOldest(imc.storage)
	}

	e := entry{val: value}
	if imc.ttl > 0 {
		e.exp = time.Now().Add(imc.ttl)
	}
	imc.storage[key] = &e
	return nil
}

func (imc InMemoryCache) Invalidate(_ context.Context, key string) error {
	imc.mu.Lock()
	defer imc.mu.Unlock()
	delete(imc.storage, key)
	return nil
}

func (imc *InMemoryCache) invalidateAll() {
	imc.storage = make(map[string]*entry)
}

func (imc InMemoryCache) hasSpace() bool {
	if imc.sizeLimit > 0 && len(imc.storage) >= imc.sizeLimit {
		return false
	}
	return true
}

func (imc InMemoryCache) encodeKey(key string) string {
	if imc.keyHash == KEY_HASH_SHA256 {
		return encodeSHA256(key)
	} else if imc.keyHash == KEY_HASH_SHA512 {
		return encodeSHA512(key)
	} else if imc.keyHash == KEY_HASH_HEX {
		return encodeHex(key)
	} else if imc.keyHash == KEY_HASH_NONE {
		return key
	}
	return encodeBase64(key)
}

func getSortedMapKeys(m map[string]*entry) []string {
	keys := getMapKeys(m)
	sort.Strings(keys)
	return keys
}

func getMapKeys(m map[string]*entry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func cacheKeyJson(fnname string, fn interface{}, args ...interface{}) (string, error) {
	keyParts := make([]interface{}, 0, len(args)+1)
	keyParts = append(keyParts, fnname)

	for i, arg := range args {
		if !isJSONSerializable(arg) {
			return "", fmt.Errorf("argument %d is not JSON serializable: %T", i, arg)
		}
		keyParts = append(keyParts, arg)
	}

	// Encode the full key as JSON array
	jsonKey, err := json.Marshal(keyParts)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cache key: %w", err)
	}

	return string(jsonKey), nil
}

func isJSONSerializable(val interface{}) bool {
	_, err := json.Marshal(val)
	return err == nil
}

func (imc *InMemoryCache) startCleanupIfNeeded(interval time.Duration) {
	if interval != 0 {
		imc.cleanupStopChan = make(chan struct{})
		imc.startCleanup(interval)
	}
}

func (imc *InMemoryCache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				imc.removeExpired()
			case <-imc.cleanupStopChan:
				return // exit goroutine when stop signal is received
			}
		}
	}()
}

func (imc *InMemoryCache) removeExpired() {
	now := time.Now()
	imc.mu.Lock()
	defer imc.mu.Unlock()
	for key, e := range imc.storage {
		if e.exp.Before(now) {
			delete(imc.storage, key)
		}
	}
}
