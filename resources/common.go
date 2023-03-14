package resources

import (
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/weplanx/utils/kv"
	"go.mongodb.org/mongo-driver/mongo"
)

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)

type M = map[string]interface{}

type DSL struct {
	Namespace     string
	MongoClient   *mongo.Client
	Db            *mongo.Database
	Redis         *redis.Client
	DynamicValues *kv.DynamicValues
	Js            nats.JetStreamContext
}

func New(options ...Option) (x *DSL, err error) {
	x = new(DSL)
	for _, v := range options {
		v(x)
	}
	return
}

type Option func(x *DSL)

func SetNamespace(v string) Option {
	return func(x *DSL) {
		x.Namespace = v
	}
}

func SetMongoClient(v *mongo.Client) Option {
	return func(x *DSL) {
		x.MongoClient = v
	}
}

func SetDatabase(v *mongo.Database) Option {
	return func(x *DSL) {
		x.Db = v
	}
}

func SetRedis(v *redis.Client) Option {
	return func(x *DSL) {
		x.Redis = v
	}
}

func SetJetStream(v nats.JetStreamContext) Option {
	return func(x *DSL) {
		x.Js = v
	}
}

func SetDynamicValues(v *kv.DynamicValues) Option {
	return func(x *DSL) {
		x.DynamicValues = v
	}
}