package test

import (
	"time"

	"github.com/amirrezaask/go-sith/database"

	"github.com/go-redis/redismock/v9"
)

type redisMock struct {
	redismock.ClientMock
}

// TODO: a way of reporting these errors better than the current shit
func (r *redisMock) ExpectLockForKeyWithDuration(key string, dur time.Duration) {
	r.Regexp().ExpectExists(key).SetVal(0)
	r.Regexp().ExpectSet(key, 1, dur).SetVal("OK")
}
func (r *redisMock) ExpectUnlockForKey(key string) {
	r.Regexp().ExpectDel(key).SetVal(1)
}

func (t *T) Redis(target **database.Redis) *redisMock {
	client, mock := redismock.NewClientMock()
	*target = &database.Redis{client}
	return &redisMock{mock}
}
