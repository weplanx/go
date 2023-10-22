package rest

import (
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/weplanx/go/cipher"
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

func SetCipher(v *cipher.Cipher) Option {
	return func(x *Service) {
		x.Cipher = v
	}
}

type M = map[string]interface{}

var ErrCollectionForbidden = errors.NewPublic("the collection is forbidden")
var ErrTxnNotExist = errors.NewPublic("the txn does not exist")
var ErrTxnTimeOut = errors.NewPublic("the transaction has timed out")
