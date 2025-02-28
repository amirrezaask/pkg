package lock

import (
	"context"
	"time"

	"github.com/amirrezaask/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type DistributedLock struct {
	*redis.Client
}

func (r *DistributedLock) Unlock(ctx context.Context, key string) error {
	return r.Del(ctx, key).Err()
}
func (r *DistributedLock) Lock(ctx context.Context, key string, dur time.Duration) error {
	count, err := r.Exists(ctx, key).Result()
	if err != nil {
		return err
	}

	if count != 0 {
		return errors.Newf("cannot aquire distributed lock for key('%s')", key)
	}

	return r.Set(ctx, key, 1, dur).Err()
}
