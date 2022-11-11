package kv_test

import (
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/nats-io/nats.go"
	"github.com/weplanx/utils/helper"
	"github.com/weplanx/utils/kv"
	"log"
	"os"
	"testing"
)

var (
	nc       *nats.Conn
	js       nats.JetStreamContext
	keyvalue nats.KeyValue
	service  *kv.Service
	r        *route.Engine
)

type M = map[string]interface{}

func TestMain(m *testing.M) {
	if err := UseNats("dev"); err != nil {
		log.Fatalln(err)
	}
	dv := kv.DEFAULT
	service = &kv.Service{
		KV: kv.New(
			kv.SetNamespace("dev"),
			kv.SetKeyValue(keyvalue),
			kv.SetDynamicValues(&dv),
		),
	}
	r = route.NewEngine(config.NewOptions([]config.Option{}))
	r.Use(ErrHandler())
	helper.BindKV(r.Group("values"), &kv.Controller{KVService: service})
	os.Exit(m.Run())
}
