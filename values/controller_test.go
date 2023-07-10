package values_test

import (
	"bytes"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

func TestSetBadValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"key1": "value1",
		},
	})
	w := ut.PerformRequest(engine, "PATCH", "/values",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestSet(t *testing.T) {
	err := service.Reset()
	assert.NoError(t, err)
	body, _ := sonic.Marshal(M{
		"update": M{
			"Wechat": "abcdefg",
		},
	})
	w := ut.PerformRequest(engine, "PATCH", "/values",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 204, resp.StatusCode())

	var data M
	data, err = service.Get("Wechat")
	assert.NoError(t, err)
	assert.Equal(t, "abcdefg", data["Wechat"])
}

func TestSetBadService(t *testing.T) {
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	body, _ := sonic.Marshal(M{
		"update": M{
			"Wechat": "abcdefg",
		},
	})
	w := ut.PerformRequest(engine, "PATCH", "/values",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

func TestGet(t *testing.T) {
	err := service.Reset()
	assert.NoError(t, err)
	w := ut.PerformRequest(engine, "GET", "/values",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(time.Minute*15), result["LoginTTL"])
	assert.Equal(t, float64(5), result["LoginFailures"])
	assert.Equal(t, float64(10), result["IpLoginFailures"])
	assert.Equal(t, float64(1), result["PwdStrategy"])
	assert.Equal(t, float64(time.Hour*24*365), result["PwdTTL"])
}

func TestGetSpecify(t *testing.T) {
	u := url.URL{Path: "/values"}
	query := u.Query()
	query.Add("keys", "LoginTTL")
	query.Add("keys", "PwdTTL")
	u.RawQuery = query.Encode()
	w := ut.PerformRequest(engine, "GET", u.RequestURI(),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, float64(time.Minute*15), result["LoginTTL"])
	assert.Equal(t, float64(time.Hour*24*365), result["PwdTTL"])
}

func TestGetBadService(t *testing.T) {
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	w := ut.PerformRequest(engine, "GET", "/values",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

func TestRemove(t *testing.T) {
	err := service.Reset()
	assert.NoError(t, err)
	w := ut.PerformRequest(engine, "DELETE", fmt.Sprintf(`/values/%s`, "LoginTTL"),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 204, resp.StatusCode())

	var data M
	data, err = service.Get("LoginTTL")
	assert.NoError(t, err)
	assert.NotNil(t, data["LoginTTL"])
}

func TestRemoveBadService(t *testing.T) {
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	w := ut.PerformRequest(engine, "DELETE", fmt.Sprintf(`/values/%s`, "LoginFailures"),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}
