package cache

import (
	"context"
	"fmt"

	"github.com/valkey-io/valkey-go"
)

type ValkeyCache struct {
	CacheBase
	client valkey.Client
}

type ValkeyCacheConfig struct {
	CacheConfig
	Url      string
	Password string
	Db       int
}

func NewValkeyCache(ctx context.Context, config ValkeyCacheConfig) *ValkeyCache {
	cacheBase := NewCacheBase(config.CacheConfig)
	valkeyClient, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{config.Url},
		Password:    config.Password})
	if err != nil {
		return &ValkeyCache{}
	}
	valkeyCache := &ValkeyCache{
		CacheBase: *cacheBase,
		client:    valkeyClient,
	}

	valkeyCache.startMonitorIfNeeded(config.MonitorInterval)

	_ = valkeyCache.invalidateAll(ctx)
	return valkeyCache
}

func (vc ValkeyCache) cacheKey(fnname string, fn interface{}, args ...interface{}) (string, error) {
	key, err := cacheKeyJson(fnname, fn, args...)
	if err != nil {
		return "", fmt.Errorf("Cache: could not create json cache key.")
	}
	return vc.encodeKey(key), nil
}

func (vc *ValkeyCache) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := vc.client.Do(ctx, vc.client.B().Get().Key(key).Build()).ToString()
	if err == valkey.Nil {
		return "", false, nil // miss
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

// ttl = 0 <- infinite
func (vc *ValkeyCache) Set(ctx context.Context, key string, value string) error {
	return vc.client.Do(ctx, vc.client.B().Set().Key(key).Value(value).Px(vc.ttl).Build()).Error()
}

func (vc *ValkeyCache) Invalidate(ctx context.Context, key string) error {
	return vc.client.Do(ctx, vc.client.B().Del().Key(key).Build()).Error()
}

func (vc *ValkeyCache) invalidateAll(ctx context.Context) error {
	return vc.client.Do(ctx, vc.client.B().Flushall().Build()).Error()
}

func (vc ValkeyCache) encodeKey(key string) string {
	if vc.keyHash == KEY_HASH_SHA256 {
		return encodeSHA256(key)
	} else if vc.keyHash == KEY_HASH_SHA512 {
		return encodeSHA512(key)
	} else if vc.keyHash == KEY_HASH_HEX {
		return encodeHex(key)
	} else if vc.keyHash == KEY_HASH_NONE {
		return key
	}
	return encodeBase64(key)
}
