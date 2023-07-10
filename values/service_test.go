package values_test

import (
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go-wpx/values"
	"sync"
	"testing"
)

func TestService_Fetch(t *testing.T) {
	err := service.Reset()
	assert.NoError(t, err)

	data1 := values.DynamicValues{}
	err = service.Fetch(&data1)
	assert.NoError(t, err)
	assert.Equal(t, values.DEFAULT.LoginFailures, data1.LoginFailures)
	assert.Equal(t, values.DEFAULT.IpLoginFailures, data1.IpLoginFailures)
	assert.Equal(t, values.DEFAULT.SessionTTL, data1.SessionTTL)
	assert.Equal(t, values.DEFAULT.LoginTTL, data1.LoginTTL)
	assert.Equal(t, values.DEFAULT.PwdTTL, data1.PwdTTL)

	data2 := make(map[string]interface{})
	err = service.Fetch(&data2)
	assert.NoError(t, err)
	assert.Equal(t, float64(values.DEFAULT.LoginFailures), data2["LoginFailures"])
	assert.Equal(t, float64(values.DEFAULT.IpLoginFailures), data2["IpLoginFailures"])
	assert.Equal(t, float64(values.DEFAULT.SessionTTL), data2["SessionTTL"])
	assert.Equal(t, float64(values.DEFAULT.LoginTTL), data2["LoginTTL"])
	assert.Equal(t, float64(values.DEFAULT.PwdTTL), data2["PwdTTL"])
}

func TestService_Set(t *testing.T) {
	err := service.Set(map[string]interface{}{
		"LoginFailures":    5,
		"TencentSecretKey": "123456",
		"Wechat":           "abcdefg",
	})
	assert.NoError(t, err)
}

func TestService_Get(t *testing.T) {
	data1, err := service.Get()
	assert.NoError(t, err)
	t.Log(data1)
	assert.Equal(t, float64(values.DEFAULT.LoginFailures), data1["LoginFailures"])
	assert.Equal(t, float64(values.DEFAULT.IpLoginFailures), data1["IpLoginFailures"])
	assert.Equal(t, float64(values.DEFAULT.SessionTTL), data1["SessionTTL"])
	assert.Equal(t, float64(values.DEFAULT.PwdStrategy), data1["PwdStrategy"])
	assert.Equal(t, float64(values.DEFAULT.LoginTTL), data1["LoginTTL"])
	assert.Equal(t, float64(values.DEFAULT.PwdTTL), data1["PwdTTL"])
	assert.Equal(t, "*", data1["TencentSecretKey"])
	assert.Equal(t, "abcdefg", data1["Wechat"])

	keys := []string{
		"LoginFailures",
		"Cloud",
		"TencentSecretKey",
		"LarkAppSecret",
		"Wechat",
		"XXX",
	}
	data2, err := service.Get(keys...)
	assert.NoError(t, err)
	t.Log(data2)
	assert.Equal(t, float64(5), data2["LoginFailures"])
	assert.Equal(t, "*", data2["TencentSecretKey"])
	assert.Equal(t, "abcdefg", data2["Wechat"])
	assert.Nil(t, data2["Cloud"])
	assert.Nil(t, data2["LarkAppSecret"])
	assert.Nil(t, data2["XXX"])
}

func TestService_Update(t *testing.T) {
	data := values.DEFAULT
	data.IpLoginFailures = 3
	data.Cloud = "tencent"
	data.LarkAppId = "asdasd"
	data.LarkAppSecret = "123456"
	err := service.Update(data)
	assert.NoError(t, err)
	result, err := service.Get()
	assert.NoError(t, err)
	t.Log(result)
	assert.Equal(t, float64(data.IpLoginFailures), result["IpLoginFailures"])
	assert.Equal(t, data.Cloud, result["Cloud"])
	assert.Equal(t, data.LarkAppId, result["LarkAppId"])
	assert.Equal(t, "*", result["LarkAppSecret"])
}

func TestService_Remove(t *testing.T) {
	keys := []string{"LarkAppId", "LarkAppSecret"}
	err := service.Remove(keys...)
	assert.NoError(t, err)
	result, err := service.Get(keys...)
	t.Log(result)
	assert.Nil(t, result["LarkAppId"])
	assert.Nil(t, result["LarkAppSecret"])
}

func TestService_SyncNotExists(t *testing.T) {
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	err = service.Sync(nil)
	assert.Error(t, err, nats.ErrKeyNotFound)
}

func TestService_Sync(t *testing.T) {
	err := service.Reset()
	assert.NoError(t, err)

	ok := make(chan interface{})
	go service.Sync(ok)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for data := range ok {
			assert.Equal(t, int64(10), data.(*values.DynamicValues).LoginFailures)
			break
		}
		wg.Done()
	}()
	err = service.Set(map[string]interface{}{
		"LoginFailures": 10,
	})
	assert.NoError(t, err)
	wg.Wait()
}
