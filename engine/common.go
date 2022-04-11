package engine

import (
	"errors"
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

var (
	BodyEmpty = errors.New("the request body data cannot be empty")
)

type M = map[string]interface{}

type Params struct {
	*Uri
	*Headers
}

type Uri struct {
	// 模型命名
	Model string `uri:"model" binding:"omitempty,key"`
	// 文档 ID
	Id string `uri:"id" binding:"omitempty,objectId"`
}

type Headers struct {
	// 方法
	Action string `header:"wpx-action" binding:"omitempty,oneof=create bulk-create bulk-delete"`
	// 查询类型
	Type string `header:"wpx-type" binding:"omitempty"`
	// 最大返回数量
	Limit int64 `header:"wpx-limit" binding:"omitempty,gt=0,lt=10000"`
	// 跳过数量
	Skip int64 `header:"wpx-skip" binding:"omitempty,gte=0"`
	// 分页码
	Index int64 `header:"wpx-page" binding:"omitempty,gt=0,number"`
	// 分页大小
	Size int64 `header:"wpx-page-size" binding:"omitempty,number"`
	// 总数
	Total int64 `header:"wpx-total"`
	// 格式化过滤
	FormatFilter string `header:"wpx-format-filter" binding:"omitempty,gt=0"`
	// 格式化文档
	FormatDoc string `header:"wpx-format-doc" binding:"omitempty,gt=0"`
}
