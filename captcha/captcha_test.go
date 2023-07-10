package captcha_test

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go-wpx/captcha"
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
	x = captcha.New(
		captcha.SetNamespace("dev"),
		captcha.SetRedis(redis.NewClient(opts)),
	)
	os.Exit(m.Run())
}

func TestCreate(t *testing.T) {
	var err error
	ctx := context.TODO()
	err = x.Create(ctx, "dev1", "abcd", time.Second*60)
	assert.NoError(t, err)
	var ttl time.Duration
	ttl, err = x.Redis.TTL(ctx, x.Key("dev1")).Result()
	assert.NoError(t, err)
	t.Log(ttl.Seconds())
	err = x.Create(ctx, "dev2", "abcd", time.Millisecond)
	assert.NoError(t, err)
}

func TestVerify(t *testing.T) {
	var err error
	err = x.Verify(context.TODO(), "dev1", "abc")
	assert.ErrorIs(t, err, captcha.ErrCaptchaInconsistent)
	err = x.Verify(context.TODO(), "dev1", "abcd")
	assert.NoError(t, err)
	time.Sleep(time.Second)
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
