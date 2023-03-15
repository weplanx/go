package values_test

import (
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/utils/values"
	"sync"
	"testing"
	"time"
)

func TestLoadExistsValues(t *testing.T) {
	err := service.Load()
	assert.NoError(t, err)
	err = service.Load()
	assert.NoError(t, err)
}

func TestLoadBadValues(t *testing.T) {
	_, err := keyvalue.Put("values", []byte("abc"))
	assert.NoError(t, err)
	err = service.Load()
	assert.Error(t, err)
}

func TestLoadBucketCleared(t *testing.T) {
	time.Sleep(time.Second)
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := service.Load()
		assert.Error(t, err)
		wg.Done()
	}()
	go func() {
		// 误执行
		err := js.DeleteKeyValue("dev")
		assert.NoError(t, err)
	}()
	wg.Wait()
	err = service.Load()
	assert.Error(t, err)
}

func TestSyncBucketCleared(t *testing.T) {
	err := service.Sync(nil)
	assert.Error(t, err)
}

func TestSync(t *testing.T) {
	var err error
	keyvalue, err = js.CreateKeyValue(&nats.KeyValueConfig{Bucket: "dev"})
	assert.NoError(t, err)

	option := values.SyncOption{
		Updated: make(chan *values.DynamicValues),
		Err:     make(chan error),
	}
	go service.Sync(&option)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		select {
		case x := <-option.Updated:
			assert.Equal(t, values.DEFAULT.LoginTTL, x.LoginTTL)
			assert.Equal(t, values.DEFAULT.LoginFailures, x.LoginFailures)
			assert.Equal(t, values.DEFAULT.IpLoginFailures, x.IpLoginFailures)
			assert.Equal(t, values.DEFAULT.PwdStrategy, x.PwdStrategy)
			assert.Equal(t, values.DEFAULT.PwdTTL, x.PwdTTL)
			assert.Equal(t, "tencent", x.Cloud)
			break
		}
		wg.Done()
	}()
	go func() {
		select {
		case e := <-option.Err:
			assert.Error(t, e)
			break
		}
		wg.Done()
	}()

	time.Sleep(time.Second * 3)

	err = service.Set(M{
		"cloud": "tencent",
	})
	assert.NoError(t, err)

	_, err = keyvalue.Put("values", []byte("abc"))
	assert.NoError(t, err)

	wg.Wait()
}

func TestSetBadValues(t *testing.T) {
	_, err := keyvalue.Put("values", []byte("abc"))
	assert.NoError(t, err)

	err = service.Set(M{})
	assert.Error(t, err)
}

func TestGetBadValues(t *testing.T) {
	_, err := service.Get(map[string]int64{})
	assert.Error(t, err)
}

func TestGetSECRETValues(t *testing.T) {
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	err = service.Load()
	assert.NoError(t, err)

	err = service.Set(M{
		"tencent_secret_id":  "123456",
		"tencent_secret_key": "abc",
		"feishu_app_secret":  "",
	})
	assert.NoError(t, err)

	v, err := service.Get(map[string]int64{
		"tencent_secret_id":  1,
		"tencent_secret_key": 1,
		"feishu_app_secret":  1,
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(v))
	assert.Equal(t, "123456", v["tencent_secret_id"])
	assert.Equal(t, "*", v["tencent_secret_key"])
	assert.Equal(t, "-", v["feishu_app_secret"])
}

func TestUpdateBucketCleared(t *testing.T) {
	err := js.DeleteKeyValue("dev")
	assert.NoError(t, err)

	err = service.Update(map[string]interface{}{
		"tencent_secret_id": "654321",
	})
	assert.Error(t, err)
}
