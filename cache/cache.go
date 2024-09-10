package cache

type Cacher interface {
	Remember(key string, value any) error
	Get(key string) (any, error)
}

type CacheSequel struct{}
type CacheRedis struct{}
