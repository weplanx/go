package values_test

import (
	"context"
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
	"github.com/weplanx/utils/helper"
	"github.com/weplanx/utils/values"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	nc       *nats.Conn
	js       nats.JetStreamContext
	keyvalue nats.KeyValue
	service  *values.Service
	r        *route.Engine
)

type M = map[string]interface{}

func TestMain(m *testing.M) {
	if err := UseNats("dev"); err != nil {
		log.Fatalln(err)
	}
	service = values.New(
		values.SetNamespace("dev"),
		values.SetKeyValue(keyvalue),
		values.SetDynamicValues(&values.DEFAULT),
	)
	r = route.NewEngine(config.NewOptions([]config.Option{}))
	r.Use(ErrHandler())
	helper.BindKV(r.Group(""), &values.Controller{Service: service})
	os.Exit(m.Run())
}

func UseNats(namespace string) (err error) {
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
	js.DeleteKeyValue(namespace)
	if keyvalue, err = js.CreateKeyValue(&nats.KeyValueConfig{Bucket: namespace}); err != nil {
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
