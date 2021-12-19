package mcache

import (
	"context"
	"time"

	"github.com/OrlovEvgeny/go-mcache"

	"github.com/woodpecker-ci/woodpecker/server/cache"
)

type mem struct {
	cache *mcache.CacheDriver
}

// make sure mcache match interface
var _ cache.Cache = &mem{}

func NewMCache() cache.Cache {
	return &mem{
		cache: mcache.New(),
	}
}

func (m mem) Set(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	return m.cache.Set(key, value, expiration)
}

func (m mem) Get(_ context.Context, key string, objStore *interface{}) (bool, error) {
	item, exist := m.cache.Get(key)
	*objStore = item
	return exist, nil
}

func (m mem) Del(_ context.Context, keys ...string) error {
	for i := range keys {
		if err := m.cache.Set(keys[i], nil, time.Nanosecond); err != nil {
			return err
		}
	}
	return nil
}
