package engine

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/wire"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"regexp"
)

const ModelNameKey = "model-name"

type Engine struct {
	App string
	Js  nats.JetStreamContext
}

type OptionFunc func(engine *Engine)

func SetApp(v string) OptionFunc {
	return func(engine *Engine) {
		engine.App = v
	}
}

func UseEvents(js nats.JetStreamContext) OptionFunc {
	return func(engine *Engine) {
		engine.Js = js
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
	Size  int64 `header:"x-page-size" binding:"omitempty,oneof=10 20 50 100"`
}

func RegisterValidation() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("objectId", func(fl validator.FieldLevel) bool {
			return primitive.IsValidObjectID(fl.Field().String())
		})
		v.RegisterValidation("key", func(fl validator.FieldLevel) bool {
			matched, _ := regexp.MatchString(`^[a-z_]+$`, fl.Field().String())
			return matched
		})
		v.RegisterValidation("sort", func(fl validator.FieldLevel) bool {
			matched, _ := regexp.MatchString(`^[a-z_]+\.(1|-1)$`, fl.Field().String())
			return matched
		})
	}
}

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)
