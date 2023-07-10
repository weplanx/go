package sessions_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestService_Set(t *testing.T) {
	err := service.Set(context.TODO(), "dev", "dev")
	assert.NoError(t, err)

	count, err := rdb.Exists(context.TODO(), service.Key("dev")).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	result, err := rdb.Get(context.TODO(), service.Key("dev")).Result()
	assert.NoError(t, err)
	assert.Equal(t, "dev", result)
}

func TestService_Verify(t *testing.T) {
	result1, err := service.Verify(context.TODO(), "dev", "dev")
	assert.NoError(t, err)
	assert.True(t, result1)

	result2, err := service.Verify(context.TODO(), "dev", "abc")
	assert.NoError(t, err)
	assert.False(t, result2)
}

func TestService_Renew(t *testing.T) {
	time.Sleep(time.Second * 2)
	prev, err := rdb.TTL(context.TODO(), service.Key("dev")).Result()
	assert.NoError(t, err)
	t.Log(prev)

	err = service.Renew(context.TODO(), "dev")
	assert.NoError(t, err)

	now, err := rdb.TTL(context.TODO(), service.Key("dev")).Result()
	assert.NoError(t, err)
	assert.Less(t, prev, now)
}

func TestService_Remove(t *testing.T) {
	err := service.Remove(context.TODO(), "dev")
	assert.NoError(t, err)

	count, err := rdb.Exists(context.TODO(), service.Key("dev")).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
