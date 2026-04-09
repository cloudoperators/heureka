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
	CacheKey(fnname string, fn any, args ...any) (string, error)
	Get(key string) (string, bool, error)
	Set(key string, value string, ttl time.Duration) error
	Invalidate(key string) error
	IncHit()
	IncMiss()
	IncShared()
	GetStat() Stat
	LaunchRefresh(fn func())
	GetSingleflightWrapper() SingleflightWrapper
}

type SingleflightWrapper interface {
	Do(string, func() (any, error)) (any, error, bool)
}

func NewCache(ctx context.Context, wg *sync.WaitGroup, config any) Cache {
	switch c := config.(type) {
	case InMemoryCacheConfig:
		return NewInMemoryCache(ctx, wg, c)
	case ValkeyCacheConfig:
		return NewValkeyCache(ctx, wg, c)
	}

	return nil
}

func getCallParameters(fn any, args ...any) (reflect.Value, []reflect.Value, error) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return reflect.Value{}, []reflect.Value{}, errors.New(
			"expected function parameter is not a function",
		)
	}

	if len(args) != v.Type().NumIn() {
		return reflect.Value{}, []reflect.Value{}, errors.New(
			"incorrect number of arguments for the function",
		)
	}

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		argVal := reflect.ValueOf(arg)
		if !argVal.Type().AssignableTo(v.Type().In(i)) {
			return reflect.Value{}, []reflect.Value{}, fmt.Errorf(
				"argument %d has incorrect type",
				i,
			)
		}

		in[i] = argVal
	}

	return v, in, nil
}

func getReturnValues[T any](out []reflect.Value) (T, error) {
	var zero T
	if len(out) != 2 {
		return zero, fmt.Errorf("function call returned incorrect number of values")
	}

	// Assert first return to T
	result, ok := out[0].Interface().(T)
	if !ok {
		return zero, fmt.Errorf("type assertion to %T failed", zero)
	}

	// Assert second return to error
	errInterface := out[1].Interface()
	if errInterface != nil {
		err, ok := errInterface.(error)
		if !ok {
			return zero, errors.New("second return value is not an error")
		}

		return zero, fmt.Errorf("execution failed: %w", err)
	}

	return result, nil
}

func CallCached[T any](c Cache, ttl time.Duration, fnname string, fn any, args ...any) (T, error) {
	if c == nil {
		return callDisabled[T](fn, args...)
	}

	return callEnabled[T](c, ttl, fnname, fn, args...)
}

func callEnabled[T any](c Cache, ttl time.Duration, fnname string, fn any, args ...any) (T, error) {
	var zero T

	v, in, err := getCallParameters(fn, args...)
	if err != nil {
		return zero, fmt.Errorf("cache (param): Get call parameters failed: %w", err)
	}

	key, err := c.CacheKey(fnname, fn, args...)
	if err != nil {
		return zero, fmt.Errorf("cache (key): Could not create cache key")
	}

	if s, ok, err := c.Get(key); err == nil && ok {
		val, err := decode[T](s)
		if err == nil {
			c.IncHit()
			c.LaunchRefresh(func() {
				_, _ = callFn[T](c, ttl, key, v, in)
			})

			return val, nil
		}

		_ = c.Invalidate(key) // poison-pill protection
	} else if err != nil {
		return zero, fmt.Errorf("cache (get): %w", err)
	}

	c.IncMiss()

	return callFn[T](c, ttl, key, v, in)
}

func callDisabled[T any](fn any, args ...any) (T, error) {
	var zero T

	v, in, err := getCallParameters(fn, args...)
	if err != nil {
		return zero, fmt.Errorf("noCache (param): Get call parameters failed: %w", err)
	}

	out := v.Call(in)

	return getReturnValues[T](out)
}

func callFn[T any](
	c Cache,
	ttl time.Duration,
	key string,
	v reflect.Value,
	in []reflect.Value,
) (T, error) {
	var zero T

	fnExecuted := false
	untypedResult, err, shared := c.GetSingleflightWrapper().Do(key, func() (any, error) {
		fnExecuted = true
		out := v.Call(in)

		result, err := getReturnValues[T](out)
		if err != nil {
			return zero, fmt.Errorf("cache (fcall): Return value error: %w", err)
		} else if enc, encErr := encode(result); encErr == nil {
			err = c.Set(key, enc, ttl)
			if err != nil {
				return zero, fmt.Errorf("cache (set): %w", err)
			}
		} else {
			return zero, fmt.Errorf("cache (encode): %w", err)
		}

		return result, nil
	})

	if shared && !fnExecuted {
		c.IncShared()
	}

	if err != nil {
		return zero, err
	}

	result, ok := untypedResult.(T)
	if !ok {
		return zero, fmt.Errorf("cache (type): cannot cast result to expected type")
	}

	return result, nil
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
	switch keyHash {
	case KEY_HASH_BASE64:
		return decodeBase64(key)
	case KEY_HASH_HEX:
		return decodeHex(key)
	case KEY_HASH_NONE:
		return key, nil
	}

	return "", fmt.Errorf("cache: Key hash '%s' could not be decoded", keyHash.String())
}
