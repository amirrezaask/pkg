package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNoEntry      = errors.New("no entry")
	ErrEntryExpired = errors.New("entry expired")
)

type Cacher interface {
	Remember(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (any, error)
}
