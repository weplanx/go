package rest_test

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
	"github.com/redis/go-redis/v9"
	"github.com/weplanx/go/help"
	"github.com/weplanx/go/rest"
	"github.com/weplanx/go/values"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	mgo      *mongo.Client
	db       *mongo.Database
	rdb      *redis.Client
	js       nats.JetStreamContext
	keyvalue nats.KeyValue
	service  *rest.Service
	engine   *route.Engine
)

type M = map[string]interface{}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := UseMongo(ctx); err != nil {
		panic(err)
	}
	if err := UseRedis(); err != nil {
		panic(err)
	}
	if err := UseNats(ctx); err != nil {
		panic(err)
	}
	namespace := os.Getenv("NAMESPACE")
	service = rest.New(
		rest.SetNamespace(namespace),
		rest.SetMongoClient(mgo),
		rest.SetDatabase(db),
		rest.SetRedis(rdb),
		rest.SetJetStream(js),
		rest.SetKeyValue(keyvalue),
		rest.SetDynamicValues(&values.DynamicValues{RestControls: map[string]*values.RestControl{
			"users": {
				Keys: []string{"name", "department", "roles", "create_time", "update_time"},
			},
			"projects": {
				Event: true,
			},
		}}),
	)
	if err := MockDb(ctx); err != nil {
		panic(err)
	}
	if err := MockStream(ctx); err != nil {
		panic(err)
	}
	help.RegValidate()
	engine = route.NewEngine(config.NewOptions([]config.Option{}))
	engine.Use(ErrHandler())
	help.RestRoutes(engine.Group(""), &rest.Controller{Service: service})

	os.Exit(m.Run())
}

func UseMongo(ctx context.Context) (err error) {
	if mgo, err = mongo.Connect(ctx,
		options.Client().ApplyURI(os.Getenv("DATABASE_URL")),
	); err != nil {
		return
	}
	option := options.Database().
		SetWriteConcern(writeconcern.Majority())
	db = mgo.Database(os.Getenv("DATABASE_NAME"), option)
	if err = db.Drop(ctx); err != nil {
		return
	}
	return
}

func MockDb(ctx context.Context) (err error) {
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
	if err = db.CreateCollection(ctx, "users", usersOption); err != nil {
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
	if err = db.CreateCollection(ctx, "orders", ordersOption); err != nil {
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
	if err = db.CreateCollection(ctx, "projects", projectsOption); err != nil {
		return
	}
	return
}

func UseRedis() (err error) {
	var opts *redis.Options
	opts, err = redis.ParseURL(os.Getenv("DATABASE_REDIS"))
	if err != nil {
		return
	}
	rdb = redis.NewClient(opts)
	return
}

func UseNats(ctx context.Context) (err error) {
	var auth nats.Option
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
		panic("nkey failed")
	}
	auth = nats.Nkey(pub, func(nonce []byte) ([]byte, error) {
		sig, _ := kp.Sign(nonce)
		return sig, nil
	})
	var nc *nats.Conn
	if nc, err = nats.Connect(
		os.Getenv("NATS_HOSTS"),
		nats.MaxReconnects(5),
		nats.ReconnectWait(2*time.Second),
		nats.ReconnectJitter(500*time.Millisecond, 2*time.Second),
		auth,
	); err != nil {
		return
	}
	if js, err = nc.JetStream(nats.PublishAsyncMaxPending(256), nats.Context(ctx)); err != nil {
		return
	}
	return
}

func MockStream(ctx context.Context) (err error) {
	for k, v := range service.Values.RestControls {
		if v.Event {
			name := fmt.Sprintf(`%s:events:%s`, service.Namespace, k)
			subject := fmt.Sprintf(`%s.events.%s`, service.Namespace, k)
			js.DeleteStream(name)
			if _, err := js.AddStream(&nats.StreamConfig{
				Name:      name,
				Subjects:  []string{subject},
				Retention: nats.WorkQueuePolicy,
			}, nats.Context(ctx)); err != nil {
				panic(err)
			}
		}
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
