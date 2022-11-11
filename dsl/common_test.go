package dsl_test

import (
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/nats-io/nats.go"
	"github.com/weplanx/utils/dsl"
	"github.com/weplanx/utils/helper"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"testing"
)

var (
	mgo *mongo.Client
	db  *mongo.Database
	nc  *nats.Conn
	js  nats.JetStreamContext
	r   *route.Engine
)

type M = map[string]interface{}

func TestMain(m *testing.M) {
	if err := UseMongoDB(); err != nil {
		panic(err)
	}
	if err := UseNats(); err != nil {
		panic(err)
	}
	service := &dsl.Service{
		DSL: dsl.New(
			dsl.SetNamespace("dev"),
			dsl.SetDatabase(db),
			dsl.SetEvent(js),
			dsl.SetValues(map[string]dsl.Value{}),
		),
	}
	helper.RegValidate()
	r = route.NewEngine(config.NewOptions([]config.Option{}))
	r.Use(ErrHandler())
	helper.BindDSL(r.Group("/:collection"), &dsl.Controller{DSLService: service})
	os.Exit(m.Run())
}
