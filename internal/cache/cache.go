package cache

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"time"
)

type Cache struct {
	stat  Stat
	store map[string]*CacheEntry
	ttl   time.Duration
}

type Stat struct {
	hit  int
	miss int
}

type Config struct {
	Ttl time.Duration
}

type CacheEntry struct {
	t0 time.Time
	val interface{}
}

func NewCache(config Config) *Cache {
	c := &Cache{
		store: make(map[string]*CacheEntry),
		ttl:   config.Ttl,
	}
	return c
}

// TODO: idk why this is not working:
//func (c *Cache)CallCached[T any](fn interface{}, args ...interface{}) (T, error) {
func CallCached[T any](c *Cache, fn interface{}, args ...interface{}) (T, error) {
	var zero T
	v := reflect.ValueOf(fn)

	// Check fn is a function
	if v.Kind() != reflect.Func {
		return zero, errors.New("Cache: first parameter is not a function")
	}

	// Check fn has expected number of parameters
	if len(args) != v.Type().NumIn() {
		return zero, errors.New("Cache: incorrect number of arguments")
	}

	// Check fn has expected types of parameters
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		argVal := reflect.ValueOf(arg)
		if !argVal.Type().AssignableTo(v.Type().In(i)) {
			return zero, fmt.Errorf("Cache: argument %d has incorrect type", i)
		}
		in[i] = argVal
	}

	fName := getFunctionName(fn)
	fmt.Println("A: ", fName)
	key, err := cacheKey(fn, args...)
	if err != nil {
		return zero, fmt.Errorf("Cache: could not create cache key.")
	}
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
	// storeCache() -> { cacheEntry.t0 = time.Now(), cacheEntryVal = result, c.store[key] = cacheEntry }
	//TODO: add mutex

	// Call fn function
	out := v.Call(in)

	// Handle return values
	if len(out) != 2 {
		return zero, fmt.Errorf("Cache: Function call returned incorrect number of values")
	}

	// Assert first return to T
	result, ok := out[0].Interface().(T)
	if !ok {
		return zero, fmt.Errorf("Cache: type assertion to %T failed", zero)
	}

	// Assert second return to error
	errInterface := out[1].Interface()
	if errInterface != nil {
		err, ok := errInterface.(error)
		if !ok {
			return zero, errors.New("Cache: second return value is not an error")
		}
		return zero, err
	}

	return result, nil
}

func getFunctionName(fn interface{}) string {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return "<not a function>"
	}

	return runtime.FuncForPC(v.Pointer()).Name()
}

func cacheKeyJson(fn interface{}, args ...interface{}) (string, error) {
	keyParts := make([]interface{}, 0, len(args)+1)

	fnName := getFunctionName(fn)
	keyParts = append(keyParts, fnName)

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

func cacheKey(fn interface{}, args ...interface{}) (string, error) {
	key, err := cacheKeyJson(fn, args...)
	if err != nil {
		return "", fmt.Errorf("Cache: could not create json cache key.")
	}
	return encodeBase64(key), nil
}

func isJSONSerializable(val interface{}) bool {
	_, err := json.Marshal(val)
	return err == nil
}

func encodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func decodeBase64(encoded string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}

func (c *Cache) ClearStats() {
	c.stat.hit = 0
	c.stat.miss = 0
}

func (c *Cache) Stats() Stat {
	return c.stat
}

func (c *Cache) StatsStr() string {
	total := c.stat.hit + c.stat.miss
	if total > 0 {
		return fmt.Sprintf("hit: %d, miss: %d, h/(h+m): %f", c.stat.hit, c.stat.miss, float32(c.stat.hit)/float32(total))
	}
	return "hit: 0, miss: 0, h/(h+m): N/A"
}

func (c *Cache) Ttl(t time.Duration) {
	c.ttl = t
}

func (c Cache) GetKeys() []string {
	return getSortedMapKeys(c.store)
}

func getMapKeys(m map[string]*CacheEntry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getSortedMapKeys(m map[string]*CacheEntry) []string {
	keys := getMapKeys(m)
	sort.Strings(keys)
	return keys
}

func DecodeKey(k string) (string, error) {
	return decodeBase64(k)
}

func (c *Cache) InvalidateCache() {
	c.store = make(map[string]*CacheEntry)
}

//TODO:
//Consider Cache object per app object handler
//Tests in internal/cache/cache_test.go
