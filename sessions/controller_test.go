package sessions_test

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/help"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestListsEmpty(t *testing.T) {
	resp, err := R("GET", "/sessions", nil)
	assert.NoError(t, err)
	var uids []string
	err = sonic.Unmarshal(resp.Body(), &uids)
	assert.NoError(t, err)
	assert.Empty(t, uids)
}

func TestRemoveNotUid(t *testing.T) {
	resp, err := R("DELETE", "/sessions/xid", nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestClearEmpty(t *testing.T) {
	resp, err := R("POST", "/sessions/clear", M{})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var r M
	err = sonic.Unmarshal(resp.Body(), &r)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), r["DeletedCount"])
}

func TestRemove(t *testing.T) {
	ctx := context.TODO()
	uid := primitive.NewObjectID().Hex()
	status := service.Set(ctx, uid, help.Uuid())
	assert.Equal(t, "OK", status)

	resp, err := R("DELETE", fmt.Sprintf(`/sessions/%s`, uid), nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var r M
	err = sonic.Unmarshal(resp.Body(), &r)
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

	resp1, err := R("GET", "/sessions", nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode())

	var uids []string
	err = sonic.Unmarshal(resp1.Body(), &uids)
	assert.NoError(t, err)
	assert.ElementsMatch(t, users, uids)

	resp2, err := R("POST", "/sessions/clear", M{})
	assert.NoError(t, err)
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
