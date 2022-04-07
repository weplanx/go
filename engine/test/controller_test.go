package test

import (
	"bytes"
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/password"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"net/http/httptest"
	"testing"
)

var VerifyId primitive.ObjectID

func TestActionsCreate(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"name":  "alpha",
		"batch": "624a8facb4e5d150793d6353",
		"code":  "5auBnD$L",
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/verify", bytes.NewBuffer(body))
	req.Header.Set("wpx-action", "create")
	req.Header.Add("wpx-format-doc", "batch:oid")
	req.Header.Add("wpx-format-doc", "code:password")
	r.ServeHTTP(res, req)
	assert.Equal(t, 201, res.Code)

	var result M
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &result); err != nil {
		t.Error(err)
	}
	var data M
	if err = db.Collection("verify").
		FindOne(context.TODO(), M{"name": "alpha"}).
		Decode(&data); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, data)
	VerifyId = data["_id"].(primitive.ObjectID)
	assert.Equal(t, "alpha", data["name"])
	batch, _ := primitive.ObjectIDFromHex("624a8facb4e5d150793d6353")
	assert.Equal(t, batch, data["batch"])
	assert.NoError(t, password.Verify("5auBnD$L", data["code"].(string)))
}

func TestActionsBulkCreate(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(services)
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(body))
	req.Header.Set("wpx-action", "bulk-create")
	r.ServeHTTP(res, req)
	assert.Equal(t, 201, res.Code)

	var result M
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &result); err != nil {
		t.Error(err)
	}

	var cursor *mongo.Cursor
	if cursor, err = db.Collection("services").Find(context.TODO(), bson.M{}); err != nil {
		t.Error(err)
	}
	data := make([]M, 0)
	if err = cursor.All(context.TODO(), &data); err != nil {
		t.Error(err)
	}
	assert.Len(t, data, 50)
	Comparison(t, services, data)
}

func TestActionsBulkDelete(t *testing.T) {
	var err error
	var cursor *mongo.Cursor
	option := options.Find().SetLimit(10)
	if cursor, err = db.Collection("services").Find(context.TODO(), bson.M{}, option); err != nil {
		t.Error(err)
	}
	data := make([]M, 10)
	if err = cursor.All(context.TODO(), &data); err != nil {
		t.Error(err)
	}
	oids := make([]primitive.ObjectID, 10)
	ids := make([]string, 10)
	for k, v := range data {
		oids[k] = v["_id"].(primitive.ObjectID)
		ids[k] = oids[k].Hex()
	}

	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"_id": M{"$in": ids},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(body))
	req.Header.Set("wpx-action", "bulk-delete")
	req.Header.Add("wpx-format-filter", "_id.$in:oids")

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	count, err := db.Collection("services").
		CountDocuments(context.TODO(), bson.M{"_id": bson.M{"$in": oids}})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(0), count)
}

func TestCount(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/services/_count", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 204, res.Code)
	assert.Equal(t, "40", res.Header().Get("wpx-total"))
}

func TestExists(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/services/_exists", nil)
	query := req.URL.Query()
	query.Set("filter", `{"number":"35433318"}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 204, res.Code)
	assert.Equal(t, "true", res.Header().Get("wpx-exists"))
}

func TestGet(t *testing.T) {
	var err error
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	var values []M
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &values); err != nil {
		t.Error(err)
	}
	assert.Len(t, values, 40)
	var cursor *mongo.Cursor
	if cursor, err = db.Collection("services").Find(context.TODO(), bson.M{}); err != nil {
		t.Error(err)
	}
	data := make([]M, 0)
	if err = cursor.All(context.TODO(), &data); err != nil {
		t.Error(err)
	}
	assert.Len(t, data, 40)
	Comparison(t, data, values)
}

func TestGetFindOne(t *testing.T) {
	var err error
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	req.Header.Set("wpx-type", "find-one")
	query := req.URL.Query()
	query.Set("filter", `{"number":"35433318"}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	var value M
	if err = jsoniter.Unmarshal(res.Body.Bytes(), &value); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, value)
	var data M
	if err = db.Collection("services").
		FindOne(context.TODO(), bson.M{"number": "35433318"}).
		Decode(&data); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, data)
	assert.Equal(t, data["name"], value["name"])
	assert.Equal(t, data["price"], value["price"])
}

func TestGetFindByPage(t *testing.T) {
	var err error
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	req.Header.Set("wpx-type", "find-by-page")
	req.Header.Set("wpx-page", "2")
	req.Header.Set("wpx-page-size", "5")
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "40", res.Header().Get("wpx-total"))
	var values []M
	if err = jsoniter.Unmarshal(res.Body.Bytes(), &values); err != nil {
		t.Error(err)
	}
	assert.Len(t, values, 5)
	option := options.Find().
		SetSort(bson.M{"_id": -1}).
		SetLimit(5).
		SetSkip(5)
	var cursor *mongo.Cursor
	if cursor, err = db.Collection("services").
		Find(context.TODO(), bson.M{}, option); err != nil {
		t.Error(err)
	}
	var data []M
	if err = cursor.All(context.TODO(), &data); err != nil {
		t.Error(err)
	}
	Comparison(t, data, values)
}

func TestById(t *testing.T) {
	var err error
	var data M
	if err = db.Collection("services").
		FindOne(context.TODO(), bson.M{"number": "35433318"}).
		Decode(&data); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, data)
	url := fmt.Sprintf(`/services/%s`, data["_id"].(primitive.ObjectID).Hex())
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	var value M
	if err = jsoniter.Unmarshal(res.Body.Bytes(), &value); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, value)
	assert.Equal(t, data["name"], value["name"])
	assert.Equal(t, data["price"], value["price"])
}

func TestPatch(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"$set": M{
			"price": 9981,
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query := req.URL.Query()
	query.Set("filter", `{"number":{"$in":["35433318","08945471"]}}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	var cursor *mongo.Cursor
	filter := bson.M{"number": bson.M{"$in": bson.A{"35433318", "08945471"}}}
	if cursor, err = db.Collection("services").Find(context.TODO(), filter); err != nil {
		t.Error(err)
	}
	var data []M
	if err = cursor.All(context.TODO(), &data); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, data)
	hmap := make(map[string]M)
	for _, v := range services {
		hmap[v["number"].(string)] = v
	}
	for _, v := range data {
		assert.NotEmpty(t, hmap[v["number"].(string)])
		doc := hmap[v["number"].(string)]
		assert.Equal(t, doc["name"], v["name"])
		assert.Equal(t, float64(9981), v["price"])
	}
}

func TestPatchById(t *testing.T) {
	var err error
	var data M
	if err = db.Collection("services").
		FindOne(context.TODO(), bson.M{"number": "84106393"}).
		Decode(&data); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, data)
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"$set": M{
			"price": 7749,
		},
	})
	if err != nil {
		t.Error(err)
	}
	url := fmt.Sprintf(`/services/%s`, data["_id"].(primitive.ObjectID).Hex())
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var value M
	if err = db.Collection("services").
		FindOne(context.TODO(), bson.M{"number": "84106393"}).
		Decode(&value); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, value)
	assert.Equal(t, data["name"], value["name"])
	assert.Equal(t, float64(7749), value["price"])
}

func TestPut(t *testing.T) {
	var err error
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"name":  "beta",
		"batch": "624eace416ae8713b57adbcf",
		"code":  "gjcOSPLc",
	})
	if err != nil {
		t.Error(err)
	}
	url := fmt.Sprintf(`/verify/%s`, VerifyId.Hex())
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Add("wpx-format-doc", "batch:oid")
	req.Header.Add("wpx-format-doc", "code:password")
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var data M
	if err = db.Collection("verify").
		FindOne(context.TODO(), M{"name": "beta"}).
		Decode(&data); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, data)
	VerifyId = data["_id"].(primitive.ObjectID)
	assert.Equal(t, "beta", data["name"])
	batch, _ := primitive.ObjectIDFromHex("624eace416ae8713b57adbcf")
	assert.Equal(t, batch, data["batch"])
	assert.NoError(t, password.Verify("gjcOSPLc", data["code"].(string)))
}

func TestDelete(t *testing.T) {
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/verify/%s`, VerifyId.Hex())
	req, _ := http.NewRequest("DELETE", url, nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	count, err := db.Collection("verify").CountDocuments(context.TODO(), M{"_id": VerifyId})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(0), count)
}
