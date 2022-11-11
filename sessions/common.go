package sessions

import (
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"github.com/weplanx/utils/kv"
)

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)

type Sessions struct {
	Namespace     string
	Redis         *redis.Client
	DynamicValues *kv.DynamicValues
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

func SetDynamicValues(v *kv.DynamicValues) Option {
	return func(x *Sessions) {
		x.DynamicValues = v
	}
}
