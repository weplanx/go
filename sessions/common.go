package sessions

import (
	"github.com/redis/go-redis/v9"
	"github.com/weplanx/go/values"
)

func New(options ...Option) *Service {
	x := new(Service)
	for _, v := range options {
		v(x)
	}
	return x
}

type Option func(x *Service)

func SetRedis(v *redis.Client) Option {
	return func(x *Service) {
		x.RDb = v
	}
}

func SetDynamicValues(v *values.DynamicValues) Option {
	return func(x *Service) {
		x.Values = v
	}
}
