package test

import (
	"bytes"
	"context"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestActionsCreate(t *testing.T) {
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
		t.Error(err)
	}
	filter := M{"name": "test1"}
	count, err := db.Collection("pages").
		CountDocuments(context.TODO(), filter)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(1), count)
}

func TestActionsBulkCreate(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"action": "bulk-create",
		"docs":   services,
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t, 201, res.Code)

	var result M
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &result); err != nil {
		t.Error(err)
	}

	count, err := db.Collection("services").
		CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(50), count)
}

//func TestActionsBulkDelete(t *testing.T) {
//	res := httptest.NewRecorder()
//	body, err := jsoniter.Marshal(M{
//		"action": "bulk-delete",
//		"filter": M{"number": M{"$in": []string{"35568587", "92044481"}}},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(body))
//}
