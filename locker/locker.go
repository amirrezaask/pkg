package locker

type Locker interface {
	Lock(key string) error
	Unlock(key string) error
}
