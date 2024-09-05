package database

import (
	"context"
	"github.com/amirrezaask/go-sith/errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// TODO: Add metrics here...

type Redis struct {
	*redis.Client
}

func (r *Redis) SimpleUnlock(ctx context.Context, key string) error {
	return r.Del(ctx, key).Err()
}
func (r *Redis) SimpleLock(ctx context.Context, key string, dur time.Duration) error {
	count, err := r.Exists(ctx, key).Result()
	if err != nil {
		return err
	}

	if count != 0 {
		return errors.Newf("cannot aquire simple lock for key('%s')", key)
	}

	return r.Set(ctx, key, 1, dur).Err()
}

func NewRedis(ctx context.Context, hostPort string, db int,
	username string, password string) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     hostPort,
		DB:       db,
		Username: username,
		Password: password,
	})
	statusCmd := client.Ping(ctx)
	if err := statusCmd.Err(); err != nil {
		return nil, err
	}
	return &Redis{client}, nil
}
