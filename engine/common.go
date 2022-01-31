package engine

import (
	"fmt"
	"github.com/google/wire"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
)

const ModelNameKey = "model-name"

type Engine struct {
	App     string
	Options map[string]Option
	Js      nats.JetStreamContext
}

type Option struct {
	Projection bson.M `yaml:"projection"`
}

type OptionFunc func(engine *Engine)

func SetApp(v string) OptionFunc {
	return func(engine *Engine) {
		engine.App = v
	}
}

func UseStaticOptions(v map[string]Option) OptionFunc {
	return func(engine *Engine) {
		engine.Options = v
	}
}

func UseEvents(v nats.JetStreamContext) OptionFunc {
	return func(engine *Engine) {
		engine.Js = v
	}
}

func New(options ...OptionFunc) *Engine {
	x := &Engine{App: ""}
	for _, v := range options {
		v(x)
	}
	return x
}

type EventValue struct {
	Id       interface{} `json:"id"`
	Query    interface{} `json:"query"`
	Body     interface{} `json:"body"`
	Response interface{} `json:"response"`
}

func (x *Engine) Publish(model string, event string, v EventValue) (err error) {
	if x.Js == nil {
		return
	}
	var data []byte
	if data, err = jsoniter.Marshal(&v); err != nil {
		return
	}
	if _, err = x.Js.Publish(
		fmt.Sprintf(`%s.%s.%s`, x.App, model, event),
		data,
	); err != nil {
		return
	}
	return
}

type Pagination struct {
	Index int64 `header:"x-page" binding:"omitempty,gt=0,number"`
	Size  int64 `header:"x-page-size" binding:"omitempty,number"`
}

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)
