package test

import (
	"context"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/engine"
	"github.com/weplanx/go/helper"
	"github.com/weplanx/go/route"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"testing"
	"time"
)

var (
	r      *gin.Engine
	client *mongo.Client
	db     *mongo.Database
	nc     *nats.Conn
	js     nats.JetStreamContext
)

type M = map[string]interface{}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	r = gin.Default()
	if err := SetMongoDB(); err != nil {
		panic(err)
	}
	if err := SetNats(); err != nil {
		panic(err)
	}
	e := engine.New(
		engine.SetApp("example"),
		engine.UseStaticOptions(map[string]engine.Option{
			"users": {
				Event: true,
				Field: []string{"name", "batch"},
			},
		}),
		engine.UseEvents(js),
	)
	service := engine.Service{
		Engine: e,
		Db:     db,
	}
	controller := engine.Controller{
		Engine:  e,
		Service: &service,
	}
	helper.ExtendValidation()
	api := r.Group("")
	{
		route.Engine(api, &controller)
		api.GET("svc/:id", route.Use(controller.GetById, route.SetModel("services")))
	}

	os.Exit(m.Run())
}

func SetMongoDB() (err error) {
	if client, err = mongo.Connect(context.TODO(),
		options.Client().ApplyURI(os.Getenv("TEST_DB")),
	); err != nil {
		return
	}
	db = client.Database("example")
	if err = db.Drop(context.TODO()); err != nil {
		return
	}
	return
}

func SetNats() (err error) {
	var kp nkeys.KeyPair
	if kp, err = nkeys.FromSeed([]byte(os.Getenv("TEST_NATS_NKEY"))); err != nil {
		return
	}
	defer kp.Wipe()
	var pub string
	if pub, err = kp.PublicKey(); err != nil {
		return
	}
	if !nkeys.IsValidPublicUserKey(pub) {
		panic("nkey 验证失败")
	}
	if nc, err = nats.Connect(
		os.Getenv("TEST_NATS"),
		nats.MaxReconnects(5),
		nats.ReconnectWait(2*time.Second),
		nats.ReconnectJitter(500*time.Millisecond, 2*time.Second),
		nats.Nkey(pub, func(nonce []byte) ([]byte, error) {
			sig, _ := kp.Sign(nonce)
			return sig, nil
		}),
	); err != nil {
		panic(err)
	}
	if js, err = nc.JetStream(nats.PublishAsyncMaxPending(256)); err != nil {
		panic(err)
	}
	js.DeleteStream("example:events:users")
	return
}

func Comparison(t *testing.T, exptcted []M, actual []M) {
	assert.Equal(t, len(exptcted), len(actual))
	hmap := make(map[string]M)
	for _, v := range exptcted {
		hmap[v["number"].(string)] = v
	}
	for _, v := range actual {
		assert.NotNil(t, hmap[v["number"].(string)])
		doc := hmap[v["number"].(string)]
		assert.Equal(t, doc["name"], v["name"])
		assert.Equal(t, doc["price"], v["price"])
	}
}

func Event(t *testing.T, expected string) {
	sub, err := js.PullSubscribe("example.events.users", "example:events:users")
	if err != nil {
		t.Error(err)
	}
	msg, err := sub.Fetch(1)
	if err != nil {
		t.Error(err)
	}
	assert.Len(t, msg, 1)
	var e M
	if err = jsoniter.Unmarshal(msg[0].Data, &e); err != nil {
		t.Error(err)
	}
	assert.Equal(t, expected, e["action"])
}
