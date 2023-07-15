package locker

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type Locker struct {
	Namespace string
	RDb       *redis.Client
}

func New(options ...Option) *Locker {
	x := new(Locker)
	for _, v := range options {
		v(x)
	}
	return x
}

type Option func(x *Locker)

func SetNamespace(v string) Option {
	return func(x *Locker) {
		x.Namespace = v
	}
}

func SetRedis(v *redis.Client) Option {
	return func(x *Locker) {
		x.RDb = v
	}
}

func (x *Locker) Key(name string) string {
	return fmt.Sprintf(`%s:locker:%s`, x.Namespace, name)
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
