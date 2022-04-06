package engine

import (
	"fmt"
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
	"log"
)

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)

type Engine struct {
	App     string
	Js      nats.JetStreamContext
	Options map[string]Option
}

func New(options ...OptionFunc) *Engine {
	x := &Engine{App: ""}
	for _, v := range options {
		v(x)
	}
	return x
}

type OptionFunc func(engine *Engine)

type Option struct {
	Event bool     `yaml:"event"`
	Field []string `yaml:"field"`
}

func SetApp(name string) OptionFunc {
	return func(engine *Engine) {
		engine.App = name
	}
}

func UseStaticOptions(options map[string]Option) OptionFunc {
	return func(engine *Engine) {
		engine.Options = options
	}
}

func UseEvents(js nats.JetStreamContext) OptionFunc {
	return func(engine *Engine) {
		for k, v := range engine.Options {
			if v.Event {
				name := fmt.Sprintf(`%s:events:%s`, engine.App, k)
				subject := fmt.Sprintf(`%s.events.%s`, engine.App, k)
				if _, err := js.AddStream(&nats.StreamConfig{
					Name:      name,
					Subjects:  []string{subject},
					Retention: nats.WorkQueuePolicy,
				}); err != nil {
					log.Fatalln(err)
				}
			}
		}
		engine.Js = js
	}
}

type M = map[string]interface{}

type Params struct {
	Model string `uri:"model" binding:"omitempty,key"`
	Id    string `uri:"id" binding:"omitempty,objectId"`
}

type Pagination struct {
	Index int64 `header:"x-page" binding:"omitempty,gt=0,number"`
	Size  int64 `header:"x-page-size" binding:"omitempty,number"`
	Total int64
}
