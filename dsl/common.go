package dsl

import (
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/mongo"
)

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)

type M = map[string]interface{}

type DSL struct {
	Namespace string
	Db        *mongo.Database
	Js        nats.JetStreamContext
	Values    map[string]Value
}

type Value struct {
	Event   bool
	Project M
}

func New(options ...Option) *DSL {
	x := new(DSL)
	for _, v := range options {
		v(x)
	}
	return x
}

type Option func(x *DSL)

func SetNamespace(v string) Option {
	return func(x *DSL) {
		x.Namespace = v
	}
}

func SetDatabase(v *mongo.Database) Option {
	return func(x *DSL) {
		x.Db = v
	}
}

func SetEvent(v nats.JetStreamContext) Option {
	return func(x *DSL) {
		x.Js = v
	}
}

func SetValues(v map[string]Value) Option {
	return func(x *DSL) {
		x.Values = v
	}
}
