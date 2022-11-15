package dsl_test

import (
	"context"
	"fmt"
	"github.com/bytedance/go-tagexpr/v2/binding"
	"github.com/bytedance/go-tagexpr/v2/validator"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/bytedance/sonic/decoder"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/weplanx/utils/dsl"
	"github.com/weplanx/utils/helper"
	"github.com/weplanx/utils/kv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"log"
	"net/http"
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
		log.Fatalln(err)
	}
	if err := UseNats(); err != nil {
		log.Fatalln(err)
	}
	dv := &kv.DynamicValues{
		DSL: map[string]kv.DSLOption{
			"users": {
				Keys: map[string]int64{
					"name":        1,
					"department":  1,
					"roles":       1,
					"create_time": 1,
					"update_time": 1,
				},
			},
			"projects": {
				Event: true,
			},
		},
	}
	for k, v := range dv.DSL {
		if v.Event {
			js.DeleteStream(fmt.Sprintf(`%s:events:%s`, "dev", k))
		}
	}
	x, err := dsl.New(
		dsl.SetNamespace("dev"),
		dsl.SetDatabase(db),
		dsl.SetDynamicValues(dv),
		dsl.SetJetStream(js),
	)
	if err != nil {
		log.Fatalln(err)
	}

	service := &dsl.Service{DSL: x}
	helper.RegValidate()
	r = route.NewEngine(config.NewOptions([]config.Option{}))
	r.Use(ErrHandler())
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
	usersOption := options.CreateCollection().
		SetValidator(bson.D{
			{"$jsonSchema", bson.D{
				{"title", "users"},
				{"required", bson.A{"_id", "name", "password", "department", "roles", "create_time", "update_time"}},
				{"properties", bson.D{
					{"_id", bson.M{"bsonType": "objectId"}},
					{"name", bson.M{"bsonType": "string"}},
					{"password", bson.M{"bsonType": "string"}},
					{"department", bson.M{"bsonType": []string{"null", "objectId"}}},
					{"roles", bson.M{
						"bsonType": "array",
						"items":    bson.M{"bsonType": "objectId"},
					}},
					{"create_time", bson.M{"bsonType": "date"}},
					{"update_time", bson.M{"bsonType": "date"}},
				}},
				{"additionalProperties", false},
			}},
		})
	if err = db.CreateCollection(context.TODO(), "users", usersOption); err != nil {
		return
	}
	ordersOption := options.CreateCollection().
		SetValidator(bson.D{
			{"$jsonSchema", bson.D{
				{"title", "orders"},
				{"required", bson.A{"_id", "no", "customer", "phone", "cost", "time", "create_time", "update_time"}},
				{"properties", bson.D{
					{"_id", bson.M{"bsonType": "objectId"}},
					{"no", bson.M{"bsonType": "string"}},
					{"customer", bson.M{"bsonType": "string"}},
					{"phone", bson.M{"bsonType": "string"}},
					{"cost", bson.M{"bsonType": "number"}},
					{"time", bson.M{"bsonType": "date"}},
					{"sort", bson.M{"bsonType": []string{"null", "number"}}},
					{"create_time", bson.M{"bsonType": "date"}},
					{"update_time", bson.M{"bsonType": "date"}},
				}},
				{"additionalProperties", false},
			}},
		})
	if err = db.CreateCollection(context.TODO(), "orders", ordersOption); err != nil {
		return
	}
	projectsOption := options.CreateCollection().SetValidator(bson.D{
		{"$jsonSchema", bson.D{
			{"title", "projects"},
			{"required", bson.A{"_id", "name", "namespace", "secret", "create_time", "update_time"}},
			{"properties", bson.D{
				{"_id", bson.M{"bsonType": "objectId"}},
				{"name", bson.M{"bsonType": "string"}},
				{"namespace", bson.M{"bsonType": "string"}},
				{"secret", bson.M{"bsonType": "string"}},
				{"expire_time", bson.M{"bsonType": []string{"null", "date"}}},
				{"sort", bson.M{"bsonType": []string{"null", "number"}}},
				{"create_time", bson.M{"bsonType": "date"}},
				{"update_time", bson.M{"bsonType": "date"}},
			}},
			{"additionalProperties", false},
		}},
	})
	if err = db.CreateCollection(context.TODO(), "projects", projectsOption); err != nil {
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
		return
	}
	if js, err = nc.JetStream(nats.PublishAsyncMaxPending(256)); err != nil {
		return
	}
	return
}

func ErrHandler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Next(ctx)
		err := c.Errors.Last()
		if err == nil {
			return
		}

		if err.IsType(errors.ErrorTypePublic) {
			statusCode := http.StatusBadRequest
			result := utils.H{"message": err.Error()}
			if meta, ok := err.Meta.(map[string]interface{}); ok {
				if meta["statusCode"] != nil {
					statusCode = meta["statusCode"].(int)
				}
				if meta["code"] != nil {
					result["code"] = meta["code"]
				}
			}
			c.JSON(statusCode, result)
			return
		}

		switch e := err.Err.(type) {
		case decoder.SyntaxError:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Description(),
			})
			break
		case *binding.Error:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Error(),
			})
			break
		case *validator.Error:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Error(),
			})
			break
		default:
			logger.Error(err)
			c.Status(http.StatusInternalServerError)
		}
	}
}
