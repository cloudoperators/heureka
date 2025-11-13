// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type Cache interface {
	CacheKey(fnname string, fn interface{}, args ...interface{}) (string, error)
	Get(key string) (string, bool, error)
	Set(key string, value string, ttl time.Duration) error
	Invalidate(key string) error
	IncHit()
	IncMiss()
	GetStat() Stat
	LaunchRefresh(fn func())
}

func NewCache(ctx context.Context, wg *sync.WaitGroup, config interface{}) Cache {
	switch c := config.(type) {
	case InMemoryCacheConfig:
		return NewInMemoryCache(ctx, wg, c)
	case ValkeyCacheConfig:
		return NewValkeyCache(ctx, wg, c)
	}
	return nil
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

func CallCached[T any](c Cache, ttl time.Duration, fnname string, fn interface{}, args ...interface{}) (T, error) {
	if c == nil {
		return callDisabled[T](fn, args...)
	}
	return callEnabled[T](c, ttl, fnname, fn, args...)
}

func callEnabled[T any](c Cache, ttl time.Duration, fnname string, fn interface{}, args ...interface{}) (T, error) {
	var zero T
	v, in, err := getCallParameters(fn, args...)
	if err != nil {
		return zero, fmt.Errorf("Cache (param): Get call parameters failed: %w", err)
	}

	key, err := c.CacheKey(fnname, fn, args...)
	if err != nil {
		return zero, fmt.Errorf("Cache (key): Could not create cache key.")
	}

	if s, ok, err := c.Get(key); err == nil && ok {
		val, err := decode[T](s)
		if err == nil {
			c.IncHit()
			c.LaunchRefresh(func() {
				out := v.Call(in)
				result, err := getReturnValues[T](out)
				if err == nil {
					if enc, encErr := encode(result); encErr == nil {
						_ = c.Set(key, enc, ttl)
					}
				}
			})
			return val, nil
		}
		_ = c.Invalidate(key) // poison-pill protection
	} else if err != nil {
		return zero, fmt.Errorf("Cache (get): %w", err)
	}

	c.IncMiss()
	out := v.Call(in)

	result, err := getReturnValues[T](out)
	if err != nil {
		return zero, fmt.Errorf("Cache (fcall): Return value error: %w", err)
	} else if enc, encErr := encode(result); encErr == nil {
		err = c.Set(key, enc, ttl)
		if err != nil {
			return zero, fmt.Errorf("Cache (set): %w", err)
		}
	} else {
		return zero, fmt.Errorf("Cache (encode): %w", err)
	}

	return result, nil
}

func callDisabled[T any](fn interface{}, args ...interface{}) (T, error) {
	var zero T
	v, in, err := getCallParameters(fn, args...)
	if err != nil {
		return zero, fmt.Errorf("NoCache (param): Get call parameters failed: %w", err)
	}
	out := v.Call(in)
	return getReturnValues[T](out)
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

func DecodeKey(key string, keyHash KeyHashType) (string, error) {
	if keyHash == KEY_HASH_BASE64 {
		return decodeBase64(key)
	} else if keyHash == KEY_HASH_HEX {
		return decodeHex(key)
	} else if keyHash == KEY_HASH_NONE {
		return key, nil
	}
	return "", fmt.Errorf("Cache: Key hash '%s' could not be decoded", keyHash.String())
}
