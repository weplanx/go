package resources

import (
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/weplanx/utils/values"
	"go.mongodb.org/mongo-driver/mongo"
)

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
)

type M = map[string]interface{}

func New(options ...Option) (x *Service, err error) {
	x = new(Service)
	for _, v := range options {
		v(x)
	}
	return
}

type Option func(x *Service)

func SetNamespace(v string) Option {
	return func(x *Service) {
		x.Namespace = v
	}
}

func SetMongoClient(v *mongo.Client) Option {
	return func(x *Service) {
		x.MongoClient = v
	}
}

func SetDatabase(v *mongo.Database) Option {
	return func(x *Service) {
		x.Db = v
	}
}

func SetRedis(v *redis.Client) Option {
	return func(x *Service) {
		x.Redis = v
	}
}

func SetJetStream(v nats.JetStreamContext) Option {
	return func(x *Service) {
		x.Js = v
	}
}

func SetDynamicValues(v *values.DynamicValues) Option {
	return func(x *Service) {
		x.DynamicValues = v
	}
}
