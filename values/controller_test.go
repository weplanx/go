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
	err := service.Load()
	assert.NoError(t, err)
	body, _ := sonic.Marshal(M{
		"data": M{
			"xkey": "a_test",
		},
	})
	w := ut.PerformRequest(engine, "PATCH", "/values",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 204, resp.StatusCode())

	entry, err := keyvalue.Get("values")
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	var data M
	err = sonic.Unmarshal(entry.Value(), &data)
	assert.NoError(t, err)
	assert.Equal(t, "a_test", data["xkey"])
}

func TestSetBadService(t *testing.T) {
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	body, _ := sonic.Marshal(M{
		"data": M{
			"xkey": "a_test",
		},
	})
	w := ut.PerformRequest(engine, "PATCH", "/values",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

func TestGetBadValidate(t *testing.T) {
	u := url.URL{Path: "/values"}
	query := u.Query()
	query.Set("keys", `['$#']`)
	u.RawQuery = query.Encode()
	w := ut.PerformRequest(engine, "GET", u.RequestURI(),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestGet(t *testing.T) {
	err := service.Load()
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

	assert.Nil(t, result["xkey"])
	assert.Equal(t, float64(time.Minute*15), result["login_ttl"])
	assert.Equal(t, float64(5), result["login_failures"])
	assert.Equal(t, float64(10), result["ip_login_failures"])
	assert.Equal(t, float64(1), result["pwd_strategy"])
	assert.Equal(t, float64(time.Hour*24*365), result["pwd_ttl"])
}

func TestGetSpecify(t *testing.T) {
	u := url.URL{Path: "/values"}
	query := u.Query()
	keys, _ := sonic.MarshalString(M{
		"login_ttl": 1,
		"pwd_ttl":   1,
	})
	query.Set("keys", keys)
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
	assert.Equal(t, float64(time.Minute*15), result["login_ttl"])
	assert.Equal(t, float64(time.Hour*24*365), result["pwd_ttl"])
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

func TestRemoveBadValidate(t *testing.T) {
	w := ut.PerformRequest(engine, "DELETE", fmt.Sprintf(`/values/%s`, "123456"),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestRemove(t *testing.T) {
	err := service.Load()
	assert.NoError(t, err)
	w := ut.PerformRequest(engine, "DELETE", fmt.Sprintf(`/values/%s`, "login_ttl"),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 204, resp.StatusCode())

	entry, err := keyvalue.Get("values")
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	var data M
	err = sonic.Unmarshal(entry.Value(), &data)
	assert.NoError(t, err)
	assert.Nil(t, data["login_ttl"])
}

func TestRemoveBadService(t *testing.T) {
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	w := ut.PerformRequest(engine, "DELETE", fmt.Sprintf(`/values/%s`, "login_failures"),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())

}
