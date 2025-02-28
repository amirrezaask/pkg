package lock

import (
	"context"
)

type Locker interface {
	Lock(ctx context.Context, key string) error
	Unlock(ctx context.Context, key string) error
}

