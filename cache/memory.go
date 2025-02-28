package cache

import (
	"context"
	"time"
)

type memoryCacher struct {
	data map[string]cacheValue
}

type cacheValue struct {
	v                 any
	registeredTime    time.Time
	shouldExpireAfter *time.Duration
}

func NewMemoryCacher() Cacher {
	return &memoryCacher{
		data: map[string]cacheValue{},
	}
}

func (m *memoryCacher) Remember(ctx context.Context, key string, value any, ttl time.Duration) error {
	m.data[key] = cacheValue{
		v:                 value,
		registeredTime:    time.Now(),
		shouldExpireAfter: &ttl,
	}

	return nil
}
func (m *memoryCacher) Get(ctx context.Context, key string) (any, error) {
	value, ok := m.data[key]
	if !ok {
		return nil, ErrNoEntry
	}

	if value.shouldExpireAfter != nil {
		if value.registeredTime.Add(*value.shouldExpireAfter).Before(time.Now()) {
			delete(m.data, key)
			return nil, ErrEntryExpired
		}
	}

	return value.v, nil
}
