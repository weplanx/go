package sessions

import (
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"time"
)

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)

type Sessions struct {
	Namespace string
	Redis     *redis.Client
	Values    *Values
}

type Values struct {
	// 会话周期（秒）
	// 用户在 1 小时 内没有操作，将结束会话。
	SessionTTL time.Duration `json:"session_ttl"`
}

func New(options ...Option) *Sessions {
	x := new(Sessions)
	for _, v := range options {
		v(x)
	}
	return x
}

type Option func(x *Sessions)

func SetNamespace(v string) Option {
	return func(x *Sessions) {
		x.Namespace = v
	}
}

func SetRedis(v *redis.Client) Option {
	return func(x *Sessions) {
		x.Redis = v
	}
}

func SetValues(v *Values) Option {
	return func(x *Sessions) {
		x.Values = v
	}
}
