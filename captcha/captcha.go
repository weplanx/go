package captcha

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type Captcha struct {
	RDb *redis.Client
}

func New(rdb *redis.Client) *Captcha {
	return &Captcha{RDb: rdb}
}

func (x *Captcha) Key(name string) string {
	return fmt.Sprintf(`captcha:%s`, name)
}

func (x *Captcha) Create(ctx context.Context, name string, code string, ttl time.Duration) string {
	return x.RDb.Set(ctx, x.Key(name), code, ttl).Val()
}

func (x *Captcha) Exists(ctx context.Context, name string) bool {
	return x.RDb.Exists(ctx, x.Key(name)).Val() != 0
}

var (
	ErrCaptchaNotExists    = errors.NewPublic("the captcha does not exists")
	ErrCaptchaInconsistent = errors.NewPublic("tha captcha is invalid")
)

func (x *Captcha) Verify(ctx context.Context, name string, code string) error {
	if !x.Exists(ctx, name) {
		return ErrCaptchaNotExists
	}
	result := x.RDb.Get(ctx, x.Key(name)).Val()
	if result != code {
		return ErrCaptchaInconsistent
	}
	return nil
}

func (x *Captcha) Delete(ctx context.Context, name string) int64 {
	return x.RDb.Del(ctx, x.Key(name)).Val()
}
