package test

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/weplanx/go/engine"
	"github.com/weplanx/go/helper"
	"github.com/weplanx/go/route"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"testing"
	"time"
)

var r *gin.Engine
var db *mongo.Database
var nc *nats.Conn
var js nats.JetStreamContext

type M = map[string]interface{}

func SetMongoDB() (err error) {
	var client *mongo.Client
	if client, err = mongo.Connect(context.TODO(),
		options.Client().ApplyURI(os.Getenv("TEST_DB")),
	); err != nil {
		return
	}
	db = client.Database("example")
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
	return
}

func TestMain(m *testing.M) {
	r = gin.Default()
	gin.SetMode(gin.TestMode)
	if err := SetMongoDB(); err != nil {
		panic(err)
	}
	if err := SetNats(); err != nil {
		panic(err)
	}
	e := engine.New(
		engine.SetApp("test"),
		engine.UseStaticOptions(map[string]engine.Option{
			"pages": {
				Event: true,
			},
			"users": {
				Field: []string{"name", "alias"},
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
		// alias
		api.GET("svc/:id", route.Use(controller.GetById, route.SetModel("services")))
	}
	if err := db.Drop(context.TODO()); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

var services = []map[string]interface{}{
	{"number": "55826199", "name": "Handmade Soft Salad", "price": 727.00},
	{"number": "57277117", "name": "Intelligent Fresh Shoes", "price": 47.00},
	{"number": "52697132", "name": "Practical Metal Chips", "price": 859.00},
	{"number": "66502334", "name": "Ergonomic Wooden Pizza", "price": 328.00},
	{"number": "43678700", "name": "Intelligent Cotton Chips", "price": 489.00},
	{"number": "66204618", "name": "Sleek Rubber Cheese", "price": 986.00},
	{"number": "82877045", "name": "Unbranded Fresh Ball", "price": 915.00},
	{"number": "11254621", "name": "Handmade Metal Keyboard", "price": 244.00},
	{"number": "24443471", "name": "Rustic Frozen Gloves", "price": 500.00},
	{"number": "74371061", "name": "Awesome Frozen Gloves", "price": 214.00},
	{"number": "39012286", "name": "Sleek Steel Bike", "price": 428.00},
	{"number": "58946201", "name": "Handmade Plastic Pizza", "price": 913.00},
	{"number": "08945471", "name": "Generic Metal Pizza", "price": 810.00},
	{"number": "40208673", "name": "Handcrafted Granite Shoes", "price": 429.00},
	{"number": "84106393", "name": "Refined Steel Bike", "price": 339.00},
	{"number": "52669450", "name": "Handmade Frozen Keyboard", "price": 684.00},
	{"number": "15525688", "name": "Tasty Cotton Pants", "price": 995.00},
	{"number": "38438365", "name": "Awesome Soft Soap", "price": 142.00},
	{"number": "48780690", "name": "Intelligent Cotton Gloves", "price": 297.00},
	{"number": "62787493", "name": "Rustic Frozen Salad", "price": 542.00},
	{"number": "35433318", "name": "Small Soft Keyboard", "price": 703.00},
	{"number": "87239908", "name": "Handmade Granite Sausages", "price": 88.00},
	{"number": "63793023", "name": "Intelligent Soft Bike", "price": 630.00},
	{"number": "60599531", "name": "Unbranded Wooden Bacon", "price": 98.00},
	{"number": "10550233", "name": "Intelligent Steel Tuna", "price": 499.00},
	{"number": "89885575", "name": "Unbranded Frozen Chicken", "price": 667.00},
	{"number": "90424834", "name": "Handcrafted Wooden Shoes", "price": 516.00},
	{"number": "77762017", "name": "Generic Rubber Table", "price": 725.00},
	{"number": "07605361", "name": "Incredible Metal Towels", "price": 261.00},
	{"number": "92417878", "name": "Small Fresh Table", "price": 662.00},
	{"number": "12181549", "name": "Refined Soft Ball", "price": 385.00},
	{"number": "23740764", "name": "Unbranded Soft Mouse", "price": 710.00},
	{"number": "75813798", "name": "Tasty Metal Chips", "price": 506.00},
	{"number": "70353191", "name": "Tasty Cotton Hat", "price": 480.00},
	{"number": "67153899", "name": "Generic Frozen Bike", "price": 261.00},
	{"number": "14395918", "name": "Awesome Steel Towels", "price": 796.00},
	{"number": "24957863", "name": "Ergonomic Soft Chair", "price": 599.00},
	{"number": "84480037", "name": "Fantastic Metal Salad", "price": 273.00},
	{"number": "10531004", "name": "Tasty Rubber Bike", "price": 696.00},
	{"number": "37050804", "name": "Intelligent Soft Pants", "price": 451.00},
	{"number": "15757338", "name": "Fantastic Fresh Soap", "price": 281.00},
	{"number": "83666844", "name": "Rustic Wooden Shoes", "price": 477.00},
	{"number": "60049311", "name": "Refined Steel Pizza", "price": 719.00},
	{"number": "25627132", "name": "Licensed Wooden Bacon", "price": 585.00},
	{"number": "44243580", "name": "Handmade Granite Fish", "price": 3.00},
	{"number": "67644215", "name": "Refined Plastic Keyboard", "price": 796.00},
	{"number": "99821780", "name": "Refined Frozen Pants", "price": 569.00},
	{"number": "09613501", "name": "Handcrafted Soft Sausages", "price": 826.00},
	{"number": "35568587", "name": "Practical Soft Sausages", "price": 500.00},
	{"number": "92044481", "name": "Sleek Soft Soap", "price": 309.00},
}