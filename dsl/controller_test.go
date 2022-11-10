package dsl_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
	"github.com/weplanx/utils/passlib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/url"
	"testing"
	"time"
)

var roles = []string{"635797539db7928aaebbe6e5", "635797c19db7928aaebbe6e6"}
var userId string

func TestCreate(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"name":       "weplanx",
			"password":   "5auBnD$L",
			"department": "624a8facb4e5d150793d6353",
			"roles":      roles,
		},
		"format": M{
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
	userId = result["InsertedID"].(string)
	id, _ := primitive.ObjectIDFromHex(userId)
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

type Order struct {
	No       string             `json:"no" faker:"cc_number,unique"`
	Customer string             `json:"customer" faker:"name"`
	Phone    string             `json:"phone" faker:"phone_number"`
	Cost     float64            `json:"cost" faker:"amount"`
	Time     string             `json:"time" faker:"timestamp"`
	TmpTime  time.Time          `json:"-" faker:"-"`
	TmpId    primitive.ObjectID `json:"-" faker:"-"`
}

var orderIds []string
var orderMap map[string]Order

func TestBulkCreate(t *testing.T) {
	orders := make([]Order, 10)
	orderMap = make(map[string]Order)
	for i := 0; i < 10; i++ {
		err := faker.FakeData(&orders[i])
		assert.NoError(t, err)
		orders[i].TmpTime, err = time.Parse(`2006-01-02 15:04:05`, orders[i].Time)
		assert.NoError(t, err)
		orders[i].Time = orders[i].TmpTime.Format(time.RFC3339)
		orderMap[orders[i].No] = orders[i]
	}
	body, _ := sonic.Marshal(M{
		"data": orders,
		"format": M{
			"time": "date",
		},
	})
	w := ut.PerformRequest(r, "POST", "/orders/bulk-create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 201, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	ids := result["InsertedIDs"].([]interface{})
	assert.Equal(t, 10, len(ids))
	oids := make([]primitive.ObjectID, 10)
	for i := 0; i < 10; i++ {
		orderIds = append(orderIds, ids[i].(string))
		oids[i], _ = primitive.ObjectIDFromHex(ids[i].(string))
	}

	cursor, err := db.Collection("orders").Find(context.TODO(), bson.M{"_id": bson.M{"$in": oids}})
	assert.NoError(t, err)
	data := make([]M, 10)
	err = cursor.All(context.TODO(), &data)
	assert.NoError(t, err)

	for _, v := range data {
		order, ok := orderMap[v["no"].(string)]
		assert.True(t, ok)
		assert.Equal(t, order.Customer, v["customer"])
		assert.Equal(t, order.Phone, v["phone"])
		assert.Equal(t, order.Cost, v["cost"])
		assert.Equal(t, primitive.NewDateTimeFromTime(order.TmpTime), v["time"])
	}
}

func TestSize(t *testing.T) {
	w := ut.PerformRequest(r, "GET", "/orders/_size",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Empty(t, resp.Body())
	assert.Equal(t, "10", resp.Header.Get("x-total"))
	assert.Equal(t, 204, resp.StatusCode())
}

func TestSizeWithFilterAndFormat(t *testing.T) {
	u := url.URL{Path: "/orders/_size"}
	oids := orderIds[:5]
	filter, _ := sonic.MarshalString(M{
		"_id": M{"$in": oids},
	})
	query := u.Query()
	query.Set("filter", filter)
	format, _ := sonic.MarshalString(M{
		"_id.$in": "oids",
	})
	query.Set("format", format)
	u.RawQuery = query.Encode()
	w := ut.PerformRequest(r, "GET", u.RequestURI(),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Empty(t, resp.Body())
	assert.Equal(t, "5", resp.Header.Get("x-total"))
	assert.Equal(t, 204, resp.StatusCode())
}

func TestController_Find(t *testing.T) {
	w := ut.PerformRequest(r, "GET", "/orders",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result []M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	for _, v := range result {
		order, ok := orderMap[v["no"].(string)]
		assert.True(t, ok)
		assert.Equal(t, order.Customer, v["customer"])
		assert.Equal(t, order.Phone, v["phone"])
		assert.Equal(t, order.Cost, v["cost"])
		date, _ := time.Parse(time.RFC3339, v["time"].(string))
		assert.Equal(t, order.TmpTime.Unix(), date.Unix())
	}
}

func TestController_FindOne(t *testing.T) {
	u := url.URL{Path: "/users/_one"}
	filter, _ := sonic.MarshalString(M{"name": "weplanx"})
	query := u.Query()
	query.Set("filter", filter)
	u.RawQuery = query.Encode()
	w := ut.PerformRequest(r, "GET", u.RequestURI(),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	check, err := passlib.Verify("5auBnD$L", result["password"].(string))
	assert.NoError(t, err)
	assert.True(t, check)
	assert.Equal(t, "624a8facb4e5d150793d6353", result["department"])
	assert.ElementsMatch(t, roles, result["roles"])
}

func TestController_FindById(t *testing.T) {
	w := ut.PerformRequest(r, "GET", fmt.Sprintf(`/users/%s`, userId),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	check, err := passlib.Verify("5auBnD$L", result["password"].(string))
	assert.NoError(t, err)
	assert.True(t, check)
	assert.Equal(t, "624a8facb4e5d150793d6353", result["department"])
	assert.ElementsMatch(t, roles, result["roles"])
}

func TestController_Update(t *testing.T) {
	u := url.URL{Path: fmt.Sprintf(`/users`)}
	filter, _ := sonic.MarshalString(M{"name": "weplanx"})
	query := u.Query()
	query.Set("filter", filter)
	u.RawQuery = query.Encode()
	body, _ := sonic.Marshal(M{
		"data": M{
			"$set": M{
				"department": "63579b8b9db7928aaebbe705",
			},
		},
		"format": M{
			"$set.department": "oid",
		},
	})
	w := ut.PerformRequest(r, "PATCH", u.RequestURI(),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	var data M
	id, _ := primitive.ObjectIDFromHex(userId)
	err = db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	assert.NoError(t, err)
	assert.Equal(t, "weplanx", data["name"])
	check, err := passlib.Verify("5auBnD$L", data["password"].(string))
	assert.NoError(t, err)
	assert.True(t, check)
	departmentId, _ := primitive.ObjectIDFromHex("63579b8b9db7928aaebbe705")
	assert.Equal(t, departmentId, data["department"])
	roleIds := make([]primitive.ObjectID, len(roles))
	for k, v := range roles {
		roleIds[k], _ = primitive.ObjectIDFromHex(v)
	}
	assert.ElementsMatch(t, roleIds, data["roles"])
}

func TestController_UpdateById(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"$set": M{
				"department": "62cbf9ac465f45091e981b1e",
			},
		},
		"format": M{
			"$set.department": "oid",
		},
	})
	w := ut.PerformRequest(r, "PATCH", fmt.Sprintf(`/users/%s`, userId),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	var data M
	id, _ := primitive.ObjectIDFromHex(userId)
	err = db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	assert.NoError(t, err)
	assert.Equal(t, "weplanx", data["name"])
	check, err := passlib.Verify("5auBnD$L", data["password"].(string))
	assert.NoError(t, err)
	assert.True(t, check)
	departmentId, _ := primitive.ObjectIDFromHex("62cbf9ac465f45091e981b1e")
	assert.Equal(t, departmentId, data["department"])
	roleIds := make([]primitive.ObjectID, len(roles))
	for k, v := range roles {
		roleIds[k], _ = primitive.ObjectIDFromHex(v)
	}
	assert.ElementsMatch(t, roleIds, data["roles"])
}

func TestController_Replace(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"name":       "kain",
			"password":   "123456",
			"department": nil,
			"roles":      []string{},
		},
		"format": M{
			"password":   "password",
			"department": "oid",
			"roles":      "oids",
		},
	})
	w := ut.PerformRequest(r, "PUT", fmt.Sprintf(`/users/%s`, userId),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	var data M
	id, _ := primitive.ObjectIDFromHex(userId)
	err = db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	assert.NoError(t, err)
	assert.Equal(t, "kain", data["name"])
	check, err := passlib.Verify("123456", data["password"].(string))
	assert.NoError(t, err)
	assert.True(t, check)
	assert.Empty(t, data["department"])
	assert.Equal(t, primitive.A{}, data["roles"])
}

func TestController_Delete(t *testing.T) {
	w := ut.PerformRequest(r, "DELETE", fmt.Sprintf(`/users/%s`, userId),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), result["DeletedCount"])

	var data M
	id, _ := primitive.ObjectIDFromHex(userId)
	err = db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	assert.Error(t, err)
	assert.Equal(t, err, mongo.ErrNoDocuments)
}

func TestController_BulkDelete(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"_id": M{"$in": orderIds[5:]},
		},
		"format": M{
			"_id.$in": "oids",
		},
	})
	w := ut.PerformRequest(r, "POST", "/orders/bulk-delete",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	t.Log(result)

	var ids []primitive.ObjectID
	for _, v := range orderIds[5:] {
		id, _ := primitive.ObjectIDFromHex(v)
		ids = append(ids, id)
	}
	n, err := db.Collection("orders").CountDocuments(context.TODO(), bson.M{"_id": bson.M{"$in": ids}})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestController_Sort(t *testing.T) {
	sources := orderIds[:5]
	sources = funk.Reverse(sources).([]string)
	body, _ := sonic.Marshal(M{
		"data": sources,
	})
	w := ut.PerformRequest(r, "POST", "/orders/sort",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 204, resp.StatusCode())
	assert.Empty(t, resp.Body())

	var ids []primitive.ObjectID
	for _, v := range orderIds[5:] {
		id, _ := primitive.ObjectIDFromHex(v)
		ids = append(ids, id)
	}
	cursor, err := db.Collection("orders").Find(context.TODO(), bson.M{"_id": bson.M{"$in": ids}})
	assert.NoError(t, err)
	var data []M
	err = cursor.All(context.TODO(), &data)
	assert.NoError(t, err)

	for _, v := range data {
		index := v["sort"].(int)
		assert.Equal(t, sources[index], v["_id"].(primitive.ObjectID).Hex())
	}
}
