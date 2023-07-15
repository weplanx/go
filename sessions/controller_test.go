package sessions_test

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestListsEmpty(t *testing.T) {
	w := ut.PerformRequest(engine, "GET", "/sessions",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	var uids []string
	err := sonic.Unmarshal(resp.Body(), &uids)
	assert.NoError(t, err)
	assert.Empty(t, uids)
}

func TestRemoveNotUid(t *testing.T) {
	w := ut.PerformRequest(engine, "DELETE", "/sessions/xid",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestClearEmpty(t *testing.T) {
	w := ut.PerformRequest(engine, "POST", "/sessions/clear",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())

	var r M
	err := sonic.Unmarshal(resp.Body(), &r)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), r["DeletedCount"])
}

func TestRemove(t *testing.T) {
	ctx := context.TODO()
	uid := primitive.NewObjectID().Hex()
	status := service.Set(ctx, uid, uuid.New().String())
	assert.Equal(t, "OK", status)

	w := ut.PerformRequest(engine, "DELETE", fmt.Sprintf(`/sessions/%s`, uid),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())

	var r M
	err := sonic.Unmarshal(resp.Body(), &r)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), r["DeletedCount"])

	count := rdb.Exists(ctx, service.Key(uid)).Val()
	assert.Equal(t, int64(0), count)
}

func TestListsAndClear(t *testing.T) {
	ctx := context.TODO()
	users := make([]interface{}, 2000)
	for i := 0; i < 2000; i++ {
		users[i] = primitive.NewObjectID().Hex()
		status := service.Set(ctx, users[i].(string), users[i].(string))
		assert.Equal(t, "OK", status)
	}

	w1 := ut.PerformRequest(engine, "GET", "/sessions",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp1 := w1.Result()
	assert.Equal(t, 200, resp1.StatusCode())

	var uids []string
	err := sonic.Unmarshal(resp1.Body(), &uids)
	assert.NoError(t, err)
	assert.ElementsMatch(t, users, uids)

	w2 := ut.PerformRequest(engine, "POST", "/sessions/clear",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp2 := w2.Result()
	assert.Equal(t, 200, resp2.StatusCode())

	var r M
	err = sonic.Unmarshal(resp2.Body(), &r)
	assert.NoError(t, err)
	assert.Equal(t, float64(2000), r["DeletedCount"])

	usersKey := make([]string, 2000)
	for i := 0; i < 2000; i++ {
		usersKey[i] = service.Key(users[i].(string))
	}
	count := rdb.Exists(ctx, usersKey...).Val()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
