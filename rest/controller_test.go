package rest_test

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/help"
	"github.com/weplanx/go/passlib"
	"github.com/weplanx/go/rest"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"
)

var roles = []string{"635797539db7928aaebbe6e5", "635797c19db7928aaebbe6e6"}
var userId string

func TestCreateValidateBad(t *testing.T) {
	resp, err := Req("POST", "/users/create", M{})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestCreateValidateBad2(t *testing.T) {
	resp, err := Req("POST", "/ /create", M{
		"data": M{},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
	t.Log(string(resp.Body()))
}

func TestCreateForbid(t *testing.T) {
	resp, err := Req("POST", "/permissions/create", M{
		"data": M{
			"name": "kain",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestCreateForbidStatusFalse(t *testing.T) {
	resp, err := Req("POST", "/levels/create", M{
		"data": M{
			"name": "level1",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestCreateTransformBad(t *testing.T) {
	resp, err := Req("POST", "/users/create", M{
		"data": M{
			"department": "123",
		},
		"xdata": M{
			"department": "oid",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestCreateTxnNotExists(t *testing.T) {
	txn := help.Uuid()
	resp, err := Req("POST", "/users/create", M{
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
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestCreate(t *testing.T) {
	resp, err := Req("POST", "/users/create", M{
		"data": M{
			"name":       "weplanx",
			"password":   "5auBnD$L",
			"department": "624a8facb4e5d150793d6353",
			"roles":      roles,
			"phone":      "123456789",
		},
		"xdata": M{
			"password":   "password",
			"department": "oid",
			"roles":      "oids",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	userId = result["InsertedID"].(string)
	id, _ := primitive.ObjectIDFromHex(userId)
	var data M
	err = service.Db.Collection("users").
		FindOne(context.TODO(), bson.M{"_id": id}).
		Decode(&data)
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

func TestCreateDbJSONSchemaBad(t *testing.T) {
	resp, err := Req("POST", "/users/create", M{
		"data": M{
			"name": "weplanx",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

var projectId string

func TestCreateEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	expire := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	resp, err := Req("POST", "/projects/create", M{
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
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	projectId = result["InsertedID"].(string)

	select {
	case msg := <-ch:
		assert.Equal(t, rest.ActionCreate, msg.Action)
		data := msg.Data.(M)
		assert.Equal(t, "默认项目", data["name"])
		assert.Equal(t, "default", data["namespace"])
		assert.Equal(t, "abcd", data["secret"])
		assert.Equal(t, expire, data["expire_time"])
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestCreateEventBad(t *testing.T) {
	RemoveStream(t)
	expire := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	resp, err := Req("POST", "/projects/create", M{
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
	assert.NoError(t, err)
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

func TestBulkCreateValidateBad(t *testing.T) {
	resp, err := Req("POST", "/orders/bulk_create", M{})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkCreateForbid(t *testing.T) {
	resp, err := Req("POST", "/permissions/bulk_create", M{
		"data": []M{
			{"name": "kain"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkCreateTransformBad(t *testing.T) {
	resp, err := Req("POST", "/orders/bulk_create", M{
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
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkCreateTxnNotExists(t *testing.T) {
	txn := help.Uuid()
	resp, err := Req("POST", "/roles/bulk_create", M{
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
	assert.NoError(t, err)
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
	resp, err := Req("POST", "/orders/bulk_create", M{
		"data": orders,
		"xdata": M{
			"time": "timestamp",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
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

func TestBulkCreateDbJSONSchemaBad(t *testing.T) {
	resp, err := Req("POST", "/orders/bulk_create", M{
		"data": []Order{
			{
				No:       "123456",
				Customer: "Joe",
				Phone:    "11225566",
				Cost:     66.00,
			},
		},
	})
	assert.NoError(t, err)
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

	resp, err := Req("POST", "/projects/bulk_create", M{
		"data": data,
		"xdata": M{
			"expire_time": "timestamp",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	ids := result["InsertedIDs"].([]interface{})
	for i := 0; i < 10; i++ {
		projectIds = append(projectIds, ids[i].(string))
	}

	select {
	case msg := <-ch:
		assert.Equal(t, rest.ActionBulkCreate, msg.Action)
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

func TestBulkCreateEventBad(t *testing.T) {
	RemoveStream(t)
	expire1 := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	expire2 := time.Now().Add(time.Hour * 36).Format(time.RFC3339)
	resp, err := Req("POST", "/projects/bulk_create", M{
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
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestSizeValidateBad(t *testing.T) {
	resp, err := Req("POST", "/orders/size", M{})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestSizeForbid(t *testing.T) {
	resp, err := Req("POST", "/permissions/size", M{
		"filter": M{},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestSizeTransformBad(t *testing.T) {
	resp, err := Req("POST", "/orders/size", M{
		"filter": M{
			"_id": M{"$in": []string{"123456"}},
		},
		"xfilter": M{
			"_id->$in": "oids",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestSize(t *testing.T) {
	resp, err := Req("POST", "/orders/size", M{
		"filter": M{},
	})
	assert.NoError(t, err)
	assert.Empty(t, resp.Body())
	assert.Equal(t, "200", resp.Header.Get("x-total"))
	assert.Equal(t, 204, resp.StatusCode())
}

func TestSizeFilterAndXfilter(t *testing.T) {
	oids := orderIds[:5]
	resp, err := Req("POST", "/orders/size", M{
		"filter": M{
			"_id": M{"$in": oids},
		},
		"xfilter": M{
			"_id->$in": "oids",
		},
	})
	assert.NoError(t, err)
	assert.Empty(t, resp.Body())
	assert.Equal(t, "5", resp.Header.Get("x-total"))
	assert.Equal(t, 204, resp.StatusCode())
}

func TestSizeFilterBad(t *testing.T) {
	resp, err := Req("POST", "/orders/size", M{
		"filter": M{
			"abc": M{"$": "v"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

func TestFindValidateBad(t *testing.T) {
	resp, err := Req("POST", "/orders/find", M{})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindForbid(t *testing.T) {
	resp, err := Req("POST", "/permissions/find", M{
		"filter": M{},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindTransformBad(t *testing.T) {
	resp, err := Req("POST", "/orders/find", M{
		"filter": M{
			"_id": M{"$in": []string{"123456"}},
		},
		"xfilter": M{
			"_id->$in": "oids",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFind(t *testing.T) {
	resp, err := Req("POST", "/orders/find", M{
		"filter": M{},
	})
	assert.NoError(t, err)
	assert.Equal(t, "200", resp.Header.Get("x-total"))
	assert.Equal(t, 200, resp.StatusCode())

	var result []M
	err = sonic.Unmarshal(resp.Body(), &result)
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

func TestFindWithSensitives(t *testing.T) {
	resp, err := Req("POST", "/users/find", M{
		"filter": M{},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result []M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	for _, v := range result {
		assert.Equal(t, "*", v["phone"])
	}
}

func TestFindSort(t *testing.T) {
	u := Url("/orders/find", Params{
		{"sort", "cost:1"},
	})
	resp, err := Req("POST", u, M{
		"filter": M{},
	})
	assert.NoError(t, err)
	assert.Equal(t, "200", resp.Header.Get("x-total"))
	assert.Equal(t, 200, resp.StatusCode())

	var result []M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.Equal(t, 100, len(result))

	for i := 0; i < 99; i++ {
		assert.LessOrEqual(t, result[i]["cost"], result[i+1]["cost"])
	}
}

func TestFindFilterBad(t *testing.T) {
	resp, err := Req("POST", "/orders/find", M{
		"filter": M{
			"abc": M{"$": "v"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

func TestFindKeysBad(t *testing.T) {
	u := Url("/orders/find", Params{
		{"keys", "abc1"},
	})
	resp, err := Req("POST", u, nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindOneValidateBad(t *testing.T) {
	resp, err := Req("POST", "/orders/find_one", M{})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindOneForbid(t *testing.T) {
	resp, err := Req("POST", "/permissions/find_one", M{
		"filter": M{
			"name": "kain",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindOneTransformBad(t *testing.T) {
	resp, err := Req("POST", "/orders/find_one", M{
		"filter": M{
			"_id": M{"$in": []string{"123456"}},
		},
		"xfilter": M{
			"_id->$in": "oids",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindOne(t *testing.T) {
	resp, err := Req("POST", "/users/find_one", M{
		"filter": M{
			"name": "weplanx",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	assert.Empty(t, result["password"])
	assert.Equal(t, "624a8facb4e5d150793d6353", result["department"])
	assert.ElementsMatch(t, roles, result["roles"])
	assert.NotEmpty(t, result["create_time"])
	assert.NotEmpty(t, result["update_time"])
}

func TestFindOneFilterBad(t *testing.T) {
	resp, err := Req("POST", "/orders/find_one", M{
		"filter": M{
			"abc": M{
				"$": "v",
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

func TestFindByIdValidateBad(t *testing.T) {
	u := Url(fmt.Sprintf(`/users/%s`, "123"), Params{
		{"keys", "$$$$"},
	})
	resp, err := Req("GET", u, nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindByIdForbid(t *testing.T) {
	id := primitive.NewObjectID().Hex()
	resp, err := Req("GET", fmt.Sprintf(`/permissions/%s`, id), nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestFindByIdNotExists(t *testing.T) {
	u := Url(fmt.Sprintf(`/users/%s`, primitive.NewObjectID().Hex()), Params{})
	resp, err := Req("GET", u, nil)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

func TestFindById(t *testing.T) {
	u := Url(fmt.Sprintf(`/users/%s`, userId), Params{
		{"keys", "name"},
		{"keys", "password"},
		{"keys", "department"},
		{"keys", "roles"},
	})
	resp, err := Req("GET", u, nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())
	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	assert.Empty(t, result["password"])
	assert.Equal(t, "624a8facb4e5d150793d6353", result["department"])
	assert.ElementsMatch(t, roles, result["roles"])
	assert.Empty(t, result["create_time"])
	assert.Empty(t, result["update_time"])
}

func TestUpdateValidateBad(t *testing.T) {
	resp, err := Req("POST", "/users/update", M{})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateForbid(t *testing.T) {
	resp, err := Req("POST", "/permissions/update", M{
		"filter": M{
			"name": "kain",
		},
		"data": M{
			"$set": M{
				"name": "xxxx",
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateTransformFilterBad(t *testing.T) {
	resp, err := Req("POST", "/users/update", M{
		"filter": M{
			"_id": M{"$in": []string{"123456"}},
		},
		"xfilter": M{
			"_id->$in": "oids",
		},
		"data": M{
			"$set": M{
				"department": "63579b8b9db7928aaebbe705",
			},
		},
		"xdata": M{
			"$set->department": "oid",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateTransformDataBad(t *testing.T) {
	resp, err := Req("POST", "/users/update", M{
		"filter": M{
			"name": "weplanx",
		},
		"data": M{
			"$set": M{
				"department": "123456",
			},
		},
		"xdata": M{
			"$set->department": "oid",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateTxnNotExists(t *testing.T) {
	txn := help.Uuid()
	resp, err := Req("POST", "/roles/update", M{
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
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdate(t *testing.T) {
	resp, err := Req("POST", "/users/update", M{
		"filter": M{
			"name": "weplanx",
		},
		"data": M{
			"$set": M{
				"department": "63579b8b9db7928aaebbe705",
			},
		},
		"xdata": M{
			"$set->department": "oid",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
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
	resp, err := Req("POST", "/users/update", M{
		"filter": M{
			"name": "weplanx",
		},
		"data": M{
			"$push": M{
				"roles": "62ce35710d94671a2e4a7d4c",
			},
		},
		"xdata": M{
			"$push->roles": "oid",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
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

func TestUpdateDbJSONSchemaBad(t *testing.T) {
	resp, err := Req("POST", "/users/update", M{
		"filter": M{
			"name": "weplanx",
		},
		"data": M{
			"$set": M{
				"department": "123456",
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

func TestUpdateEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	expire := time.Now().Add(time.Hour * 72).Format(time.RFC3339)
	resp, err := Req("POST", "/projects/update", M{
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
			"$set->expire_time": "timestamp",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Equal(t, rest.ActionUpdate, msg.Action)
		assert.Equal(t, "default", msg.Filter["namespace"])
		data := msg.Data.(M)["$set"].(M)
		assert.Equal(t, "qwer", data["secret"])
		assert.Equal(t, expire, data["expire_time"])
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestUpdateEventBad(t *testing.T) {
	RemoveStream(t)
	expire := time.Now().Add(time.Hour * 72).Format(time.RFC3339)
	resp, err := Req("POST", "/projects/update", M{
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
			"$set->expire_time": "timestamp",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestUpdateByIdValidateBad(t *testing.T) {
	resp, err := Req("PATCH", fmt.Sprintf(`/users/%s`, userId), M{})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateByIdForbid(t *testing.T) {
	id := primitive.NewObjectID().Hex()
	resp, err := Req("PATCH", fmt.Sprintf(`/permissions/%s`, id), M{
		"data": M{
			"$set": M{
				"name": "xxxx",
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateByIdTransformDataBad(t *testing.T) {
	resp, err := Req("PATCH", fmt.Sprintf(`/users/%s`, userId), M{
		"data": M{
			"$set": M{
				"department": "123456",
			},
		},
		"xdata": M{
			"$set->department": "oid",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateByIdTxnNotExists(t *testing.T) {
	txn := help.Uuid()
	resp, err := Req("PATCH", fmt.Sprintf(`/users/%s`, userId), M{
		"data": M{
			"$set": M{
				"department": "62cbf9ac465f45091e981b1e",
			},
		},
		"xdata": M{
			"$set->department": "oid",
		},
		"txn": txn,
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestUpdateById(t *testing.T) {
	resp, err := Req("PATCH", fmt.Sprintf(`/users/%s`, userId), M{
		"data": M{
			"$set": M{
				"department": "62cbf9ac465f45091e981b1e",
			},
		},
		"xdata": M{
			"$set->department": "oid",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
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
	resp, err := Req("PATCH", fmt.Sprintf(`/users/%s`, userId), M{
		"data": M{
			"$push": M{
				"roles": "62ce35b9b1d8fe7e38ef4c8c",
			},
		},
		"xdata": M{
			"$push->roles": "oid",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
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

func TestUpdateByIdDbJSONSchemaBad(t *testing.T) {
	resp, err := Req("PATCH", fmt.Sprintf(`/users/%s`, userId), M{
		"data": M{
			"$set": M{
				"department": "123456",
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

func TestUpdateByIdEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	expire := time.Now().Add(time.Hour * 12).Format(time.RFC3339)
	resp, err := Req("PATCH", fmt.Sprintf(`/projects/%s`, projectId), M{
		"data": M{
			"$set": M{
				"expire_time": expire,
			},
		},
		"xdata": M{
			"$set->expire_time": "timestamp",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Equal(t, rest.ActionUpdateById, msg.Action)
		assert.Equal(t, projectId, msg.Id)
		assert.Empty(t, msg.Filter)
		data := msg.Data.(M)["$set"].(M)
		assert.Equal(t, expire, data["expire_time"])
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestUpdateByIdEventBad(t *testing.T) {
	RemoveStream(t)
	expire := time.Now().Add(time.Hour * 12).Format(time.RFC3339)
	resp, err := Req("PATCH", fmt.Sprintf(`/projects/%s`, projectId), M{
		"data": M{
			"$set": M{
				"expire_time": expire,
			},
		},
		"xdata": M{
			"$set->expire_time": "timestamp",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestReplaceValidateBad(t *testing.T) {
	resp, err := Req("PUT", fmt.Sprintf(`/$$$$/%s`, "123"), M{})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestReplaceForbid(t *testing.T) {
	id := primitive.NewObjectID().Hex()
	resp, err := Req("PUT", fmt.Sprintf(`/permissions/%s`, id), M{
		"data": M{
			"name": "xxxx",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestReplaceTransformBad(t *testing.T) {
	resp, err := Req("PUT", fmt.Sprintf(`/users/%s`, userId), M{
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
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestReplaceTxnNotExists(t *testing.T) {
	txn := help.Uuid()
	resp, err := Req("PUT", fmt.Sprintf(`/users/%s`, userId), M{
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
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestReplace(t *testing.T) {
	resp, err := Req("PUT", fmt.Sprintf(`/users/%s`, userId), M{
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
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
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

func TestReplaceDbJSONSchemaBad(t *testing.T) {
	resp, err := Req("PUT", fmt.Sprintf(`/users/%s`, userId), M{
		"data": M{
			"name": "kain",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

func TestReplaceEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	expire := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	resp, err := Req("PUT", fmt.Sprintf(`/projects/%s`, projectId), M{
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
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Equal(t, rest.ActionReplace, msg.Action)
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

func TestReplaceEventBad(t *testing.T) {
	RemoveStream(t)
	expire := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	resp, err := Req("PUT", fmt.Sprintf(`/projects/%s`, projectId), M{
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
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestDeleteValidateBad(t *testing.T) {
	resp, err := Req("DELETE", fmt.Sprintf(`/$$$$/%s`, userId), nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestDeleteForbid(t *testing.T) {
	id := primitive.NewObjectID().Hex()
	resp, err := Req("DELETE", fmt.Sprintf(`/permissions/%s`, id), nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestDeleteTxnNotExists(t *testing.T) {
	txn := help.Uuid()
	resp, err := Req("DELETE", fmt.Sprintf(`/users/%s?txn=%s`, userId, txn), nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestDelete(t *testing.T) {
	resp, err := Req("DELETE", fmt.Sprintf(`/users/%s`, userId), nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
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

	resp, err := Req("DELETE", fmt.Sprintf(`/projects/%s`, projectId), nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), result["DeletedCount"])

	select {
	case msg := <-ch:
		assert.Equal(t, rest.ActionDelete, msg.Action)
		assert.Equal(t, projectId, msg.Id)
		t.Log(msg.Data)
		assert.Equal(t, projectId, msg.Data.(M)["_id"])
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestDeleteEventBad(t *testing.T) {
	RemoveStream(t)
	resp, err := Req("DELETE", fmt.Sprintf(`/projects/%s`, projectId), nil)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestBulkDeleteValidateBad(t *testing.T) {
	resp, err := Req("POST", "/orders/bulk_delete", M{
		"filter": M{},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkDeleteForbid(t *testing.T) {
	resp, err := Req("POST", "/permissions/bulk_delete", M{
		"filter": M{
			"name": "kain",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkDeleteTransformBad(t *testing.T) {
	resp, err := Req("POST", "/orders/bulk_delete", M{
		"filter": M{
			"_id": M{"$in": []string{"12345"}},
		},
		"xfilter": M{
			"_id->$in": "oids",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkDeleteTxnNotExists(t *testing.T) {
	txn := help.Uuid()
	resp, err := Req("POST", "/roles/bulk_delete", M{
		"filter": M{
			"key": "*",
		},
		"txn": txn,
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestBulkDelete(t *testing.T) {
	resp, err := Req("POST", "/orders/bulk_delete", M{
		"filter": M{
			"_id": M{"$in": orderIds[5:]},
		},
		"xfilter": M{
			"_id->$in": "oids",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
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

func TestBulkDeleteFilterBad(t *testing.T) {
	resp, err := Req("POST", "/orders/bulk_delete", M{
		"filter": M{
			"abc": M{"$": "v"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

func TestBulkDeleteEvent(t *testing.T) {
	ch := make(chan rest.PublishDto)
	go MockSubscribe(t, ch)

	resp, err := Req("POST", "/projects/bulk_delete", M{
		"filter": M{
			"namespace": M{"$in": []string{"test8", "test9"}},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Equal(t, rest.ActionBulkDelete, msg.Action)
		assert.Equal(t, M{"$in": []interface{}{"test8", "test9"}}, msg.Filter["namespace"])
		assert.Equal(t, 2, len(msg.Data.([]interface{})))
		assert.Equal(t, result, msg.Result)
		break
	}
}

func TestBulkDeleteEventBad(t *testing.T) {
	RemoveStream(t)
	resp, err := Req("POST", "/projects/bulk_delete", M{
		"filter": M{
			"namespace": M{"$in": []string{"test1", "test2"}},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

func TestSortValidateBad(t *testing.T) {
	resp, err := Req("POST", "/orders/sort", M{
		"data": M{
			"key":    "sort",
			"values": []string{"12", "444"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
}

func TestSortForbid(t *testing.T) {
	resp, err := Req("POST", "/permissions/sort", M{
		"data": M{
			"key":    "sort",
			"values": []string{primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex()},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestSortTxnNotExists(t *testing.T) {
	txn := help.Uuid()
	sources := orderIds[:5]
	help.Reverse[string](sources)
	resp, err := Req("POST", "/orders/sort", M{
		"data": M{
			"key":    "sort",
			"values": sources,
		},
		"txn": txn,
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode())
}

func TestSort(t *testing.T) {
	sources := orderIds[:5]
	help.Reverse[string](sources)
	resp, err := Req("POST", "/orders/sort", M{
		"data": M{
			"key":    "sort",
			"values": sources,
		},
	})
	assert.NoError(t, err)
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

	help.Reverse[string](projectIds)
	resp, err := Req("POST", "/projects/sort", M{
		"data": M{
			"key":    "sort",
			"values": projectIds,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode())
	assert.Empty(t, resp.Body())

	select {
	case msg := <-ch:
		assert.Equal(t, rest.ActionSort, msg.Action)
		values := make([]string, 10)
		data := msg.Data.(M)
		assert.Equal(t, "sort", data["key"])
		for i, v := range data["values"].([]interface{}) {
			values[i] = v.(string)
		}
		assert.Equal(t, projectIds, values)
		break
	}
}

func TestSortEventBad(t *testing.T) {
	RemoveStream(t)

	help.Reverse[string](projectIds)
	resp, err := Req("POST", "/projects/sort", M{
		"data": M{
			"key":    "sort",
			"values": projectIds,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode())
	RecoverStream(t)
}

var couponId primitive.ObjectID

func TestMoreTransform(t *testing.T) {
	resp, err := Req("POST", "/coupons/create", M{
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
			"pd":                "timestamp",
			"valid":             "timestamps",
			"metadata->$->date": "date",
			"metadata->$->wm":   "dates",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	id := result["InsertedID"].(string)
	couponId, _ = primitive.ObjectIDFromHex(id)

	var coupon M
	err = service.Db.Collection("coupons").FindOne(context.TODO(), bson.M{
		"_id": couponId,
	}).Decode(&coupon)
	assert.NoError(t, err)

	assert.Equal(t, "体验卡", coupon["name"])
	assert.Equal(t, primitive.DateTime(1681336800906), coupon["pd"])
	assert.Equal(t,
		primitive.A{primitive.DateTime(1681336800906), primitive.DateTime(1681367405586)},
		coupon["valid"],
	)
}

func TestCipherTransform(t *testing.T) {
	cards := []string{
		"11223344",
		"55667788",
	}
	cardsText, _ := sonic.MarshalString(cards)
	idcard := M{
		"no": "789987",
		"x1": "www",
		"x2": "ccc",
	}
	idcardText, _ := sonic.MarshalString(idcard)
	resp, err := Req("POST", "/members/create", M{
		"data": M{
			"name":   "用户A",
			"phone":  "12345678",
			"cards":  cardsText,
			"idcard": idcardText,
		},
		"xdata": M{
			"phone":  "cipher",
			"cards":  "cipher",
			"idcard": "cipher",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())

	var result M
	err = sonic.Unmarshal(resp.Body(), &result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	id := result["InsertedID"].(string)
	oid, _ := primitive.ObjectIDFromHex(id)

	var member M
	err = service.Db.Collection("members").FindOne(context.TODO(), bson.M{
		"_id": oid,
	}).Decode(&member)
	assert.NoError(t, err)
	assert.Equal(t, "用户A", member["name"])

	var cardsV []string
	cardsBytes, err := service.Cipher.Decode(member["cards"].(string))
	assert.NoError(t, err)
	err = sonic.Unmarshal(cardsBytes, &cardsV)
	assert.NoError(t, err)
	//t.Log(cardsV)
	assert.ElementsMatch(t, cards, cardsV)

	var idcardV M
	idcardBytes, err := service.Cipher.Decode(member["idcard"].(string))
	assert.NoError(t, err)
	err = sonic.Unmarshal(idcardBytes, &idcardV)
	assert.NoError(t, err)
	//t.Log(idcardV)
	assert.Equal(t, idcard, idcardV)
}

func TestTxBulkCreate(t *testing.T) {
	ctx := context.TODO()
	err := service.Db.Collection("x_test").Drop(ctx)
	assert.NoError(t, err)

	Transaction(t, func(txn string) {
		resp, err := Req("POST", "/x_test/bulk_create", M{
			"data": []M{
				{"name": "abc"},
				{"name": "xxx"},
			},
			"txn": txn,
		})
		assert.NoError(t, err)
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
		resp, err := Req("POST", "/x_test/update", M{
			"filter": M{"name": "kain"},
			"data": M{
				"$set": M{"name": "xxxx"},
			},
			"txn": txn,
		})
		assert.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode())

		var result M
		err = service.Db.Collection("x_test").FindOne(ctx, bson.M{
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
		resp, err := Req("PATCH", fmt.Sprintf(`/x_test/%s`, id.Hex()), M{
			"data": M{
				"$set": M{"name": "xxxx"},
			},
			"txn": txn,
		})
		assert.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode())

		var result M
		err = service.Db.Collection("x_test").FindOne(ctx, bson.M{
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
		resp, err := Req("PUT", fmt.Sprintf(`/x_test/%s`, id.Hex()), M{
			"data": M{
				"name": "xxxx",
			},
			"txn": txn,
		})
		assert.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode())

		var result M
		err = service.Db.Collection("x_test").FindOne(ctx, bson.M{
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
		u := Url(fmt.Sprintf(`/x_test/%s`, id.Hex()), Params{
			{"txn", txn},
		})
		resp, err := Req("DELETE", u, nil)
		assert.NoError(t, err)
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
		resp, err := Req("POST", "/x_test/bulk_delete", M{
			"filter": M{
				"name": "kain",
			},
			"txn": txn,
		})
		assert.NoError(t, err)
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
		resp, err := Req("POST", "/x_test/sort", M{
			"data": M{
				"key":    "sort",
				"values": sources,
			},
			"txn": txn,
		})
		assert.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode())

		cursor, err := service.Db.Collection("x_test").Find(ctx, bson.M{})
		assert.NoError(t, err)
		var result []M
		err = cursor.All(ctx, &result)
		assert.NoError(t, err)
		//t.Log(result)
	})

	cursor, err := service.Db.Collection("x_test").Find(ctx, bson.M{})
	assert.NoError(t, err)
	var result []M
	err = cursor.All(ctx, &result)
	assert.NoError(t, err)
	//t.Log(result)
}

func TestTransaction(t *testing.T) {
	ctx := context.TODO()
	err := service.Db.Collection("x_users").Drop(ctx)
	assert.NoError(t, err)
	err = service.Db.Collection("x_roles").Drop(ctx)
	assert.NoError(t, err)

	Transaction(t, func(txn string) {
		resp1, err := Req("POST", "/x_roles/create", M{
			"data": M{
				"name": "admin",
				"key":  "*",
			},
			"txn": txn,
		})
		assert.NoError(t, err)
		assert.Equal(t, 204, resp1.StatusCode())

		resp2, err := Req("POST", "/x_users/create", M{
			"data": M{
				"name":  "kainxxxx",
				"roles": []string{"*"},
			},
			"txn": txn,
		})
		assert.NoError(t, err)
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
	resp1, err := Req("POST", "/commit", nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp1.StatusCode())

	resp2, err := Req("POST", "/commit", M{
		"txn": help.Uuid(),
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode())
}

func TestCommitTimeout(t *testing.T) {
	service.Values.RestTxnTimeout = time.Second
	resp1, err := Req("POST", "/transaction", nil)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp1.StatusCode())

	var result1 M
	err = sonic.Unmarshal(resp1.Body(), &result1)
	assert.NoError(t, err)
	txn := result1["txn"].(string)

	time.Sleep(time.Second)

	resp2, err := Req("POST", "/commit", M{
		"txn": txn,
	})
	assert.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode())
}

func TestUpdateByIdWithArrayFilters(t *testing.T) {
	resp, err := Req("PATCH", fmt.Sprintf(`/coupons/%s`, couponId.Hex()), M{
		"data": M{
			"$set": M{
				"metadata.$[i].version": "v1",
			},
		},
		"arrayFilters": []M{
			{"i.name": "aps"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode(), string(resp.Body()))

	ctx := context.TODO()
	var data M
	err = service.Db.Collection("coupons").
		FindOne(ctx, bson.M{"_id": couponId}).
		Decode(&data)
	assert.NoError(t, err)

	assert.Equal(t, "体验卡", data["name"])
	assert.Equal(t, "v1", data["metadata"].(primitive.A)[0].(M)["version"])
}

func TestUpdateWithArrayFilters(t *testing.T) {
	resp, err := Req("POST", "/coupons/update", M{
		"filter": M{
			"name": "体验卡",
		},
		"data": M{
			"$set": M{
				"metadata.$[i].version": "v2",
			},
		},
		"arrayFilters": []M{
			{"i.name": "aps"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode(), string(resp.Body()))

	ctx := context.TODO()
	var data M
	err = service.Db.Collection("coupons").
		FindOne(ctx, bson.M{"_id": couponId}).
		Decode(&data)
	assert.NoError(t, err)

	assert.Equal(t, "体验卡", data["name"])
	assert.Equal(t, "v2", data["metadata"].(primitive.A)[0].(M)["version"])
}
