package dsl_test

import (
	"context"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/weplanx/utils/dsl"
	"github.com/weplanx/utils/helper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"os"
	"testing"
	"time"
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
		DSL: dsl.New(db, dsl.SetNamespace("dev"), dsl.SetEvent(js)),
	}
	r = route.NewEngine(config.NewOptions([]config.Option{}))
	helper.RegValidate()
	helper.BindDSL(r.Group("/:collection"), &dsl.Controller{DSLService: service})
	os.Exit(m.Run())
}

func UseMongoDB() (err error) {
	if mgo, err = mongo.Connect(context.TODO(),
		options.Client().ApplyURI(os.Getenv("DATABASE_MONGO")),
	); err != nil {
		return
	}
	db = mgo.Database("development",
		options.Database().SetWriteConcern(writeconcern.New(writeconcern.WMajority())),
	)
	if err = db.Drop(context.TODO()); err != nil {
		return
	}
	return
}

func UseNats() (err error) {
	var auth nats.Option
	if os.Getenv("NATS_TOKEN") != "" {
		auth = nats.Token(os.Getenv("NATS_TOKEN"))
	}
	if os.Getenv("NATS_NKEY") != "" {
		var kp nkeys.KeyPair
		if kp, err = nkeys.FromSeed([]byte(os.Getenv("NATS_NKEY"))); err != nil {
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
		auth = nats.Nkey(pub, func(nonce []byte) ([]byte, error) {
			sig, _ := kp.Sign(nonce)
			return sig, nil
		})
	}
	if nc, err = nats.Connect(
		os.Getenv("NATS_HOSTS"),
		nats.MaxReconnects(5),
		nats.ReconnectWait(2*time.Second),
		nats.ReconnectJitter(500*time.Millisecond, 2*time.Second),
		auth,
	); err != nil {
		panic(err)
	}
	if js, err = nc.JetStream(nats.PublishAsyncMaxPending(256)); err != nil {
		panic(err)
	}
	return
}
