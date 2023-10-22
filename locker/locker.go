package locker

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type Locker struct {
	RDb *redis.Client
}

func New(rdb *redis.Client) *Locker {
	return &Locker{RDb: rdb}
}

func (x *Locker) Key(name string) string {
	return fmt.Sprintf(`locker:%s`, name)
}

func (x *Locker) Update(ctx context.Context, name string, ttl time.Duration) int64 {
	key := x.Key(name)
	count := x.RDb.Exists(ctx, key).Val()
	if count != 0 {
		return x.RDb.Incr(ctx, key).Val()
	}
	if x.RDb.Set(ctx, key, 1, ttl).Val() == "OK" {
		return 1
	}
	return 0
}

var (
	ErrLockerNotExists = errors.NewPublic("the locker does not exists")
	ErrLocked          = errors.NewPublic("tha locker to be locked")
)

func (x *Locker) Verify(ctx context.Context, name string, max int64) (err error) {
	key := x.Key(name)
	count := x.RDb.Exists(ctx, key).Val()
	if count == 0 {
		return ErrLockerNotExists
	}
	var n int64
	if n, err = x.RDb.Get(ctx, key).Int64(); err != nil {
		return
	}
	if n >= max {
		return ErrLocked
	}
	return
}

func (x *Locker) Delete(ctx context.Context, name string) int64 {
	return x.RDb.Del(ctx, x.Key(name)).Val()
}
