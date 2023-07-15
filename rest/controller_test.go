package rest_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
	"github.com/weplanx/go/passlib"
	"github.com/weplanx/go/rest"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/url"
	"testing"
	"time"
)

var roles = []string{"635797539db7928aaebbe6e5", "635797c19db7928aaebbe6e6"}
var userId string

func TestCreateBadValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{})
	w := ut.PerformRequest(engine, "POST", "/users/create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestCreateBadTransform(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"department": "123",
		},
		"xdata": M{
			"department": "oid",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/users/create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestCreateTxnNotExists(t *testing.T) {
	txn := uuid.New().String()
	body, _ := sonic.Marshal(M{
		"data": M{
			"name":       "weplanx",
			"password":   "5auBnD$L",
			"department": "624a8facb4e5d150793d6353",
			"roles":      roles,
		},
		"xdata": M{
			"password":   "password",
			"department": "oid",
			"roles":      "oids",
		},
		"txn": txn,
	})
	w := ut.PerformRequest(engine, "POST", "/users/create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestCreate(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"name":       "weplanx",
			"password":   "5auBnD$L",
			"department": "624a8facb4e5d150793d6353",
			"roles":      roles,
		},
		"xdata": M{
			"password":   "password",
			"department": "oid",
			"roles":      "oids",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/users/create",
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
	err = service.Db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	t.Log(data)
	assert.NoError(t, err)
	assert.Equal(t, "weplanx", data["name"])
	err = passlib.Verify("5auBnD$L", data["password"].(string))
	assert.NoError(t, err)
	departmentId, _ := primitive.ObjectIDFromHex("624a8facb4e5d150793d6353")
	assert.Equal(t, departmentId, data["department"])
	roleIds := make([]primitive.ObjectID, len(roles))
	for k, v := range roles {
		roleIds[k], _ = primitive.ObjectIDFromHex(v)
	}
	assert.ElementsMatch(t, roleIds, data["roles"])

}

func TestCreateBadDbValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"name": "weplanx",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/users/create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

var projectId string

func TestCreateEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	expire := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	body, _ := sonic.Marshal(M{
		"data": M{
			"name":        "默认项目",
			"namespace":   "default",
			"secret":      "abcd",
			"expire_time": expire,
		},
		"xdata": M{
			"expire_time": "timestamp",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/projects/create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 201, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	projectId = result["InsertedID"].(string)

	select {
	case msg := <-ch:
		assert.Equal(t, "create", msg.Action)
		data := msg.Data.(M)
		assert.Equal(t, "默认项目", data["name"])
		assert.Equal(t, "default", data["namespace"])
		assert.Equal(t, "abcd", data["secret"])
		assert.Equal(t, expire, data["expire_time"])
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestCreateBadEvent(t *testing.T) {
	expire := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	body, _ := sonic.Marshal(M{
		"data": M{
			"name":        "默认项目",
			"namespace":   "default",
			"secret":      "abcd",
			"expire_time": expire,
		},
		"xdata": M{
			"expire_time": "timestamp",
		},
	})
	RemoveStream(t)
	w := ut.PerformRequest(engine, "POST", "/projects/create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
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

func TestBulkCreateBadValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{})
	w := ut.PerformRequest(engine, "POST", "/orders/bulk_create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkCreateBadTransform(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": []Order{
			{
				No:       "123456",
				Customer: "Joe",
				Phone:    "11225566",
				Cost:     66.00,
				Time:     "bad-time",
			},
		},
		"xdata": M{
			"time": "timestamp",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/bulk_create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkCreateTxnNotExists(t *testing.T) {
	txn := uuid.New().String()
	body, _ := sonic.Marshal(M{
		"data": []M{
			{
				"name": "admin",
				"key":  "*",
			},
			{
				"name": "staff",
				"key":  "xyz",
			},
		},
		"txn": txn,
	})
	w := ut.PerformRequest(engine, "POST", "/roles/bulk_create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

var orderIds []string
var orderMap map[string]Order

func TestBulkCreate(t *testing.T) {
	orders := make([]Order, 200)
	orderMap = make(map[string]Order)
	for i := 0; i < 200; i++ {
		err := faker.FakeData(&orders[i])
		assert.NoError(t, err)
		orders[i].TmpTime, err = time.Parse(`2006-01-02 15:04:05`, orders[i].Time)
		assert.NoError(t, err)
		orders[i].Time = orders[i].TmpTime.Format(time.RFC3339)
		orderMap[orders[i].No] = orders[i]
	}
	body, _ := sonic.Marshal(M{
		"data": orders,
		"xdata": M{
			"time": "timestamp",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/bulk_create",
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
	assert.Equal(t, 200, len(ids))
	oids := make([]primitive.ObjectID, 200)
	for i := 0; i < 200; i++ {
		orderIds = append(orderIds, ids[i].(string))
		oids[i], _ = primitive.ObjectIDFromHex(ids[i].(string))
	}

	cursor, err := service.Db.Collection("orders").
		Find(context.TODO(), bson.M{"_id": bson.M{"$in": oids}})
	assert.NoError(t, err)
	data := make([]M, 200)
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

func TestBulkCreateBadDbValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": []Order{
			{
				No:       "123456",
				Customer: "Joe",
				Phone:    "11225566",
				Cost:     66.00,
			},
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/bulk_create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

var projectIds []string

func TestBulkCreateEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	expire := make([]string, 10)
	data := make([]M, 10)
	for i := 0; i < 10; i++ {
		expire[i] = time.Now().Add(time.Hour * time.Duration(i)).Format(time.RFC3339)
		data[i] = M{
			"name":        fmt.Sprintf(`测试%d`, i),
			"namespace":   fmt.Sprintf(`test%d`, i),
			"secret":      fmt.Sprintf(`secret%d`, i),
			"expire_time": expire[i],
		}
	}

	body, _ := sonic.Marshal(M{
		"data": data,
		"xdata": M{
			"expire_time": "timestamp",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/projects/bulk_create",
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
	for i := 0; i < 10; i++ {
		projectIds = append(projectIds, ids[i].(string))
	}

	select {
	case msg := <-ch:
		assert.Equal(t, "bulk_create", msg.Action)
		assert.Equal(t, 10, len(msg.Data.([]interface{})))
		for i, v := range msg.Data.([]interface{}) {
			data := v.(M)
			assert.Equal(t, fmt.Sprintf(`测试%d`, i), data["name"])
			assert.Equal(t, fmt.Sprintf(`test%d`, i), data["namespace"])
			assert.Equal(t, fmt.Sprintf(`secret%d`, i), data["secret"])
			assert.Equal(t, expire[i], data["expire_time"])
		}
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestBulkCreateBadEvent(t *testing.T) {
	expire1 := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	expire2 := time.Now().Add(time.Hour * 36).Format(time.RFC3339)
	body, _ := sonic.Marshal(M{
		"data": []M{
			{
				"name":        "测试1",
				"namespace":   "test1",
				"secret":      "abcd",
				"expire_time": expire1,
			},
			{
				"name":        "测试2",
				"namespace":   "test2",
				"secret":      "zxcv",
				"expire_time": expire2,
			},
		},
		"xdata": M{
			"expire_time": "timestamp",
		},
	})
	RemoveStream(t)
	w := ut.PerformRequest(engine, "POST", "/projects/bulk_create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestSizeBadValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{})
	w := ut.PerformRequest(engine, "POST", "/orders/size",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestSizeBadTransform(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"_id": M{"$in": []string{"123456"}},
		},
		"xfilter": M{
			"_id.$in": "oids",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/size",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestSize(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/size",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Empty(t, resp.Body())
	assert.Equal(t, "200", resp.Header.Get("x-total"))
	assert.Equal(t, 204, resp.StatusCode())
}

func TestSizeWithFilterAndFormat(t *testing.T) {
	oids := orderIds[:5]
	body, _ := sonic.Marshal(M{
		"filter": M{
			"_id": M{"$in": oids},
		},
		"xfilter": M{
			"_id.$in": "oids",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/size",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Empty(t, resp.Body())
	assert.Equal(t, "5", resp.Header.Get("x-total"))
	assert.Equal(t, 204, resp.StatusCode())
}

func TestSizeBadFilter(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"abc": M{"$": "v"},
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/size",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

func TestFindBadValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{})
	w := ut.PerformRequest(engine, "POST", "/orders/find",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindBadTransform(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"_id": M{"$in": []string{"123456"}},
		},
		"xfilter": M{
			"_id.$in": "oids",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/find",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFind(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/find",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, "200", resp.Header.Get("x-total"))
	assert.Equal(t, 200, resp.StatusCode())
	var result []M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.Equal(t, 100, len(result))

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

func TestFindSort(t *testing.T) {
	u := url.URL{Path: "/orders/find"}
	query := u.Query()
	query.Add("sort", "cost:1")
	u.RawQuery = query.Encode()
	body, _ := sonic.Marshal(M{
		"filter": M{},
	})
	w := ut.PerformRequest(engine, "POST", u.RequestURI(),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, "200", resp.Header.Get("x-total"))
	assert.Equal(t, 200, resp.StatusCode())
	var result []M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.Equal(t, 100, len(result))

	for i := 0; i < 99; i++ {
		assert.LessOrEqual(t, result[i]["cost"], result[i+1]["cost"])
	}
}

func TestFindBadFilter(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"abc": M{"$": "v"},
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/find",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

func TestFindBadKeys(t *testing.T) {
	u := url.URL{Path: "/orders/find"}
	query := u.Query()
	query.Add("keys", `abc1`)
	u.RawQuery = query.Encode()
	w := ut.PerformRequest(engine, "POST", u.RequestURI(),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindOneBadValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{})
	w := ut.PerformRequest(engine, "POST", "/orders/find_one",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindOneBadTransform(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"_id": M{"$in": []string{"123456"}},
		},
		"xfilter": M{
			"_id.$in": "oids",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/find_one",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindOne(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"name": "weplanx",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/users/find_one",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	assert.Empty(t, result["password"])
	assert.Equal(t, "624a8facb4e5d150793d6353", result["department"])
	assert.ElementsMatch(t, roles, result["roles"])
	assert.NotEmpty(t, result["create_time"])
	assert.NotEmpty(t, result["update_time"])
}

func TestFindOneBadFilter(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"abc": M{
				"$": "v",
			},
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/find_one",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

func TestFindByIdBadValidate(t *testing.T) {
	u := url.URL{Path: fmt.Sprintf(`/users/%s`, userId)}
	query := u.Query()
	query.Set("keys", "$$$$")
	u.RawQuery = query.Encode()
	w := ut.PerformRequest(engine, "GET", u.RequestURI(),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindById(t *testing.T) {
	u := url.URL{Path: fmt.Sprintf(`/users/%s`, userId)}
	query := u.Query()
	query.Add("keys", "name")
	query.Add("keys", "password")
	query.Add("keys", "department")
	query.Add("keys", "roles")
	u.RawQuery = query.Encode()
	w := ut.PerformRequest(engine, "GET", u.RequestURI(),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	assert.Empty(t, result["password"])
	assert.Equal(t, "624a8facb4e5d150793d6353", result["department"])
	assert.ElementsMatch(t, roles, result["roles"])
	assert.Empty(t, result["create_time"])
	assert.Empty(t, result["update_time"])
}

func TestFindByIdBadKeys(t *testing.T) {
	u := url.URL{Path: fmt.Sprintf(`/users/%s`, userId)}
	query := u.Query()
	query.Add("keys", `abc1`)
	u.RawQuery = query.Encode()
	w := ut.PerformRequest(engine, "GET", u.RequestURI(),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateBadValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{})
	w := ut.PerformRequest(engine, "POST", "/users/update",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateBadFilterTransform(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"_id": M{"$in": []string{"123456"}},
		},
		"xfilter": M{
			"_id.$in": "oids",
		},
		"data": M{
			"$set": M{
				"department": "63579b8b9db7928aaebbe705",
			},
		},
		"xdata": M{
			"$set.department": "oid",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/users/update",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateBadDataTransform(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"name": "weplanx",
		},
		"data": M{
			"$set": M{
				"department": "123456",
			},
		},
		"xdata": M{
			"$set.department": "oid",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/users/update",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateTxnNotExists(t *testing.T) {
	txn := uuid.New().String()
	body, _ := sonic.Marshal(M{
		"filter": M{
			"key": "*",
		},
		"data": M{
			"$set": M{
				"name": "super",
			},
		},
		"txn": txn,
	})
	w := ut.PerformRequest(engine, "POST", "/roles/update",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdate(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"name": "weplanx",
		},
		"data": M{
			"$set": M{
				"department": "63579b8b9db7928aaebbe705",
			},
		},
		"xdata": M{
			"$set.department": "oid",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/users/update",
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
	err = service.Db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	assert.NoError(t, err)
	assert.Equal(t, "weplanx", data["name"])
	err = passlib.Verify("5auBnD$L", data["password"].(string))
	assert.NoError(t, err)
	departmentId, _ := primitive.ObjectIDFromHex("63579b8b9db7928aaebbe705")
	assert.Equal(t, departmentId, data["department"])
	roleIds := make([]primitive.ObjectID, len(roles))
	for k, v := range roles {
		roleIds[k], _ = primitive.ObjectIDFromHex(v)
	}
	assert.ElementsMatch(t, roleIds, data["roles"])
}

func TestUpdatePush(t *testing.T) {
	roles = append(roles, "62ce35710d94671a2e4a7d4c")
	body, _ := sonic.Marshal(M{
		"filter": M{
			"name": "weplanx",
		},
		"data": M{
			"$push": M{
				"roles": "62ce35710d94671a2e4a7d4c",
			},
		},
		"xdata": M{
			"$push.roles": "oid",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/users/update",
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
	err = service.Db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	assert.NoError(t, err)
	assert.Equal(t, "weplanx", data["name"])
	err = passlib.Verify("5auBnD$L", data["password"].(string))
	assert.NoError(t, err)
	departmentId, _ := primitive.ObjectIDFromHex("63579b8b9db7928aaebbe705")
	assert.Equal(t, departmentId, data["department"])
	roleIds := make([]primitive.ObjectID, len(roles))
	for k, v := range roles {
		roleIds[k], _ = primitive.ObjectIDFromHex(v)
	}
	assert.ElementsMatch(t, roleIds, data["roles"])
}

func TestUpdateBadDbValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"name": "weplanx",
		},
		"data": M{
			"$set": M{
				"department": "123456",
			},
		},
	})
	w := ut.PerformRequest(engine, "POST", "/users/update",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

func TestUpdateEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	expire := time.Now().Add(time.Hour * 72).Format(time.RFC3339)
	body, _ := sonic.Marshal(M{
		"filter": M{
			"namespace": "default",
		},
		"data": M{
			"$set": M{
				"secret":      "qwer",
				"expire_time": expire,
			},
		},
		"xdata": M{
			"$set.expire_time": "timestamp",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/projects/update",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Equal(t, "update", msg.Action)
		assert.Equal(t, "default", msg.Filter["namespace"])
		data := msg.Data.(M)["$set"].(M)
		assert.Equal(t, "qwer", data["secret"])
		assert.Equal(t, expire, data["expire_time"])
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestUpdateBadEvent(t *testing.T) {
	expire := time.Now().Add(time.Hour * 72).Format(time.RFC3339)
	body, _ := sonic.Marshal(M{
		"filter": M{
			"namespace": "default",
		},
		"data": M{
			"$set": M{
				"secret":      "qwer",
				"expire_time": expire,
			},
		},
		"xdata": M{
			"$set.expire_time": "timestamp",
		},
	})
	RemoveStream(t)
	w := ut.PerformRequest(engine, "POST", "/projects/update",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestUpdateByIdBadValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{})
	w := ut.PerformRequest(engine, "PATCH", fmt.Sprintf(`/users/%s`, userId),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateByIdBadDataTransform(t *testing.T) {
	u := url.URL{Path: fmt.Sprintf(`/users/%s`, userId)}
	body, _ := sonic.Marshal(M{
		"data": M{
			"$set": M{
				"department": "123456",
			},
		},
		"xdata": M{
			"$set.department": "oid",
		},
	})
	w := ut.PerformRequest(engine, "PATCH", u.RequestURI(),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateByIdTxnNotExists(t *testing.T) {
	txn := uuid.New().String()
	body, _ := sonic.Marshal(M{
		"data": M{
			"$set": M{
				"department": "62cbf9ac465f45091e981b1e",
			},
		},
		"xdata": M{
			"$set.department": "oid",
		},
		"txn": txn,
	})
	w := ut.PerformRequest(engine, "PATCH", fmt.Sprintf(`/users/%s`, userId),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateById(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"$set": M{
				"department": "62cbf9ac465f45091e981b1e",
			},
		},
		"xdata": M{
			"$set.department": "oid",
		},
	})
	w := ut.PerformRequest(engine, "PATCH", fmt.Sprintf(`/users/%s`, userId),
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
	err = service.Db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	assert.NoError(t, err)
	assert.Equal(t, "weplanx", data["name"])
	err = passlib.Verify("5auBnD$L", data["password"].(string))
	assert.NoError(t, err)
	departmentId, _ := primitive.ObjectIDFromHex("62cbf9ac465f45091e981b1e")
	assert.Equal(t, departmentId, data["department"])
	roleIds := make([]primitive.ObjectID, len(roles))
	for k, v := range roles {
		roleIds[k], _ = primitive.ObjectIDFromHex(v)
	}
	assert.ElementsMatch(t, roleIds, data["roles"])
}

func TestUpdateByIdPush(t *testing.T) {
	roles = append(roles, "62ce35b9b1d8fe7e38ef4c8c")
	body, _ := sonic.Marshal(M{
		"data": M{
			"$push": M{
				"roles": "62ce35b9b1d8fe7e38ef4c8c",
			},
		},
		"xdata": M{
			"$push.roles": "oid",
		},
	})
	w := ut.PerformRequest(engine, "PATCH", fmt.Sprintf(`/users/%s`, userId),
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
	err = service.Db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	assert.NoError(t, err)
	assert.Equal(t, "weplanx", data["name"])
	err = passlib.Verify("5auBnD$L", data["password"].(string))
	assert.NoError(t, err)
	departmentId, _ := primitive.ObjectIDFromHex("62cbf9ac465f45091e981b1e")
	assert.Equal(t, departmentId, data["department"])
	roleIds := make([]primitive.ObjectID, len(roles))
	for k, v := range roles {
		roleIds[k], _ = primitive.ObjectIDFromHex(v)
	}
	assert.ElementsMatch(t, roleIds, data["roles"])
}

func TestUpdateByIdBadDbValidate(t *testing.T) {
	u := url.URL{Path: fmt.Sprintf(`/users/%s`, userId)}
	body, _ := sonic.Marshal(M{
		"data": M{
			"$set": M{
				"department": "123456",
			},
		},
	})
	w := ut.PerformRequest(engine, "PATCH", u.RequestURI(),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

func TestUpdateByIdEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	expire := time.Now().Add(time.Hour * 12).Format(time.RFC3339)
	body, _ := sonic.Marshal(M{
		"data": M{
			"$set": M{
				"expire_time": expire,
			},
		},
		"xdata": M{
			"$set.expire_time": "timestamp",
		},
	})
	w := ut.PerformRequest(engine, "PATCH", fmt.Sprintf(`/projects/%s`, projectId),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Equal(t, "update_by_id", msg.Action)
		assert.Equal(t, projectId, msg.Id)
		assert.Empty(t, msg.Filter)
		data := msg.Data.(M)["$set"].(M)
		assert.Equal(t, expire, data["expire_time"])
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestUpdateByIdBadEvent(t *testing.T) {
	expire := time.Now().Add(time.Hour * 12).Format(time.RFC3339)
	body, _ := sonic.Marshal(M{
		"data": M{
			"$set": M{
				"expire_time": expire,
			},
		},
		"xdata": M{
			"$set.expire_time": "timestamp",
		},
	})
	RemoveStream(t)
	w := ut.PerformRequest(engine, "PATCH", fmt.Sprintf(`/projects/%s`, projectId),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestReplaceBadValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{})
	w := ut.PerformRequest(engine, "PUT", fmt.Sprintf(`/$$$$/%s`, userId),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestReplaceBadTransform(t *testing.T) {
	u := url.URL{Path: fmt.Sprintf(`/users/%s`, userId)}
	body, _ := sonic.Marshal(M{
		"data": M{
			"name":       "kain",
			"password":   "123456",
			"department": "123456",
			"roles":      []string{},
		},
		"xdata": M{
			"password":   "password",
			"department": "oid",
			"roles":      "oids",
		},
	})
	w := ut.PerformRequest(engine, "PUT", u.RequestURI(),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestReplaceTxnNotExists(t *testing.T) {
	txn := uuid.New().String()
	body, _ := sonic.Marshal(M{
		"data": M{
			"name":       "kain",
			"password":   "123456",
			"department": nil,
			"roles":      []string{},
		},
		"xdata": M{
			"password":   "password",
			"department": "oid",
			"roles":      "oids",
		},
		"txn": txn,
	})
	w := ut.PerformRequest(engine, "PUT", fmt.Sprintf(`/users/%s`, userId),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestReplace(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"name":       "kain",
			"password":   "123456",
			"department": nil,
			"roles":      []string{},
		},
		"xdata": M{
			"password":   "password",
			"department": "oid",
			"roles":      "oids",
		},
	})
	w := ut.PerformRequest(engine, "PUT", fmt.Sprintf(`/users/%s`, userId),
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
	err = service.Db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	assert.NoError(t, err)
	assert.Equal(t, "kain", data["name"])
	err = passlib.Verify("123456", data["password"].(string))
	assert.NoError(t, err)
	assert.Empty(t, data["department"])
	assert.Equal(t, primitive.A{}, data["roles"])
}

func TestReplaceBadDbValidate(t *testing.T) {
	u := url.URL{Path: fmt.Sprintf(`/users/%s`, userId)}
	body, _ := sonic.Marshal(M{
		"data": M{
			"name": "kain",
		},
	})
	w := ut.PerformRequest(engine, "PUT", u.RequestURI(),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

func TestReplaceEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	expire := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	body, _ := sonic.Marshal(M{
		"data": M{
			"name":        "工单项目",
			"namespace":   "orders",
			"secret":      "123456",
			"expire_time": expire,
		},
		"xdata": M{
			"expire_time": "timestamp",
		},
	})
	w := ut.PerformRequest(engine, "PUT", fmt.Sprintf(`/projects/%s`, projectId),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Equal(t, "replace", msg.Action)
		assert.Equal(t, projectId, msg.Id)
		data := msg.Data.(M)
		assert.Equal(t, "工单项目", data["name"])
		assert.Equal(t, "orders", data["namespace"])
		assert.Equal(t, "123456", data["secret"])
		assert.Equal(t, expire, data["expire_time"])
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestReplaceBadEvent(t *testing.T) {
	expire := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	body, _ := sonic.Marshal(M{
		"data": M{
			"name":        "工单项目",
			"namespace":   "orders",
			"secret":      "123456",
			"expire_time": expire,
		},
		"xdata": M{
			"expire_time": "timestamp",
		},
	})
	RemoveStream(t)
	w := ut.PerformRequest(engine, "PUT", fmt.Sprintf(`/projects/%s`, projectId),
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestDeleteBadValidate(t *testing.T) {
	w := ut.PerformRequest(engine, "DELETE", fmt.Sprintf(`/$$$$/%s`, userId),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestDeleteTxnNotExists(t *testing.T) {
	txn := uuid.New().String()
	w := ut.PerformRequest(engine, "DELETE", fmt.Sprintf(`/users/%s?txn=%s`, userId, txn),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestDelete(t *testing.T) {
	w := ut.PerformRequest(engine, "DELETE", fmt.Sprintf(`/users/%s`, userId),
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
	err = service.Db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
	assert.Error(t, err)
	assert.Equal(t, err, mongo.ErrNoDocuments)
}

func TestDeleteEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	w := ut.PerformRequest(engine, "DELETE", fmt.Sprintf(`/projects/%s`, projectId),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), result["DeletedCount"])

	select {
	case msg := <-ch:
		assert.Equal(t, "delete", msg.Action)
		assert.Equal(t, projectId, msg.Id)
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestDeleteBadEvent(t *testing.T) {
	RemoveStream(t)
	w := ut.PerformRequest(engine, "DELETE", fmt.Sprintf(`/projects/%s`, projectId),
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestBulkDeleteBadValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/bulk_delete",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkDeleteBadTransform(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"_id": M{"$in": []string{"12345"}},
		},
		"xfilter": M{
			"_id.$in": "oids",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/bulk_delete",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkDeleteTxnNotExists(t *testing.T) {
	txn := uuid.New()
	body, _ := sonic.Marshal(M{
		"filter": M{
			"key": "*",
		},
		"txn": txn,
	})
	w := ut.PerformRequest(engine, "POST", "/roles/bulk_delete",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkDelete(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"_id": M{"$in": orderIds[5:]},
		},
		"xfilter": M{
			"_id.$in": "oids",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/bulk_delete",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	var ids []primitive.ObjectID
	for _, v := range orderIds[5:] {
		id, _ := primitive.ObjectIDFromHex(v)
		ids = append(ids, id)
	}
	n, err := service.Db.Collection("orders").CountDocuments(context.TODO(), bson.M{"_id": bson.M{"$in": ids}})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestBulkDeleteBadFilter(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"abc": M{"$": "v"},
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/bulk_delete",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

func TestBulkDeleteEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	body, _ := sonic.Marshal(M{
		"filter": M{
			"namespace": M{"$in": []string{"test1", "test2"}},
		},
	})
	w := ut.PerformRequest(engine, "POST", "/projects/bulk_delete",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Equal(t, "bulk_delete", msg.Action)
		data := msg.Data.(M)
		assert.Equal(t, M{"$in": []interface{}{"test1", "test2"}}, data["namespace"])
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestBulkDeleteBadEvent(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"filter": M{
			"namespace": M{"$in": []string{"test1", "test2"}},
		},
	})
	RemoveStream(t)
	w := ut.PerformRequest(engine, "POST", "/projects/bulk_delete",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestSortBadValidate(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"key":    "sort",
			"values": []string{"12", "444"},
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/sort",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
}

func TestSortTxnNotExists(t *testing.T) {
	txn := uuid.New().String()
	sources := orderIds[:5]
	sources = funk.Reverse(sources).([]string)
	body, _ := sonic.Marshal(M{
		"data": M{
			"key":    "sort",
			"values": sources,
		},
		"txn": txn,
	})
	w := ut.PerformRequest(engine, "POST", "/orders/sort",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode())
}

func TestSort(t *testing.T) {
	sources := orderIds[:5]
	sources = funk.Reverse(sources).([]string)
	body, _ := sonic.Marshal(M{
		"data": M{
			"key":    "sort",
			"values": sources,
		},
	})
	w := ut.PerformRequest(engine, "POST", "/orders/sort",
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
	cursor, err := service.Db.Collection("orders").
		Find(context.TODO(), bson.M{"_id": bson.M{"$in": ids}})
	assert.NoError(t, err)
	var data []M
	err = cursor.All(context.TODO(), &data)
	assert.NoError(t, err)

	for _, v := range data {
		index := v["sort"].(int)
		assert.Equal(t, sources[index], v["_id"].(primitive.ObjectID).Hex())
	}
}

func TestSortEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	projectIds = funk.Reverse(projectIds).([]string)
	body, _ := sonic.Marshal(M{
		"data": M{
			"key":    "sort",
			"values": projectIds,
		},
	})
	w := ut.PerformRequest(engine, "POST", "/projects/sort",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 204, resp.StatusCode())
	assert.Empty(t, resp.Body())

	select {
	case msg := <-ch:
		assert.Equal(t, "sort", msg.Action)
		values := make([]string, 10)
		data := msg.Data.(M)
		assert.Equal(t, "sort", data["key"])
		for i, v := range data["values"].([]interface{}) {
			values[i] = v.(string)
		}
		assert.Equal(t, projectIds, values)
		t.Log(msg.Result)
		break
	}
}

func TestSortBadEvent(t *testing.T) {
	projectIds = funk.Reverse(projectIds).([]string)
	body, _ := sonic.Marshal(M{
		"data": M{
			"key":    "sort",
			"values": projectIds,
		},
	})
	RemoveStream(t)
	w := ut.PerformRequest(engine, "POST", "/projects/sort",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode())
	assert.Empty(t, resp.Body())
	RecoverStream(t)
}

func TestMoreTransform(t *testing.T) {
	body, _ := sonic.Marshal(M{
		"data": M{
			"name": "体验卡",
			"pd":   "2023-04-12T22:00:00.906Z",
			"valid": []string{
				"2023-04-12T22:00:00.906Z",
				"2023-04-13T06:30:05.586Z",
			},
			"metadata": []M{
				{
					"name": "aps",
					"date": "Fri, 14 Jul 2023 19:13:24 CST",
					"wm": []string{
						"Fri, 14 Jul 2023 19:13:24 CST",
						"Fri, 14 Jul 2023 20:14:10 CST",
					},
				},
				{
					"name": "kmx",
					"date": "Fri, 14 Jul 2023 21:15:05 CST",
					"wm": []string{
						"Fri, 14 Jul 2023 21:15:05 CST",
						"Fri, 14 Jul 2023 22:15:20 CST",
					},
				},
			},
		},
		"xdata": M{
			"pd":              "timestamp",
			"valid":           "timestamps",
			"metadata.$.date": "date",
			"metadata.$.wm":   "dates",
		},
	})
	w := ut.PerformRequest(engine, "POST", "/coupons/create",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp := w.Result()
	assert.Equal(t, 201, resp.StatusCode())
	var result M
	err := sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	id := result["InsertedID"].(string)
	oid, _ := primitive.ObjectIDFromHex(id)

	var coupon M
	err = service.Db.Collection("coupons").FindOne(context.TODO(), bson.M{
		"_id": oid,
	}).Decode(&coupon)
	assert.NoError(t, err)

	assert.Equal(t, "体验卡", coupon["name"])
	assert.Equal(t, primitive.DateTime(1681336800906), coupon["pd"])
	assert.Equal(t,
		primitive.A{primitive.DateTime(1681336800906), primitive.DateTime(1681367405586)},
		coupon["valid"],
	)
	metadata := coupon["metadata"].(primitive.A)
	assert.ElementsMatch(t, primitive.A{
		M{
			"name": "aps",
			"date": primitive.DateTime(1689333204000),
			"wm": primitive.A{
				primitive.DateTime(1689333204000),
				primitive.DateTime(1689336850000),
			},
		},
		M{
			"name": "kmx",
			"date": primitive.DateTime(1689340505000),
			"wm": primitive.A{
				primitive.DateTime(1689340505000),
				primitive.DateTime(1689344120000),
			},
		},
	}, metadata)
}

type TransactionFn = func(txn string)

func Transaction(t *testing.T, fn TransactionFn) {
	w1 := ut.PerformRequest(engine, "POST", "/transaction",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp1 := w1.Result()
	assert.Equal(t, 201, resp1.StatusCode())
	var result1 M
	err := sonic.Unmarshal(resp1.Body(), &result1)
	assert.NoError(t, err)
	txn := result1["txn"].(string)

	fn(txn)

	body, _ := sonic.Marshal(M{
		"txn": txn,
	})
	w2 := ut.PerformRequest(engine, "POST", "/commit",
		&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp2 := w2.Result()
	assert.Equal(t, 200, resp2.StatusCode())
}

func TestTxBulkCreate(t *testing.T) {
	ctx := context.TODO()
	err := service.Db.Collection("x_test").Drop(ctx)
	assert.NoError(t, err)

	Transaction(t, func(txn string) {
		body, _ := sonic.Marshal(M{
			"data": []M{
				{"name": "abc"},
				{"name": "xxx"},
			},
			"txn": txn,
		})
		w := ut.PerformRequest(engine, "POST", "/x_test/bulk_create",
			&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
			ut.Header{Key: "content-type", Value: "application/json"},
		)
		resp := w.Result()
		assert.Equal(t, 204, resp.StatusCode())

		count, err := service.Db.Collection("x_test").CountDocuments(ctx, bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	count, err := service.Db.Collection("x_test").CountDocuments(ctx, bson.M{})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestTxUpdate(t *testing.T) {
	ctx := context.TODO()
	err := service.Db.Collection("x_test").Drop(ctx)
	assert.NoError(t, err)

	r, err := service.Db.Collection("x_test").InsertOne(ctx, bson.M{
		"name": "kain",
	})
	assert.NoError(t, err)
	id := r.InsertedID

	Transaction(t, func(txn string) {
		body, _ := sonic.Marshal(M{
			"filter": M{"name": "kain"},
			"data": M{
				"$set": M{"name": "xxxx"},
			},
			"txn": txn,
		})
		w := ut.PerformRequest(engine, "POST", "/x_test/update",
			&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
			ut.Header{Key: "content-type", Value: "application/json"},
		)
		resp := w.Result()
		assert.Equal(t, 204, resp.StatusCode())

		var result M
		err := service.Db.Collection("x_test").FindOne(ctx, bson.M{
			"_id": id,
		}).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, "kain", result["name"])
	})

	var result M
	err = service.Db.Collection("x_test").FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "xxxx", result["name"])
}

func TestTxUpdateById(t *testing.T) {
	ctx := context.TODO()
	err := service.Db.Collection("x_test").Drop(ctx)
	assert.NoError(t, err)

	r, err := service.Db.Collection("x_test").InsertOne(ctx, bson.M{
		"name": "kain",
	})
	assert.NoError(t, err)
	id := r.InsertedID.(primitive.ObjectID)

	Transaction(t, func(txn string) {
		body, _ := sonic.Marshal(M{
			"data": M{
				"$set": M{"name": "xxxx"},
			},
			"txn": txn,
		})
		w := ut.PerformRequest(engine, "PATCH", fmt.Sprintf(`/x_test/%s`, id.Hex()),
			&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
			ut.Header{Key: "content-type", Value: "application/json"},
		)
		resp := w.Result()
		assert.Equal(t, 204, resp.StatusCode())

		var result M
		err := service.Db.Collection("x_test").FindOne(ctx, bson.M{
			"_id": id,
		}).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, "kain", result["name"])
	})

	var result M
	err = service.Db.Collection("x_test").FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "xxxx", result["name"])
}

func TestTxReplace(t *testing.T) {
	ctx := context.TODO()
	err := service.Db.Collection("x_test").Drop(ctx)
	assert.NoError(t, err)

	r, err := service.Db.Collection("x_test").InsertOne(ctx, bson.M{
		"name": "kain",
	})
	assert.NoError(t, err)
	id := r.InsertedID.(primitive.ObjectID)

	Transaction(t, func(txn string) {
		body, _ := sonic.Marshal(M{
			"data": M{
				"name": "xxxx",
			},
			"txn": txn,
		})
		w := ut.PerformRequest(engine, "PUT", fmt.Sprintf(`/x_test/%s`, id.Hex()),
			&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
			ut.Header{Key: "content-type", Value: "application/json"},
		)
		resp := w.Result()
		assert.Equal(t, 204, resp.StatusCode())

		var result M
		err := service.Db.Collection("x_test").FindOne(ctx, bson.M{
			"_id": id,
		}).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, "kain", result["name"])
	})

	var result M
	err = service.Db.Collection("x_test").FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "xxxx", result["name"])
}

func TestTxDelete(t *testing.T) {
	ctx := context.TODO()
	err := service.Db.Collection("x_test").Drop(ctx)
	assert.NoError(t, err)

	r, err := service.Db.Collection("x_test").InsertOne(ctx, bson.M{
		"name": "kain",
	})
	assert.NoError(t, err)
	id := r.InsertedID.(primitive.ObjectID)

	Transaction(t, func(txn string) {
		u := url.URL{Path: fmt.Sprintf(`/x_test/%s`, id.Hex())}
		query := u.Query()
		query.Add("txn", txn)
		u.RawQuery = query.Encode()
		w := ut.PerformRequest(engine, "DELETE", u.RequestURI(),
			&ut.Body{},
			ut.Header{Key: "content-type", Value: "application/json"},
		)
		resp := w.Result()
		assert.Equal(t, 204, resp.StatusCode())

		count, err := service.Db.Collection("x_test").CountDocuments(ctx, bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	count, err := service.Db.Collection("x_test").CountDocuments(ctx, bson.M{})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestTxBulkDelete(t *testing.T) {
	ctx := context.TODO()
	err := service.Db.Collection("x_test").Drop(ctx)
	assert.NoError(t, err)

	_, err = service.Db.Collection("x_test").InsertOne(ctx, bson.M{
		"name": "kain",
	})
	assert.NoError(t, err)

	Transaction(t, func(txn string) {
		body, _ := sonic.Marshal(M{
			"filter": M{
				"name": "kain",
			},
			"txn": txn,
		})
		w := ut.PerformRequest(engine, "POST", "/x_test/bulk_delete",
			&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
			ut.Header{Key: "content-type", Value: "application/json"},
		)
		resp := w.Result()
		assert.Equal(t, 204, resp.StatusCode())

		count, err := service.Db.Collection("x_test").CountDocuments(ctx, bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	count, err := service.Db.Collection("x_test").CountDocuments(ctx, bson.M{})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestTxSort(t *testing.T) {
	ctx := context.TODO()
	err := service.Db.Collection("x_test").Drop(ctx)
	assert.NoError(t, err)

	r, err := service.Db.Collection("x_test").InsertMany(ctx, []interface{}{
		bson.M{"name": "kain"},
		bson.M{"name": "xxxx"},
	})
	assert.NoError(t, err)
	var sources []string
	for _, v := range r.InsertedIDs {
		sources = append(sources, v.(primitive.ObjectID).Hex())
	}

	Transaction(t, func(txn string) {
		body, _ := sonic.Marshal(M{
			"data": M{
				"key":    "sort",
				"values": sources,
			},
			"txn": txn,
		})
		w := ut.PerformRequest(engine, "POST", "/x_test/sort",
			&ut.Body{Body: bytes.NewBuffer(body), Len: len(body)},
			ut.Header{Key: "content-type", Value: "application/json"},
		)
		resp := w.Result()
		assert.Equal(t, 204, resp.StatusCode())

		cursor, err := service.Db.Collection("x_test").Find(ctx, bson.M{})
		assert.NoError(t, err)
		var result []M
		err = cursor.All(ctx, &result)
		assert.NoError(t, err)
		t.Log(result)
	})

	cursor, err := service.Db.Collection("x_test").Find(ctx, bson.M{})
	assert.NoError(t, err)
	var result []M
	err = cursor.All(ctx, &result)
	assert.NoError(t, err)
	t.Log(result)
}

func TestTransaction(t *testing.T) {
	ctx := context.TODO()
	err := service.Db.Collection("x_users").Drop(ctx)
	assert.NoError(t, err)
	err = service.Db.Collection("x_roles").Drop(ctx)
	assert.NoError(t, err)

	Transaction(t, func(txn string) {
		body1, _ := sonic.Marshal(M{
			"data": M{
				"name": "admin",
				"key":  "*",
			},
			"txn": txn,
		})
		w1 := ut.PerformRequest(engine, "POST", "/x_roles/create",
			&ut.Body{Body: bytes.NewBuffer(body1), Len: len(body1)},
			ut.Header{Key: "content-type", Value: "application/json"},
		)
		resp1 := w1.Result()
		assert.Equal(t, 204, resp1.StatusCode())

		body2, _ := sonic.Marshal(M{
			"data": M{
				"name":  "kainxxxx",
				"roles": []string{"*"},
			},
			"txn": txn,
		})
		w2 := ut.PerformRequest(engine, "POST", "/x_users/create",
			&ut.Body{Body: bytes.NewBuffer(body2), Len: len(body2)},
			ut.Header{Key: "content-type", Value: "application/json"},
		)
		resp2 := w2.Result()
		assert.Equal(t, 204, resp2.StatusCode())
	})

	var user M
	err = service.Db.Collection("x_users").FindOne(ctx, bson.M{
		"name": "kainxxxx",
	}).Decode(&user)
	assert.NoError(t, err)
	assert.Equal(t, primitive.A{"*"}, user["roles"])

	var role M
	err = service.Db.Collection("x_roles").FindOne(ctx, bson.M{
		"key": "*",
	}).Decode(&role)
	assert.NoError(t, err)
	assert.Equal(t, "admin", role["name"])
}

func TestCommitNotTxn(t *testing.T) {
	w1 := ut.PerformRequest(engine, "POST", "/commit",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp1 := w1.Result()
	assert.Equal(t, 400, resp1.StatusCode())

	body2, _ := sonic.Marshal(M{
		"txn": uuid.New().String(),
	})
	w2 := ut.PerformRequest(engine, "POST", "/commit",
		&ut.Body{Body: bytes.NewBuffer(body2), Len: len(body2)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp2 := w2.Result()
	assert.Equal(t, 400, resp2.StatusCode())
}

func TestCommitTimeout(t *testing.T) {
	service.Values.RestTxnTimeout = time.Second
	w1 := ut.PerformRequest(engine, "POST", "/transaction",
		&ut.Body{},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp1 := w1.Result()
	assert.Equal(t, 201, resp1.StatusCode())
	var result1 M
	err := sonic.Unmarshal(resp1.Body(), &result1)
	assert.NoError(t, err)
	txn := result1["txn"].(string)

	time.Sleep(time.Second)

	body2, _ := sonic.Marshal(M{
		"txn": txn,
	})
	w2 := ut.PerformRequest(engine, "POST", "/commit",
		&ut.Body{Body: bytes.NewBuffer(body2), Len: len(body2)},
		ut.Header{Key: "content-type", Value: "application/json"},
	)
	resp2 := w2.Result()
	assert.Equal(t, 400, resp2.StatusCode())
	t.Log(string(resp2.Body()))
}
