package dsl_test

import (
	"bytes"
	"context"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/utils/dsl"
	"github.com/weplanx/utils/passlib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestController_Create(t *testing.T) {
	roles := []string{"635797539db7928aaebbe6e5", "635797c19db7928aaebbe6e6"}
	body, _ := sonic.Marshal(dsl.CreateDto{
		Data: M{
			"name":       "weplanx",
			"password":   "5auBnD$L",
			"department": "624a8facb4e5d150793d6353",
			"roles":      roles,
		},
		Format: M{
			"password":   "password",
			"department": "oid",
			"roles":      "oids",
		},
	})
	w := ut.PerformRequest(r, "POST", "/users",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 201, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	id, _ := primitive.ObjectIDFromHex(result["InsertedID"].(string))
	var data M
	err = db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	assert.NoError(t, err)
	assert.Equal(t, "weplanx", data["name"])
	check, err := passlib.Verify("5auBnD$L", data["password"].(string))
	assert.NoError(t, err)
	assert.True(t, check)
	departmentId, _ := primitive.ObjectIDFromHex("624a8facb4e5d150793d6353")
	assert.Equal(t, departmentId, data["department"])
	roleIds := make([]primitive.ObjectID, len(roles))
	for k, v := range roles {
		roleIds[k], _ = primitive.ObjectIDFromHex(v)
	}
	assert.ElementsMatch(t, roleIds, data["roles"])
}
