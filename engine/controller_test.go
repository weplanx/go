package engine

import (
	"bytes"
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/helper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"net/http/httptest"
	"testing"
)

var UserId primitive.ObjectID

func TestActionsNotAction(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/services", nil)
	req.Header.Set("wpx-action", "bulk-no")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestActionsCreateBodyErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer([]byte("hi")))
	req.Header.Set("wpx-action", "create")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestActionsCreateBodyEmpty(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("wpx-action", "create")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestActionsCreateFormatDocErr(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"name":  "alpha",
		"batch": "xxx",
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("wpx-action", "create")
	req.Header.Add("wpx-format-doc", "batch:oid")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

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
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
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
	if err = db.Collection("users").
		FindOne(context.TODO(), M{"name": "alpha"}).
		Decode(&data); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, data)
	UserId = data["_id"].(primitive.ObjectID)
	assert.Equal(t, "alpha", data["name"])
	batch, _ := primitive.ObjectIDFromHex("624a8facb4e5d150793d6353")
	assert.Equal(t, batch, data["batch"])
	assert.NoError(t, helper.PasswordVerify("5auBnD$L", data["code"].(string)))

	Event(t, "create")
}

func TestActionsBulkCreateBodyErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer([]byte("hi")))
	req.Header.Set("wpx-action", "bulk-create")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestActionsBulkCreateBodyEmpty(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer([]byte(`[]`)))
	req.Header.Set("wpx-action", "bulk-create")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestActionsBulkCreateFormatDocErr(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal([]M{
		{
			"name":  "alpha",
			"batch": "xxx",
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("wpx-action", "bulk-create")
	req.Header.Add("wpx-format-doc", "batch:oid")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
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

func TestActionsBulkDeleteBodyErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer([]byte("hi")))
	req.Header.Set("wpx-action", "bulk-delete")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestActionsBulkDeleteBodyEmpty(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("wpx-action", "bulk-delete")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestActionsBulkDeleteDbErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer([]byte(`{"abc":{"$":"v"}}`)))
	req.Header.Set("wpx-action", "bulk-delete")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestActionsBulkDeleteFormatFilterErr(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"_id": M{"$in": []string{"xxx", "xx1"}},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(body))
	req.Header.Set("wpx-action", "bulk-delete")
	req.Header.Add("wpx-format-filter", "_id.$in:oids")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
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

func TestCountUriErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/$$$/_count", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestCountFilterErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/services/_count", nil)
	query := req.URL.Query()
	query.Set("filter", `hi`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestCountDbErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/services/_count", nil)
	query := req.URL.Query()
	query.Set("filter", `{"abc":{"$":"v"}}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestCount(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/services/_count", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 204, res.Code)
	assert.Equal(t, "40", res.Header().Get("wpx-total"))
}

func TestExistsUriErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/$$$/_exists", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestExistsFilterErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/services/_exists", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestExistsDbErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/services/_exists", nil)
	query := req.URL.Query()
	query.Set("filter", `{"abc":{"$":"v"}}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
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

func TestGetUriErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/$$$", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestGetFilterErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	query := req.URL.Query()
	query.Set("filter", `hi`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestGetDbErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	query := req.URL.Query()
	query.Set("filter", `{"abc":{"$":"v"}}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
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

func TestGetWithField(t *testing.T) {
	var err error
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	query := req.URL.Query()
	query.Add("field", "number")
	req.URL.RawQuery = query.Encode()
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
	assert.Equal(t, len(data), len(values))
	hmap := make(map[string]bool)
	for _, v := range data {
		hmap[v["number"].(string)] = true
	}
	for _, v := range values {
		assert.NotNil(t, hmap[v["number"].(string)])
		assert.Empty(t, v["name"])
		assert.Empty(t, v["price"])
	}
}

func TestGetWithFormatFilter(t *testing.T) {
	var err error
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	req.Header.Add("wpx-format-filter", "batch:oid")
	query := req.URL.Query()
	query.Set("filter", `{"batch":"624a8facb4e5d150793d6353"}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	var values []M
	if err = jsoniter.Unmarshal(res.Body.Bytes(), &values); err != nil {
		t.Error(err)
	}
	assert.Len(t, values, 1)
	assert.Equal(t, "alpha", values[0]["name"])
}

func TestGetWithFormatFilterErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	req.Header.Add("wpx-format-filter", "batch:oid")
	query := req.URL.Query()
	query.Set("filter", `{"batch":"xxx"}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestGetWithLimitAndSkip(t *testing.T) {
	var err error
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	req.Header.Set("wpx-limit", "10")
	req.Header.Set("wpx-skip", "10")
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	var values []M
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &values); err != nil {
		t.Error(err)
	}
	assert.Len(t, values, 10)
	option := options.Find().
		SetSort(bson.M{"_id": -1}).
		SetLimit(10).
		SetSkip(10)
	var cursor *mongo.Cursor
	if cursor, err = db.Collection("services").Find(context.TODO(), bson.M{}, option); err != nil {
		t.Error(err)
	}
	data := make([]M, 0)
	if err = cursor.All(context.TODO(), &data); err != nil {
		t.Error(err)
	}
	assert.Len(t, data, 10)
	Comparison(t, data, values)
}

func TestGetFindOneDbErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	req.Header.Set("wpx-type", "find-one")
	query := req.URL.Query()
	query.Set("filter", `{"abc":{"$":"v"}}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
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

func TestGetFindOneWithFormatFilterErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	req.Header.Set("wpx-type", "find-one")
	req.Header.Add("wpx-format-filter", "batch:oid")
	query := req.URL.Query()
	query.Set("filter", `{"batch":"xxx"}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestGetFindOneWithFormatFilter(t *testing.T) {
	var err error
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	req.Header.Set("wpx-type", "find-one")
	req.Header.Add("wpx-format-filter", "batch:oid")
	query := req.URL.Query()
	query.Set("filter", `{"batch":"624a8facb4e5d150793d6353"}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	var value M
	if err = jsoniter.Unmarshal(res.Body.Bytes(), &value); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, value)
	assert.Equal(t, "alpha", value["name"])
}

func TestGetFindByPageDbErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	req.Header.Set("wpx-type", "find-by-page")
	query := req.URL.Query()
	query.Set("filter", `{"abc":{"$":"v"}}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
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

func TestGetFindByPageWithFormatFilterErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	req.Header.Set("wpx-type", "find-by-page")
	req.Header.Add("wpx-format-filter", "batch:oid")
	query := req.URL.Query()
	query.Set("filter", `{"batch":"xxx"}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestGetFindByPageWithSort(t *testing.T) {
	var err error
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	req.Header.Set("wpx-type", "find-by-page")
	req.Header.Set("wpx-page", "2")
	req.Header.Set("wpx-page-size", "5")
	query := req.URL.Query()
	query.Add("sort", "price.1")
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "40", res.Header().Get("wpx-total"))
	var values []M
	if err = jsoniter.Unmarshal(res.Body.Bytes(), &values); err != nil {
		t.Error(err)
	}
	assert.Len(t, values, 5)
	option := options.Find().
		SetSort(bson.M{"price": 1}).
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
	for k, v := range data {
		assert.Equal(t, v["name"], values[k]["name"])
	}
}

func TestGetByIdUriErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services/abcd", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestGetFieldErr(t *testing.T) {
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/users/%s`, UserId.Hex())
	req, _ := http.NewRequest("GET", url, nil)
	req.URL.RawQuery = `field=&`
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestGetById(t *testing.T) {
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

func TestGetByIdWithField(t *testing.T) {
	url := fmt.Sprintf(`/users/%s`, UserId.Hex())
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	var value M
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &value); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, value)
	assert.Equal(t, "alpha", value["name"])
	assert.Equal(t, "624a8facb4e5d150793d6353", value["batch"])
	assert.Empty(t, value["code"])
}

func TestGetByIdWithQueryField(t *testing.T) {
	url := fmt.Sprintf(`/users/%s`, UserId.Hex())
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	query := req.URL.Query()
	query.Add("field", "name")
	query.Add("field", "code")
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	var value M
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &value); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, value)
	assert.Equal(t, "alpha", value["name"])
	assert.Empty(t, value["batch"])
	assert.Empty(t, value["code"])
}

func TestPatchUriErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/$$$", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPatchFilterErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/services", nil)
	query := req.URL.Query()
	query.Set("filter", `hi`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPatchBodyErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer([]byte("hi")))
	query := req.URL.Query()
	query.Set("filter", `{"number":{"$in":["35433318","08945471"]}}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPatchBodyEmpty(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer([]byte(`{}`)))
	query := req.URL.Query()
	query.Set("filter", `{"number":{"$in":["35433318","08945471"]}}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPatchDbErr(t *testing.T) {
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
	query.Set("filter", `{"abc":{"$":"v"}}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
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

func TestPatchWithFormatFilterErr(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"$set": M{
			"batch": "624a94c3397dd503069e9654",
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/users", bytes.NewBuffer(body))
	req.Header.Set("wpx-format-filter", "batch:oid")
	req.Header.Set("wpx-format-doc", "batch:oid")
	query := req.URL.Query()
	query.Set("filter", `{"batch":"xxx"}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPatchWithFormatDocErr(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"$set": M{
			"batch": "xxx",
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/users", bytes.NewBuffer(body))
	req.Header.Set("wpx-format-filter", "batch:oid")
	req.Header.Set("wpx-format-doc", "batch:oid")
	query := req.URL.Query()
	query.Set("filter", `{"batch":"624a8facb4e5d150793d6353"}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPatchWithFormat(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"$set": M{
			"batch": "624a94c3397dd503069e9654",
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/users", bytes.NewBuffer(body))
	req.Header.Set("wpx-format-filter", "batch:oid")
	req.Header.Set("wpx-format-doc", "batch:oid")
	query := req.URL.Query()
	query.Set("filter", `{"batch":"624a8facb4e5d150793d6353"}`)
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var value M
	if err = db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": UserId}).
		Decode(&value); err != nil {
		t.Error(err)
	}
	batch, _ := primitive.ObjectIDFromHex("624a94c3397dd503069e9654")
	assert.NotEmpty(t, value)
	assert.Equal(t, value["batch"], batch)

	Event(t, "update")
}

func TestPatchByIdUriErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/services/abcd", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPatchByIdBodyErr(t *testing.T) {
	res := httptest.NewRecorder()
	url := "/services/624a8facb4e5d150793d6353"
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte("hi")))
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPatchByIdBodyEmpty(t *testing.T) {
	res := httptest.NewRecorder()
	url := "/services/624a8facb4e5d150793d6353"
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte(`{}`)))
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPatchByIdDbErr(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"$oid": M{},
	})
	if err != nil {
		t.Error(err)
	}
	url := "/services/624a8facb4e5d150793d6353"
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
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

func TestPatchByIdFormatDocErr(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"$set": M{
			"batch": "xxx",
		},
	})
	if err != nil {
		t.Error(err)
	}
	url := fmt.Sprintf(`/users/%s`, UserId.Hex())
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	req.Header.Set("wpx-format-doc", "batch:oid")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPatchByIdFormatDoc(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"$set": M{
			"batch": "624e8a7928188d874d5c4aed",
		},
	})
	if err != nil {
		t.Error(err)
	}
	url := fmt.Sprintf(`/users/%s`, UserId.Hex())
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	req.Header.Set("wpx-format-doc", "batch:oid")
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var value M
	if err = db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": UserId}).
		Decode(&value); err != nil {
		t.Error(err)
	}
	batch, _ := primitive.ObjectIDFromHex("624e8a7928188d874d5c4aed")
	assert.NotEmpty(t, value)
	assert.Equal(t, value["batch"], batch)

	Event(t, "update")
}

func TestPutUriErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/services/abcd", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPutBodyErr(t *testing.T) {
	res := httptest.NewRecorder()
	url := "/services/624a8facb4e5d150793d6353"
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer([]byte("hi")))
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPutBodyEmpty(t *testing.T) {
	res := httptest.NewRecorder()
	url := "/services/624a8facb4e5d150793d6353"
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(`{}`)))
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPutFormatDocErr(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"name":  "beta",
		"batch": "xxx",
		"code":  "gjcOSPLc",
	})
	if err != nil {
		t.Error(err)
	}
	url := fmt.Sprintf(`/users/%s`, UserId.Hex())
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Set("wpx-format-doc", "batch:oid")
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestPut(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(M{
		"name":  "beta",
		"batch": "624eace416ae8713b57adbcf",
		"code":  "gjcOSPLc",
	})
	if err != nil {
		t.Error(err)
	}
	url := fmt.Sprintf(`/users/%s`, UserId.Hex())
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Add("wpx-format-doc", "batch:oid")
	req.Header.Add("wpx-format-doc", "code:password")
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var data M
	if err = db.Collection("users").
		FindOne(context.TODO(), M{"name": "beta"}).
		Decode(&data); err != nil {
		t.Error(err)
	}
	assert.NotEmpty(t, data)
	UserId = data["_id"].(primitive.ObjectID)
	assert.Equal(t, "beta", data["name"])
	batch, _ := primitive.ObjectIDFromHex("624eace416ae8713b57adbcf")
	assert.Equal(t, batch, data["batch"])
	assert.NoError(t, helper.PasswordVerify("gjcOSPLc", data["code"].(string)))

	Event(t, "update")
}

func TestDeleteUriErr(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/abcd", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

func TestDelete(t *testing.T) {
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/users/%s`, UserId.Hex())
	req, _ := http.NewRequest("DELETE", url, nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	count, err := db.Collection("users").CountDocuments(context.TODO(), M{"_id": UserId})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(0), count)

	Event(t, "delete")
}
