package cache

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

func CallCached[T any](fn interface{}, args ...interface{}) (T, error) {
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
	key, err := makeCacheKey(fn, args...)
	if err != nil {
		return zero, fmt.Errorf("Cache: could not create json cache key.")
	}
	fmt.Println("B: ", key)
	fmt.Println("C: ", encodeBase64(key))

	//TODO: IMPLEMENT CACHE MAGIC HERE

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

func makeCacheKey(fn interface{}, args ...interface{}) (string, error) {
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

func isJSONSerializable(val interface{}) bool {
	_, err := json.Marshal(val)
	return err == nil
}

func encodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

//TODO:
//struct Cache {
// stat Stat
// ttl time.Duration
//}
//struct Stat {
// hit int
// miss int
//}
//func NewCache(config Config) *Cache  { c.ClearStats() }
//func CallCached -> func (c *Cache)CallCached(...)
//func (c *Cache)ClearStats() { c.hit = 0, c.miss = 0 }
//func (c Cache)StatsStr() string
//func (c Cache)stats() Stat
//func (c *Cache)SetTtl(t time.Duration)
//func (c Cache)GetKeys() []string
//func DecodeKey(k string) string
//func (c *Cache)InvalidateCache()

//Consider Cache object per app object handler
