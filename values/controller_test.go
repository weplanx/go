package values_test

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSetBadValidate(t *testing.T) {
	resp1, err := R("PATCH", "/values", M{
		"data": M{
			"key1": "value1",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp1.StatusCode())

	resp2, err := R("PATCH", "/values", M{
		"update": M{
			"Wechat$$$": "abcdefg",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode())
}

func TestSet(t *testing.T) {
	err := Reset()
	assert.NoError(t, err)
	resp, err := R("PATCH", "/values", M{
		"update": M{
			"Wechat": "abcdefg",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode())

	var data M
	data, err = service.Get("Wechat")
	assert.NoError(t, err)
	assert.Equal(t, "abcdefg", data["Wechat"])
}

func TestSetBadService(t *testing.T) {
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	resp, err := R("PATCH", "/values", M{
		"update": M{
			"Wechat": "abcdefg",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

func TestGetBadValidate(t *testing.T) {
	u := U("/values", Params{
		{"keys", "LoginTTL$$$"},
	})
	resp, err := R("GET", u, nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestGet(t *testing.T) {
	err := Reset()
	assert.NoError(t, err)
	resp, err := R("GET", "/values", nil)
	assert.NoError(t, err)
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
	u := U("/values", Params{
		{"keys", "LoginTTL"},
		{"keys", "PwdTTL"},
	})
	resp, err := R("GET", u, nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, float64(time.Minute*15), result["LoginTTL"])
	assert.Equal(t, float64(time.Hour*24*365), result["PwdTTL"])
}

func TestGetBadService(t *testing.T) {
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	resp, err := R("GET", "/values", nil)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

func TestRemoveBadValidate(t *testing.T) {
	err := Reset()
	assert.NoError(t, err)
	resp, err := R("DELETE", fmt.Sprintf(`/values/%s`, "LoginTTL$$$"), nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
	t.Log(string(resp.Body()))
}

func TestRemove(t *testing.T) {
	err := Reset()
	assert.NoError(t, err)

	resp, err := R("DELETE", fmt.Sprintf(`/values/%s`, "LoginTTL"), nil)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode())

	var data M
	data, err = service.Get("LoginTTL")
	assert.NoError(t, err)
	assert.Nil(t, data["LoginTTL"])
}

func TestRemoveBadService(t *testing.T) {
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	resp, err := R("DELETE", fmt.Sprintf(`/values/%s`, "LoginFailures"), nil)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}
