package sessions_test

import (
	"bytes"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/hertz-contrib/requestid"
	"github.com/redis/go-redis/v9"
	"github.com/weplanx/go/help"
	"github.com/weplanx/go/sessions"
	"github.com/weplanx/go/values"
	"log"
	"net/url"
	"os"
	"testing"
	"time"
)

var (
	service *sessions.Service
	rdb     *redis.Client
	engine  *route.Engine
)

type M = map[string]interface{}

var DEFAULT = values.DynamicValues{
	SessionTTL:      time.Hour,
	LoginTTL:        time.Minute * 15,
	LoginFailures:   5,
	IpLoginFailures: 10,
	IpWhitelist:     []string{},
	IpBlacklist:     []string{},
	PwdStrategy:     1,
	PwdTTL:          time.Hour * 24 * 365,
}
var v = DEFAULT

func TestMain(m *testing.M) {
	if err := UseRedis(); err != nil {
		log.Fatalln(err)
	}
	service = sessions.New(
		sessions.SetRedis(rdb),
		sessions.SetDynamicValues(&v),
	)
	engine = route.NewEngine(config.NewOptions([]config.Option{
		server.WithExitWaitTime(0),
		server.WithDisablePrintRoute(true),
		server.WithCustomValidator(help.Validator()),
	}))
	engine.Use(
		requestid.New(),
		help.EHandler(),
	)
	controller := &sessions.Controller{Service: service}
	r := engine.Group("sessions")
	{
		r.GET("", controller.Lists)
		r.DELETE(":uid", controller.Remove)
		r.POST("clear", controller.Clear)
	}
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
