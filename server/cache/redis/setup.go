package redis

import (
	"context"
	"encoding/json"
	go_redis "github.com/go-redis/redis/v8"
	"time"

	"github.com/woodpecker-ci/woodpecker/server/cache"
)

type redis struct {
	cache *go_redis.Client
}

// make sure mcache match interface
var _ cache.Cache = &redis{}

func NewMCache() cache.Cache {
	return &redis{
		cache: go_redis.NewClient(&go_redis.Options{
			Addr: "localhost:6379",
		}),
	}
}

func (r redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.cache.Set(ctx, key, b, expiration).Err()
}

func (r redis) Get(ctx context.Context, key string, objStore *interface{}) (bool, error) {
	cmd := r.cache.Get(ctx, key)
	if err := cmd.Err(); err != nil {
		return false, err
	}

	if len(cmd.Val()) == 0 {
		return false, nil
	}

	if err := json.Unmarshal([]byte(cmd.Val()), objStore); err != nil {
		return false, err
	}

	return true, nil
}

func (r redis) Del(ctx context.Context, keys ...string) error {
	return r.cache.Del(ctx, keys...).Err()
}
