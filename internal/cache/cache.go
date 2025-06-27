package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"
)

type Cache interface {
	cacheKey(fnname string, fn interface{}, args ...interface{}) (string, error) //TODO: change interface to store/load albo set/get???
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key string, value string) error
	Invalidate(ctx context.Context, key string) error
}

type CacheConfig struct {
	Ttl     time.Duration
    KeyHash KeyHashType
}

func NewCache(config interface{}) Cache {
	switch c := config.(type) {
	case InMemoryCacheConfig:
		return NewInMemoryCache(c)
	case RedisCacheConfig:
		ctx := context.Background()
		return NewRedisCache(ctx, c)
	}
	return NewNoCache()
}

func getCallParameters(fn interface{}, args ...interface{}) (reflect.Value, []reflect.Value, error) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return reflect.Value{}, []reflect.Value{}, errors.New("Expected function parameter is not a function")
	}

	if len(args) != v.Type().NumIn() {
		return reflect.Value{}, []reflect.Value{}, errors.New("Incorrect number of arguments for the function")
	}

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		argVal := reflect.ValueOf(arg)
		if !argVal.Type().AssignableTo(v.Type().In(i)) {
			return reflect.Value{}, []reflect.Value{}, fmt.Errorf("Argument %d has incorrect type", i)
		}
		in[i] = argVal
	}
	return v, in, nil
}

func getReturnValues[T any](out []reflect.Value) (T, error) {
	var zero T
	if len(out) != 2 {
		return zero, fmt.Errorf("Function call returned incorrect number of values")
	}

	// Assert first return to T
	result, ok := out[0].Interface().(T)
	if !ok {
		return zero, fmt.Errorf("Type assertion to %T failed", zero)
	}

	// Assert second return to error
	errInterface := out[1].Interface()
	if errInterface != nil {
		err, ok := errInterface.(error)
		if !ok {
			return zero, errors.New("Second return value is not an error")
		}
		return zero, fmt.Errorf("Execution failed: %w", err)
	}
	return result, nil
}

func CallCached[T any](c Cache, fnname string, fn interface{}, args ...interface{}) (T, error) {
	ctx := context.Background()
	var zero T
	v, in, err := getCallParameters(fn, args...)
	if err != nil {
		return zero, fmt.Errorf("Cache: Get call parameters failed: %w", err)
	}

	key, err := c.cacheKey(fnname, fn, args...)
	if err != nil {
		return zero, fmt.Errorf("Cache: Could not create cache key.")
	}
	debug(c)
	fmt.Println("B: ", key)

	//TODO: IMPLEMENT CACHE MAGIC HERE
	// if cacheHit && inTtl {
	//   c.stat.hit = c.stat.hit + 1
	//   async out := v.Call(in) + storeCache //goroutine
	//   result = c.store[key].val
	// } else {
	//   c.stat.miss = c.stat.miss + 1
	//   out := v.Call(in) + storeCache
	//   ...
	//   result = out[0]
	// }
	// storeCache() -> { cacheEntry.ts = time.Now(), cacheEntryVal = result, c.store[key] = cacheEntry }
	//TODO: add mutex



	// try cache
	if s, ok, err := c.Get(ctx, key); err == nil && ok {
		// assume you marshal/unmarshal T ⇄ string elsewhere
		val, err := decode[T](s)
		if err == nil {
			return val, nil
		}
		_ = c.Invalidate(ctx, key) // poison-pill protection
	} else if err != nil {
		// decide whether to ignore or propagate cache errors
		return zero, err
	}




	// Call fn function
	out := v.Call(in)

	result, err := getReturnValues[T](out)
	if err != nil {
		return zero, fmt.Errorf("Cache: Return value error: %w", err)
	}





	if err == nil {
		if enc, encErr := encode(result); encErr == nil {
			_ = c.Set(ctx, key, enc) // expiration as you like  //TODO: implement TTL
		}
	}



	debug(c)




	return result, nil
}

func debug(c Cache) {
	if inMemCache, ok := c.(*InMemoryCache); ok { //TODO: implement and use stats
		//inMemCache.DoSomething()
		println(len(inMemCache.storage))
	} else {
		fmt.Println("Could not cast to ImplA")
	}
}

// encode marshals any value to a JSON string.
func encode[T any](v T) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encode: %w", err)
	}
	return string(b), nil
}

// decode unmarshals a JSON string back into a value of type T.
func decode[T any](s string) (T, error) {
	var v T
	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		return v, fmt.Errorf("decode: %w", err)
	}
	return v, nil
}
/*
	func (c *Cache) ClearStats() {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.stat.Hit = 0
		c.stat.Miss = 0
		c.stat.Expired = 0
	}

	func (c *Cache) Stats() Stat {
		c.mu.RLock()
		defer c.mu.RUnlock()

		return c.stat
	}

	func (c *Cache) StatsStr() string {
		c.mu.RLock()
		defer c.mu.RUnlock()

		var hmr float32
		total := c.stat.Hit + c.stat.Miss
		if total > 0 {
			hmr = float32(c.stat.Hit) / float32(total)
		}
		return fmt.Sprintf("hit: %d, miss: %d, h/(h+m): %f, expired: %d", c.stat.Hit, c.stat.Miss, hmr, c.stat.Expired)
	}
*/

func DecodeKey(key string, keyHash KeyHashType) (string, error) {
	if keyHash == KEY_HASH_BASE64 {
		return decodeBase64(key)
	} else if keyHash == KEY_HASH_HEX {
		return decodeHex(key)
	}
	return "", fmt.Errorf("Cache: Key hash '%s' could not be decoded", keyHash.String())
}

//TODO:
//When ttl is set to 0 skip cache. Also clear cache when setting ttl to 0 using SetTtl(..).
//    Consider remove of SetTtl(..), use config for ttl, when ttl is 0 in config return Cache -> NoCache
//Tests in internal/cache/{cache,key}_test.go

//- golang atomic for increments/set of stats
//- RWmutex for store usage
//- add expired in cache logic
//- add limit len() < config.Limit
//- solution for cleanup (interval?, alarm list?)

//- consider using RWmutex instead of atomic for stats (maybe Wmutex is already there when needed to increment hit/miss
//- consider key as type {val, keyHashType} with .String() and Parse(key string, keyHashType)  method
//- Remove all reflection from production, use reflection/custom asserts only in testing
//  Add string parameter to CacheCall ('calleeName') with the name of the function, check all calls in testing (do not include checks in production)
//  Consider go:generate checker in cmd/call_cached_check (remove or add)

// Add context to Get/Set in Cache interface and to CallCached



//NEW TODO:
// implement STATS and use for debug
// implement clean of cache by ttl




type Stat struct {
    Hit     int64
    Miss    int64
    Expired int64
}

/*type CacheBase struct {
	stat    Stat
    keyHash KeyHashType
	ttl     time.Duration
}*/

