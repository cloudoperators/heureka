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
	stat      Stat
	ttl       time.Duration
	keyHash   KeyHashType
	storage   map[string]*entry
	sizeLimit int
	mu        sync.RWMutex
}

type InMemoryCacheConfig struct {
	CacheConfig
	SizeLimit int
}

func NewInMemoryCache(config InMemoryCacheConfig) *InMemoryCache {
	inMemoryCache := &InMemoryCache{
		ttl:       config.Ttl,
		keyHash:   config.KeyHash,
		storage:   make(map[string]*entry),
		sizeLimit: config.SizeLimit,
	}
	return inMemoryCache
}

type entry struct {
	//ts  time.Time //TODO: consider
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

func (imc InMemoryCache) Get(_ context.Context, key string) (string, bool, error) { //TODO: improve + add in background request
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

func (imc InMemoryCache) Set(_ context.Context, key string, value string) error {
	imc.mu.Lock()
	defer imc.mu.Unlock()

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
