package captcha_test

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/utils/captcha"
	"log"
	"os"
	"testing"
	"time"
)

var x *captcha.Captcha

func TestMain(m *testing.M) {
	opts, err := redis.ParseURL(os.Getenv("DATABASE_REDIS"))
	if err != nil {
		log.Fatalln(err)
	}
	x = captcha.New("dev", redis.NewClient(opts))
	os.Exit(m.Run())
}

func TestCreate(t *testing.T) {
	var err error
	err = x.Create(context.TODO(), "dev1", "abcd", time.Second*60)
	assert.NoError(t, err)
	var ttl time.Duration
	ttl, err = x.Redis.TTL(context.TODO(), x.Key("dev1")).Result()
	assert.NoError(t, err)
	t.Log(ttl.Seconds())
	err = x.Create(context.TODO(), "dev2", "abcd", time.Millisecond)
	assert.NoError(t, err)
}

func TestVerify(t *testing.T) {
	var err error
	err = x.Verify(context.TODO(), "dev1", "abc")
	assert.ErrorIs(t, err, captcha.ErrCaptchaInconsistent)
	err = x.Verify(context.TODO(), "dev1", "abcd")
	assert.NoError(t, err)
	time.Sleep(time.Nanosecond)
	err = x.Verify(context.TODO(), "dev2", "abcd")
	assert.Error(t, err)
}

func TestDelete(t *testing.T) {
	var err error
	var exists bool
	exists, err = x.Exists(context.TODO(), "dev1")
	assert.NoError(t, err)
	assert.True(t, exists)
	err = x.Delete(context.TODO(), "dev1")
	assert.NoError(t, err)
	exists, err = x.Exists(context.TODO(), "dev1")
	assert.NoError(t, err)
	assert.False(t, exists)
}
