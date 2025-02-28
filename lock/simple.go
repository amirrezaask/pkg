package lock

import (
	"context"
	"sync"

	"github.com/amirrezaask/pkg/errors"
)

type InMemoryLock struct {
	data sync.Map
}

func NewInMemoryLock() *InMemoryLock {
	return &InMemoryLock{
		data: sync.Map{},
	}
}

func (i *InMemoryLock) Lock(ctx context.Context, key string) error {
	if _, ok := i.data.Load(key); ok {
		return errors.Newf("cannot aquire lock for key('%s')", key)
	}

	i.data.Store(key, struct{}{})
	return nil
}

// implement Unlock method for InMemoryLock
func (i *InMemoryLock) Unlock(ctx context.Context, key string) error {
	i.data.Delete(key)
	return nil
}
