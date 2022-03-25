package cache

import (
	"context"
	"time"
)

type InMemoryCache struct {
	cache map[string]cacheValue
}

type cacheValue struct {
	value  []byte
	expiry *time.Time
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		cache: make(map[string]cacheValue),
	}
}

func (i *InMemoryCache) InvalidateValue(ctx context.Context, key string) error {
	delete(i.cache, key)
	return nil
}

func (i *InMemoryCache) StoreValue(ctx context.Context, key string, value []byte) error {
	i.cache[key] = cacheValue{
		value:  value,
		expiry: nil,
	}

	return nil
}

func (i *InMemoryCache) StoreValueWithExpiry(ctx context.Context, key string, value []byte, expiresIn time.Duration) error {
	expiry := time.Now().Add(expiresIn)

	i.cache[key] = cacheValue{
		value:  value,
		expiry: &expiry,
	}

	return nil
}

func (i *InMemoryCache) RetrieveValue(ctx context.Context, key string) ([]byte, error) {
	if value, ok := i.cache[key]; ok {
		if value.expiry != nil {
			if time.Now().After(*value.expiry) {
				delete(i.cache, key)
				return nil, ErrCacheMiss
			} else {
				return value.value, nil
			}
		} else {
			return value.value, nil
		}

	} else {
		return nil, ErrCacheMiss
	}
}
