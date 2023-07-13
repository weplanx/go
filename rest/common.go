package rest

import (
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/weplanx/go/values"
	"go.mongodb.org/mongo-driver/mongo"
)

func New(options ...Option) *Service {
	x := new(Service)
	for _, v := range options {
		v(x)
	}
	return x
}

type Option func(x *Service)

func SetNamespace(v string) Option {
	return func(x *Service) {
		x.Namespace = v
	}
}

func SetMongoClient(v *mongo.Client) Option {
	return func(x *Service) {
		x.Mgo = v
	}
}

func SetDatabase(v *mongo.Database) Option {
	return func(x *Service) {
		x.Db = v
	}
}

func SetRedis(v *redis.Client) Option {
	return func(x *Service) {
		x.RDb = v
	}
}

func SetJetStream(v nats.JetStreamContext) Option {
	return func(x *Service) {
		x.JetStream = v
	}
}

func SetKeyValue(v nats.KeyValue) Option {
	return func(x *Service) {
		x.KeyValue = v
	}
}

func SetDynamicValues(v *values.DynamicValues) Option {
	return func(x *Service) {
		x.Values = v
	}
}
