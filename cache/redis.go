package cache

import (
	"context"
	"time"

	"github.com/amirrezaask/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type redisCacher struct {
	*redis.Client
	ptrProvider func() any
}

func NewRedisCacher(client *redis.Client, ptrProviderFunc func() any) Cacher {
	return &redisCacher{
		Client:      client,
		ptrProvider: ptrProviderFunc,
	}
}

func (r *redisCacher) Remember(ctx context.Context, key string, value any, ttl time.Duration) error {
	return errors.Wrap(r.Set(ctx, key, value, ttl).Err(), "error in redis cacher")
}
func (r *redisCacher) Get(ctx context.Context, key string) (any, error) {
	cmd := r.Client.Get(ctx, key)

	ptr := r.ptrProvider()

	err := cmd.Scan(ptr)
	if err != nil {
		return nil, errors.Wrap(err, "error in redis cacher scan")
	}

	return ptr, nil
}
