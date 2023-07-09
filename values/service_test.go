package values_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go-wpx/values"
	"testing"
)

func TestKeys(t *testing.T) {
	keys, err := keyvalue.Keys()
	assert.NoError(t, err)
	t.Log(keys)
}

func TestService_Fetch(t *testing.T) {
	data := make(map[string]interface{})
	err := service.Fetch(&data)
	assert.NoError(t, err)
	t.Log(data)
}

func TestService_Set(t *testing.T) {
	err := service.Set(map[string]interface{}{
		"TencentSecretKey": "123456",
		"Wechat":           "abcdefg",
	})
	assert.NoError(t, err)
}

func TestService_Get(t *testing.T) {
	data := make(map[string]interface{})
	keys := []string{
		"LoginFailures",
		"Cloud",
		"TencentSecretKey",
		"LarkAppSecret",
		"Wechat",
		"XXX",
	}
	err := service.Get(keys, &data)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(data))
	assert.Equal(t, "", data["Cloud"])
	assert.Equal(t, "-", data["LarkAppSecret"])
	assert.Equal(t, int64(5), data["LoginFailures"])
	assert.Equal(t, "*", data["TencentSecretKey"])
	assert.Equal(t, "abcdefg", data["Wechat"])
	assert.Nil(t, data["XXX"])
}

func TestService_Update(t *testing.T) {
	data := values.DEFAULT
	data.IpLoginFailures = 3
	data.Cloud = "tencent"
	data.LarkAppId = "asdasd"
	data.LarkAppSecret = "123456"
	err := service.Update(data)
	assert.NoError(t, err)
	result := make(map[string]interface{})
	err = service.Fetch(&result)
	assert.NoError(t, err)
	assert.Equal(t, data.IpLoginFailures, result["IpLoginFailures"])
	assert.Equal(t, data.Cloud, result["Cloud"])
	assert.Equal(t, data.LarkAppId, result["LarkAppId"])
	assert.Equal(t, "*", result["LarkAppSecret"])
}

func TestService_Remove(t *testing.T) {
	keys := []string{"LarkAppId", "LarkAppSecret"}
	err := service.Remove(keys)
	assert.NoError(t, err)
	result := make(map[string]interface{})
	err = service.Get(keys, &result)
	t.Log(result)
	assert.Nil(t, result["LarkAppId"])
	assert.Equal(t, "-", result["LarkAppSecret"])
}
