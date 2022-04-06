package test

import (
	"bytes"
	"context"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var createInsertedID string

func TestCreate(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"action": "create",
		"doc": M{
			"name":   "test1",
			"parent": nil,
		},
		"format": M{
			"parent": "object_id",
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/pages", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t, 201, res.Code)

	var result M
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &result); err != nil {
		panic(err)
	}
	filter := M{"name": "test1"}
	count, err := db.Collection("privileges").
		CountDocuments(context.TODO(), filter)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(1), count)
	createInsertedID = result["InsertedID"].(string)
}
