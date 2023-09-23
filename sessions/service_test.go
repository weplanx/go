package sessions_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/help"
	"testing"
	"time"
)

func TestSetAndVerify(t *testing.T) {
	ctx := context.TODO()
	jti := help.Uuid()
	status := service.Set(ctx, "dev", jti)
	assert.Equal(t, "OK", status)
	count := rdb.Exists(ctx, service.Key("dev")).Val()
	assert.Equal(t, int64(1), count)
	result := rdb.Get(ctx, service.Key("dev")).Val()
	assert.Equal(t, jti, result)

	result1 := service.Verify(ctx, "dev", jti)
	assert.True(t, result1)
	result2 := service.Verify(ctx, "dev", "abc")
	assert.False(t, result2)
}

func TestRenew(t *testing.T) {
	ctx := context.TODO()
	time.Sleep(time.Second * 2)
	prev := rdb.TTL(ctx, service.Key("dev")).Val()
	result := service.Renew(ctx, "dev")
	assert.True(t, result)
	now := rdb.TTL(ctx, service.Key("dev")).Val()
	assert.Less(t, prev, now)
}

func TestServiceRemove(t *testing.T) {
	ctx := context.TODO()
	result := service.Remove(ctx, "dev")
	assert.Equal(t, int64(1), result)
	count := rdb.Exists(ctx, service.Key("dev")).Val()
	assert.Equal(t, int64(0), count)
}
