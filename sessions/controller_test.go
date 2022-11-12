package sessions_test

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

var users []interface{}

func TestController_Lists(t *testing.T) {
	users = make([]interface{}, 2000)
	for i := 0; i < 2000; i++ {
		users[i] = "user" + strconv.Itoa(i)
		err := service.Set(context.TODO(), users[i].(string), users[i].(string))
		assert.NoError(t, err)
	}

	time.Sleep(time.Second)
	w := ut.PerformRequest(r, "GET", "/sessions",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())

	var data []string
	err := sonic.Unmarshal(resp.Body(), &data)
	assert.NoError(t, err)
	assert.ElementsMatch(t, users, data)
}

func TestController_Remove(t *testing.T) {
	w := ut.PerformRequest(r, "DELETE", "/sessions/user0",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 204, resp.StatusCode())
	assert.Empty(t, resp.Body())

	count, err := rdb.Exists(context.TODO(), service.Key("user0")).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestController_Clear(t *testing.T) {
	w := ut.PerformRequest(r, "DELETE", "/sessions",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 204, resp.StatusCode())
	assert.Empty(t, resp.Body())

	usersKey := make([]string, 2000)
	for i := 0; i < 2000; i++ {
		usersKey[i] = service.Key(users[i].(string))
	}
	count, err := rdb.Exists(context.TODO(), usersKey...).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
