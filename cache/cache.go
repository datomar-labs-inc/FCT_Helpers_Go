// Package cache provides a cache abstraction that can be used to easily swap cache implementations
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"
)

var ErrCacheMiss = errors.New("cache miss")
var ErrNilValue = errors.New("nil value")

type Cache interface {
	StoreValue(ctx context.Context, key string, value []byte) error
	StoreValueWithExpiry(ctx context.Context, key string, value []byte, expiresIn time.Duration) error
	RetrieveValue(ctx context.Context, key string) ([]byte, error)
	InvalidateValue(ctx context.Context, key string) error
}

type CachedValue = any

// GetCachedJSONValue will retrieve a value from a Cache, and parse it as json
func GetCachedJSONValue[T CachedValue](ctx context.Context, cache Cache, key string, getVal func(ctx context.Context) (*T, error)) (*T, error) {
	return GetCachedJSONValueWithExpiry(ctx, cache, key, getVal, nil)
}

// GetCachedJSONValueWithExpiry will retrieve a value from a Cache, and parse it as json
func GetCachedJSONValueWithExpiry[T CachedValue](ctx context.Context, cache Cache, key string, getVal func(ctx context.Context) (*T, error), expiresIn *time.Duration) (*T, error) {
	jsonBytes, err := cache.RetrieveValue(ctx, key)
	if err != nil && err != ErrCacheMiss {
		return nil, ferr.Wrap(err)
	} else if err == ErrCacheMiss {
		value, err := getVal(ctx)
		if err != nil {
			return nil, ferr.Wrap(err)
		}

		if isNil(value) {
			return nil, ErrNilValue
		}

		newValueJSONBytes, err := json.Marshal(value)
		if err != nil {
			return value, ferr.Wrap(err)
		}

		if expiresIn != nil {
			err = cache.StoreValueWithExpiry(ctx, key, newValueJSONBytes, *expiresIn)
			if err != nil {
				return value, ferr.Wrap(err)
			}
		} else {
			err = cache.StoreValue(ctx, key, newValueJSONBytes)
			if err != nil {
				return value, ferr.Wrap(err)
			}
		}

		return value, nil
	} else {
		var val T

		err := json.Unmarshal(jsonBytes, &val)
		if err != nil {
			return nil, ferr.Wrap(err)
		}

		return &val, nil
	}
}

// GetCachedValueWithExpiry will retrieve a value from a Cache, and parse it as json
func GetCachedValueWithExpiry(ctx context.Context, cache Cache, key string, getVal func(ctx context.Context) ([]byte, error), expiresIn *time.Duration) ([]byte, error) {
	valueBytes, err := cache.RetrieveValue(ctx, key)
	if err != nil && err != ErrCacheMiss {
		return nil, ferr.Wrap(err)
	} else if err == ErrCacheMiss {
		value, err := getVal(ctx)
		if err != nil {
			return nil, ferr.Wrap(err)
		}

		if isNil(value) {
			return nil, ErrNilValue
		}

		if expiresIn != nil {
			err = cache.StoreValueWithExpiry(ctx, key, value, *expiresIn)
			if err != nil {
				return value, ferr.Wrap(err)
			}
		} else {
			err = cache.StoreValue(ctx, key, value)
			if err != nil {
				return value, ferr.Wrap(err)
			}
		}

		return value, nil
	} else {
		return valueBytes, nil
	}
}

func isNil(i any) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
