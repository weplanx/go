package dsl_test

import (
	"bytes"
	"context"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/utils/passlib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

func TestController_Create(t *testing.T) {
	roles := []string{"635797539db7928aaebbe6e5", "635797c19db7928aaebbe6e6"}
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

type Order struct {
	No       string  `json:"no" faker:"cc_number,unique"`
	Customer string  `json:"customer" faker:"name"`
	Phone    string  `json:"phone" faker:"phone_number"`
	Cost     float64 `json:"cost" faker:"amount"`
	Time     string  `json:"time" faker:"timestamp"`
}

func TestController_BulkCreate(t *testing.T) {
	orders := make([]Order, 10)
	hmap := make(map[string]Order)
	for i := 0; i < 10; i++ {
		err := faker.FakeData(&orders[i])
		assert.NoError(t, err)
		date, err := time.Parse(`2006-01-02 15:04:05`, orders[i].Time)
		assert.NoError(t, err)
		orders[i].Time = date.Format(time.RFC3339)
		hmap[orders[i].No] = orders[i]
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
		oids[i], _ = primitive.ObjectIDFromHex(ids[i].(string))
	}

	cursor, err := db.Collection("orders").Find(context.TODO(), bson.M{"_id": bson.M{"$in": oids}})
	assert.NoError(t, err)
	data := make([]M, 10)
	err = cursor.All(context.TODO(), &data)
	assert.NoError(t, err)

	for _, v := range data {
		order, ok := hmap[v["no"].(string)]
		assert.True(t, ok)
		assert.Equal(t, order.Customer, v["customer"])
		assert.Equal(t, order.Phone, v["phone"])
		assert.Equal(t, order.Cost, v["cost"])
	}
}
