package values_test

import (
	"bytes"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/hertz-contrib/requestid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/weplanx/go/cipher"
	"github.com/weplanx/go/help"
	"github.com/weplanx/go/values"
	"log"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"
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
var (
	keyvalue nats.KeyValue
	service  *values.Service
	engine   *route.Engine
)

func TestMain(m *testing.M) {
	var err error
	namespace := os.Getenv("NAMESPACE")
	if err = UseNats(namespace); err != nil {
		log.Fatalln(err)
	}
	var cipherx *cipher.Cipher
	if cipherx, err = cipher.New(os.Getenv("KEY")); err != nil {
		log.Fatalln(err)
	}
	service = values.New(
		values.SetKeyValue(keyvalue),
		values.SetCipher(cipherx),
		values.SetType(reflect.TypeOf(values.DynamicValues{})),
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
	controller := &values.Controller{Service: service}
	r := engine.Group("values")
	{
		r.GET("", controller.Get)
		r.PATCH("", controller.Set)
		r.DELETE(":key", controller.Remove)
	}
	os.Exit(m.Run())
}

func UseNats(namespace string) (err error) {
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
	var js nats.JetStreamContext
	if js, err = nc.JetStream(nats.PublishAsyncMaxPending(256)); err != nil {
		return
	}
	if keyvalue, err = js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: namespace,
	}); err != nil {
		return
	}
	return
}

func Reset() (err error) {
	data := make(map[string]interface{})
	v := reflect.ValueOf(DEFAULT)
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		data[typ.Field(i).Name] = v.Field(i).Interface()
	}
	return service.Update(data)
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
