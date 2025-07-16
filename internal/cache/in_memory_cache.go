package cache

import (
	"context"
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type InMemoryCache struct {
	CacheBase
	gc *gocache.Cache
}

type InMemoryCacheConfig struct {
	CacheConfig
	CleanupInterval time.Duration
}

func NewInMemoryCache(config InMemoryCacheConfig) *InMemoryCache {
	cleanupInterval := config.CleanupInterval
	if cleanupInterval == 0 {
		cleanupInterval = gocache.NoExpiration
	}

	cacheBase := NewCacheBase(config.CacheConfig)
	inMemoryCache := &InMemoryCache{
		CacheBase: *cacheBase,
		gc:        gocache.New(1*time.Hour, cleanupInterval),
	}

	inMemoryCache.startMonitorIfNeeded(config.MonitorInterval)

	return inMemoryCache
}

func (imc InMemoryCache) Get(_ context.Context, key string) (string, bool, error) {
	val, found := imc.gc.Get(key)
	if !found {
		return "", false, nil
	}
	valStr, ok := val.(string)
	if !ok {
		return "", false, fmt.Errorf("Cache: Get value could not be converted to string")
	}
	return valStr, true, nil
}

func (imc InMemoryCache) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = gocache.NoExpiration
	}
	imc.gc.Set(key, value, ttl)
	return nil
}

func (imc InMemoryCache) Invalidate(_ context.Context, key string) error {
	imc.gc.Delete(key)
	return nil
}

func (imc *InMemoryCache) invalidateAll() {
	imc.gc.Flush()
}
