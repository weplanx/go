package rest_test

import (
	"bytes"
	"context"
	"github.com/bytedance/go-tagexpr/v2/binding"
	"github.com/bytedance/go-tagexpr/v2/validator"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/help"
	"github.com/weplanx/go/rest"
	"github.com/weplanx/go/values"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"net/http"
	"net/url"
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

var controls = map[string]*values.RestControl{
	"users": {
		Keys:   []string{"name", "department", "roles", "create_time", "update_time"},
		Status: true,
	},
	"roles": {
		Status: true,
	},
	"projects": {
		Status: true,
		Event:  true,
	},
	"orders": {
		Status: true,
	},
	"coupons": {
		Status: true,
	},
	"x_test": {
		Status: true,
	},
	"x_roles": {
		Status: true,
	},
	"x_users": {
		Status: true,
	},
	"levels": {
		Status: false,
	},
}

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
		rest.SetDynamicValues(&values.DynamicValues{
			RestControls:   controls,
			RestTxnTimeout: time.Second * 30,
		}),
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

func R(method string, url string, body interface{}) (resp *protocol.Response, err error) {
	utBody := &ut.Body{}
	utHeaders := []ut.Header{
		{Key: "content-type", Value: "application/json"},
	}
	if body != nil {
		var b []byte
		if b, err = sonic.Marshal(body); err != nil {
			return
		}
		utBody.Body = bytes.NewBuffer(b)
		utBody.Len = len(b)
	}

	w := ut.PerformRequest(engine, method, url,
		utBody,
		utHeaders...,
	)

	resp = w.Result()
	return
}

type Params = [][2]string

func U(path string, params Params) string {
	u := url.URL{Path: path}
	query := u.Query()
	for _, v := range params {
		query.Add(v[0], v[1])
	}
	u.RawQuery = query.Encode()
	return u.RequestURI()
}

type TransactionFn = func(txn string)

func Transaction(t *testing.T, fn TransactionFn) {
	resp1, err := R("POST", "/transaction", nil)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp1.StatusCode())

	var result1 M
	err = sonic.Unmarshal(resp1.Body(), &result1)
	assert.NoError(t, err)
	txn := result1["txn"].(string)

	fn(txn)

	resp2, err := R("POST", "/commit", M{
		"txn": txn,
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode())
}
