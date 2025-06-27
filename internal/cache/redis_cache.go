package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	rdb     *redis.Client
    keyHash KeyHashType
	ttl     time.Duration
}

type RedisCacheConfig struct {
	CacheConfig
	Url      string
	Password string
	Db       int
}

func NewRedisCache(ctx context.Context, config RedisCacheConfig) *RedisCache {
	redisCache := &RedisCache{
		rdb: redis.NewClient(&redis.Options{
			Addr:     config.Url,
			Password: config.Password,
			DB:       config.Db,
		}),
		ttl: config.Ttl,
	}
	_ = redisCache.invalidateAll(ctx)
	return redisCache
}

func (rc RedisCache) cacheKey(fnname string, fn interface{}, args ...interface{}) (string, error) {
    key, err := cacheKeyJson(fnname, fn, args...)
    if err != nil {
        return "", fmt.Errorf("Cache: could not create json cache key.")
    }
    return rc.encodeKey(key), nil
}

func (rc *RedisCache) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := rc.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil // miss
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

//ttl = 0 <- infinite
func (rc *RedisCache) Set(ctx context.Context, key string, value string) error {
	return rc.rdb.Set(ctx, key, value, rc.ttl).Err()
}

func (rc *RedisCache) Invalidate(ctx context.Context, key string) error {
	return rc.rdb.Del(ctx, key).Err()
}

func (rc *RedisCache)invalidateAll(ctx context.Context) error {
	return rc.rdb.FlushAll(ctx).Err()
}

func (rc RedisCache) encodeKey(key string) string {
    if rc.keyHash == KEY_HASH_SHA256 {
        return encodeSHA256(key)
    } else if rc.keyHash == KEY_HASH_SHA512 {
        return encodeSHA512(key)
    } else if rc.keyHash == KEY_HASH_HEX {
        return encodeHex(key)
    }
    return encodeBase64(key)
}
