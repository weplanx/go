package sessions_test

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
	"github.com/redis/go-redis/v9"
	"github.com/weplanx/utils/helper"
	"github.com/weplanx/utils/sessions"
	"github.com/weplanx/utils/values"
	"log"
	"net/http"
	"os"
	"testing"
)

var (
	service *sessions.Service
	rdb     *redis.Client
	r       *route.Engine
)

type M = map[string]interface{}

func TestMain(m *testing.M) {
	if err := UseRedis(); err != nil {
		log.Fatalln(err)
	}
	service = sessions.New(
		sessions.SetNamespace("dev"),
		sessions.SetRedis(rdb),
		sessions.SetDynamicValues(&values.DEFAULT),
	)
	r = route.NewEngine(config.NewOptions([]config.Option{}))
	r.Use(ErrHandler())
	helper.BindSessions(r.Group(""), &sessions.Controller{Service: service})
	os.Exit(m.Run())
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
