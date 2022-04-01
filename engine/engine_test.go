package engine

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
	"github.com/weplanx/go/helper"
	"github.com/weplanx/go/password"
	"github.com/weplanx/go/route"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

var r *gin.Engine
var db *mongo.Database
var nc *nats.Conn
var js nats.JetStreamContext

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	r = gin.Default()
	client, err := mongo.Connect(
		context.TODO(),
		options.Client().ApplyURI(os.Getenv("TEST_DB")),
	)
	if err != nil {
		panic(err)
	}
	db = client.Database("test")
	var kp nkeys.KeyPair
	if kp, err = nkeys.FromSeed([]byte(os.Getenv("TEST_NATS_NKEY"))); err != nil {
		return
	}
	defer kp.Wipe()
	var pub string
	if pub, err = kp.PublicKey(); err != nil {
		return
	}
	if !nkeys.IsValidPublicUserKey(pub) {
		panic("nkey 验证失败")
	}
	if nc, err = nats.Connect(
		os.Getenv("TEST_NATS"),
		nats.MaxReconnects(5),
		nats.ReconnectWait(2*time.Second),
		nats.ReconnectJitter(500*time.Millisecond, 2*time.Second),
		nats.Nkey(pub, func(nonce []byte) ([]byte, error) {
			sig, _ := kp.Sign(nonce)
			return sig, nil
		}),
	); err != nil {
		panic(err)
	}
	if js, err = nc.JetStream(nats.PublishAsyncMaxPending(256)); err != nil {
		panic(err)
	}
	x := New(
		SetApp("test"),
		UseStaticOptions(map[string]Option{
			"pages": {
				Event: true,
			},
			"users": {
				Field: []string{"name", "alias"},
			},
		}),
		UseEvents(js),
	)
	service := Service{
		Engine: x,
		Db:     db,
	}
	controller := Controller{
		Engine:  x,
		Service: &service,
	}
	helper.ExtendValidation()
	r.POST("/:model", route.Use(controller.Create))
	r.GET("/:model", route.Use(controller.Find))
	r.GET("/:model/:id", route.Use(controller.FindOneById))
	r.PATCH("/:model", route.Use(controller.Update))
	r.PATCH("/:model/:id", route.Use(controller.UpdateOneById))
	r.PUT("/:model/:id", route.Use(controller.ReplaceOneById))
	r.DELETE("/:model/:id", route.Use(controller.DeleteOneById))
	r.GET("svc/:id", route.Use(controller.FindOneById, route.SetModel("services")))
	if err := db.Drop(context.TODO()); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

var mock = []map[string]interface{}{
	{"number": "55826199", "name": "Handmade Soft Salad", "price": 727.00},
	{"number": "57277117", "name": "Intelligent Fresh Shoes", "price": 47.00},
	{"number": "52697132", "name": "Practical Metal Chips", "price": 859.00},
	{"number": "66502334", "name": "Ergonomic Wooden Pizza", "price": 328.00},
	{"number": "43678700", "name": "Intelligent Cotton Chips", "price": 489.00},
	{"number": "66204618", "name": "Sleek Rubber Cheese", "price": 986.00},
	{"number": "82877045", "name": "Unbranded Fresh Ball", "price": 915.00},
	{"number": "11254621", "name": "Handmade Metal Keyboard", "price": 244.00},
	{"number": "24443471", "name": "Rustic Frozen Gloves", "price": 500.00},
	{"number": "74371061", "name": "Awesome Frozen Gloves", "price": 214.00},
	{"number": "39012286", "name": "Sleek Steel Bike", "price": 428.00},
	{"number": "58946201", "name": "Handmade Plastic Pizza", "price": 913.00},
	{"number": "08945471", "name": "Generic Metal Pizza", "price": 810.00},
	{"number": "40208673", "name": "Handcrafted Granite Shoes", "price": 429.00},
	{"number": "84106393", "name": "Refined Steel Bike", "price": 339.00},
	{"number": "52669450", "name": "Handmade Frozen Keyboard", "price": 684.00},
	{"number": "15525688", "name": "Tasty Cotton Pants", "price": 995.00},
	{"number": "38438365", "name": "Awesome Soft Soap", "price": 142.00},
	{"number": "48780690", "name": "Intelligent Cotton Gloves", "price": 297.00},
	{"number": "62787493", "name": "Rustic Frozen Salad", "price": 542.00},
	{"number": "35433318", "name": "Small Soft Keyboard", "price": 703.00},
	{"number": "87239908", "name": "Handmade Granite Sausages", "price": 88.00},
	{"number": "63793023", "name": "Intelligent Soft Bike", "price": 630.00},
	{"number": "60599531", "name": "Unbranded Wooden Bacon", "price": 98.00},
	{"number": "10550233", "name": "Intelligent Steel Tuna", "price": 499.00},
	{"number": "89885575", "name": "Unbranded Frozen Chicken", "price": 667.00},
	{"number": "90424834", "name": "Handcrafted Wooden Shoes", "price": 516.00},
	{"number": "77762017", "name": "Generic Rubber Table", "price": 725.00},
	{"number": "07605361", "name": "Incredible Metal Towels", "price": 261.00},
	{"number": "92417878", "name": "Small Fresh Table", "price": 662.00},
	{"number": "12181549", "name": "Refined Soft Ball", "price": 385.00},
	{"number": "23740764", "name": "Unbranded Soft Mouse", "price": 710.00},
	{"number": "75813798", "name": "Tasty Metal Chips", "price": 506.00},
	{"number": "70353191", "name": "Tasty Cotton Hat", "price": 480.00},
	{"number": "67153899", "name": "Generic Frozen Bike", "price": 261.00},
	{"number": "14395918", "name": "Awesome Steel Towels", "price": 796.00},
	{"number": "24957863", "name": "Ergonomic Soft Chair", "price": 599.00},
	{"number": "84480037", "name": "Fantastic Metal Salad", "price": 273.00},
	{"number": "10531004", "name": "Tasty Rubber Bike", "price": 696.00},
	{"number": "37050804", "name": "Intelligent Soft Pants", "price": 451.00},
	{"number": "15757338", "name": "Fantastic Fresh Soap", "price": 281.00},
	{"number": "83666844", "name": "Rustic Wooden Shoes", "price": 477.00},
	{"number": "60049311", "name": "Refined Steel Pizza", "price": 719.00},
	{"number": "25627132", "name": "Licensed Wooden Bacon", "price": 585.00},
	{"number": "44243580", "name": "Handmade Granite Fish", "price": 3.00},
	{"number": "67644215", "name": "Refined Plastic Keyboard", "price": 796.00},
	{"number": "99821780", "name": "Refined Frozen Pants", "price": 569.00},
	{"number": "09613501", "name": "Handcrafted Soft Sausages", "price": 826.00},
	{"number": "35568587", "name": "Practical Soft Sausages", "price": 500.00},
	{"number": "92044481", "name": "Sleek Soft Soap", "price": 309.00},
}

// 创建文档不合规的 URL
func TestCreateUrlError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/Privileges", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 创建文档不合规的请求体
func TestCreateBodyError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/privileges", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

var createInsertedID string

// 创建文档
func TestCreate(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(CreateBody{
		Doc: map[string]interface{}{
			"name": "agent",
		},
	})
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("POST", "/privileges", bytes.NewBuffer(body))

	r.ServeHTTP(res, req)
	assert.Equal(t, 201, res.Code)

	var result map[string]interface{}
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &result); err != nil {
		panic(err)
	}
	count, err := db.Collection("privileges").CountDocuments(context.TODO(), bson.M{
		"name": "agent",
	})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(1), count)
	createInsertedID = result["InsertedID"].(string)
}

// 创建文档不合规的引用
func TestCreateRefError(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(CreateBody{
		Doc: map[string]interface{}{
			"privileges": []string{"abc", "d1"},
			"name":       "Kenny Boyer",
			"account":    "Lyda_Mosciski",
			"email":      "Lempi.Larkin60@yahoo.com",
			"phone":      "(403) 332-1896 x64468",
			"address":    "924 Braulio Radial",
		},
		Ref: []string{"privileges"},
	})
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("POST", "/example", bytes.NewBuffer(body))

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 创建文档包含引用
func TestCreateWithRef(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(CreateBody{
		Doc: map[string]interface{}{
			"privileges": []string{createInsertedID},
			"name":       "Kenny Boyer",
			"account":    "Lyda_Mosciski",
			"email":      "Lempi.Larkin60@yahoo.com",
			"phone":      "(403) 332-1896 x64468",
			"address":    "924 Braulio Radial",
		},
		Ref: []string{"privileges", "nothing"},
	})
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("POST", "/example", bytes.NewBuffer(body))

	r.ServeHTTP(res, req)
	assert.Equal(t, 201, res.Code)

	var result map[string]interface{}
	oid, _ := primitive.ObjectIDFromHex(createInsertedID)
	if err = db.Collection("example").FindOne(context.TODO(), bson.M{
		"privileges": bson.M{"$in": bson.A{oid}},
	}).Decode(&result); err != nil {
		t.Error(err)
	}
	assert.Equal(t, "Kenny Boyer", result["name"])
}

// 创建文档格式转换
func TestCreateFormat(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(CreateBody{
		Doc: map[string]interface{}{
			"name":     "admin",
			"alias":    "61f7ef84dfdb15138a09cdad",
			"password": "adx8090",
		},
		Format: map[string]interface{}{
			"alias":    "object_id",
			"password": "password",
		},
	})
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))

	r.ServeHTTP(res, req)
	assert.Equal(t, 201, res.Code)

	var result map[string]interface{}
	if err = db.Collection("users").FindOne(context.TODO(), bson.M{
		"name": "admin",
	}).Decode(&result); err != nil {
		t.Error(err)
	}
	assert.Nil(t, password.Verify("adx8090", result["password"].(string)))
	assert.True(t, true, result["alias"].(primitive.ObjectID))
}

// 创建文档格式转换不存在字段忽略
func TestCreateFormatIgnore(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(CreateBody{
		Doc: map[string]interface{}{
			"name": "agent",
		},
		Format: map[string]interface{}{
			"parent": "object_id",
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/privileges", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t, 201, res.Code)
}

// 创建文档格式转换错误
func TestCreateFormatError(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(CreateBody{
		Doc: map[string]interface{}{
			"name": "agent",
		},
		Format: map[string]interface{}{
			"name": "object_id",
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/privileges", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 创建多个文档
func TestCreateMany(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(CreateBody{Docs: mock})
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t, 201, res.Code)
	count, err := db.Collection("services").
		CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(50), count)
}

// 创建多个文档不合规的引用
func TestCreateManyRefError(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(CreateBody{
		Docs: []map[string]interface{}{
			{
				"privileges": []string{"abc", "d1"},
			},
		},
		Ref: []string{"privileges"},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/privileges", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 创建多个文档格式转换错误
func TestCreateManyFormatError(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(CreateBody{
		Docs: []map[string]interface{}{
			{
				"name": "agent",
			},
		},
		Format: map[string]interface{}{
			"name": "object_id",
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("POST", "/privileges", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 获取所有文档不合规的 URL
func TestFindUrlError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/Services", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 获取所有文档
func TestFind(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var data []map[string]interface{}
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Error(err)
	}
	a, b := funk.Difference(
		funk.Map(mock, func(x map[string]interface{}) string {
			return x["name"].(string)
		}),
		funk.Map(data, func(x map[string]interface{}) string {
			return x["name"].(string)
		}),
	)
	assert.Empty(t, a)
	assert.Empty(t, b)
}

// 获取多个文档不合规的查询
func TestFindWithWhereError(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"$x": "",
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("GET", "/services", nil)
	query := req.URL.Query()
	query.Add("where", string(where))
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 获取多个文档不合规的排序
func TestFindWithSortError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	query := req.URL.Query()
	query.Add("sort", "price.2")
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

var findWithWhereData []map[string]interface{}

// 获取多个文档（过滤）
func TestFindWithWhere(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"number": map[string]interface{}{"$in": []string{"55826199", "57277117"}},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("GET", "/services", nil)
	query := req.URL.Query()
	query.Add("where", string(where))
	query.Add("sort", "price.1")
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	if err := jsoniter.Unmarshal(res.Body.Bytes(), &findWithWhereData); err != nil {
		t.Error(err)
	}
	a, b := funk.Difference(
		[]string{"Handmade Soft Salad", "Intelligent Fresh Shoes"},
		funk.Map(findWithWhereData, func(x map[string]interface{}) string {
			return x["name"].(string)
		}),
	)
	assert.Empty(t, a)
	assert.Empty(t, b)
}

// 获取多个文档（ID）
func TestFindWithId(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	query := req.URL.Query()
	for _, x := range findWithWhereData {
		query.Add("id", x["_id"].(string))
	}
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var data []map[string]interface{}
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Error(err)
	}
	a, b := funk.Difference(
		[]float64{float64(727), float64(47)},
		funk.Map(data, func(x map[string]interface{}) float64 {
			return x["price"].(float64)
		}),
	)
	assert.Empty(t, a)
	assert.Empty(t, b)
}

// 获取分页文档不合规的请求头部
func TestFindPageHeaderError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	req.Header.Set("x-page-size", "5")
	req.Header.Set("x-page", "-1")

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 获取分页文档不合规的查询条件
func TestFindPageWithWhereError(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"$x": "",
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("GET", "/services", nil)
	req.Header.Set("x-page-size", "5")
	req.Header.Set("x-page", "1")
	query := req.URL.Query()
	query.Add("where", string(where))
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 获取分页文档
func TestFindPage(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	req.Header.Set("x-page-size", "5")
	req.Header.Set("x-page", "1")
	query := req.URL.Query()
	query.Add("sort", "price.1")
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var data []map[string]interface{}
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Error(err)
	}
	assert.Equal(t, 5, len(data))
	assert.Equal(t,
		[]string{"44243580", "57277117", "87239908", "60599531", "38438365"},
		funk.Map(data, func(x map[string]interface{}) string {
			return x["number"].(string)
		}),
	)
	assert.Equal(t, "50", res.Header().Get("X-Page-Total"))
}

// 获取当个文档不合规的查询
func TestFindOneWithWhereError(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"$x": "",
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("GET", "/services", nil)
	query := req.URL.Query()
	query.Add("where", string(where))
	query.Add("single", "true")
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

var findOneData map[string]interface{}

// 获取当个文档（过滤）
func TestFindOne(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"number": "55826199",
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("GET", "/services", nil)
	query := req.URL.Query()
	query.Add("where", string(where))
	query.Add("single", "true")
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	if err = jsoniter.Unmarshal(res.Body.Bytes(), &findOneData); err != nil {
		t.Error(err)
	}
	assert.Equal(t, "Handmade Soft Salad", findOneData["name"])
}

// 获取单个文档，非 object_id 返回错误
func TestFindOneByIdNotObjectId(t *testing.T) {
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, "abc")
	req, _ := http.NewRequest("GET", url, nil)

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 获取单个文档，不存在的 ID
func TestFindOneByIdNotExists(t *testing.T) {
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, primitive.NewObjectID().Hex())
	req, _ := http.NewRequest("GET", url, nil)

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 获取单个文档（ID）
func TestFindOneById(t *testing.T) {
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, findOneData["_id"].(string))
	req, _ := http.NewRequest("GET", url, nil)

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	if err := jsoniter.Unmarshal(res.Body.Bytes(), &findOneData); err != nil {
		t.Error(err)
	}
	assert.Equal(t, float64(727), findOneData["price"])
}

// 局部更新文档不合规的 URL
func TestUpdateManyUrlError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/Services", nil)

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 局部更新文档空条件
func TestUpdateManyEmptyWhere(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/services", nil)
	query := req.URL.Query()
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 局部更新文档不合规的查询
func TestUpdateManyWithWhereError(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"$x": "",
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/services", nil)
	query := req.URL.Query()
	query.Add("where", string(where))
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 局部更新格式转换错误
func TestUpdateManyFormatError(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"number": map[string]interface{}{
			"$in": []string{"66502334", "43678700"},
		},
	})
	if err != nil {
		t.Error(err)
	}
	body, err := jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"name": "agent",
			},
		},
		Format: map[string]interface{}{
			"name": "object_id",
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query := req.URL.Query()
	query.Add("where", string(where))
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 局部更新不合规的引用
func TestUpdateManyRefError(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"number": map[string]interface{}{
			"$in": []string{"66502334", "43678700"},
		},
	})
	if err != nil {
		t.Error(err)
	}
	body, err := jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"tag": []string{"a1", "a2"},
			},
		},
		Ref: []string{"tag"},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query := req.URL.Query()
	query.Add("where", string(where))
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

var updateManyData []map[string]interface{}

// 局部更新多个文档（过滤）
func TestUpdateMany(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"number": map[string]interface{}{
			"$in": []string{"66502334", "43678700"},
		},
	})
	if err != nil {
		t.Error(err)
	}
	body, err := jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"price": 512.00,
			},
		},
	})

	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query := req.URL.Query()
	query.Add("where", string(where))
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	cursor, err := db.Collection("services").Find(context.TODO(), bson.M{
		"number": bson.M{"$in": bson.A{"66502334", "43678700"}},
	})
	if err != nil {
		t.Error(err)
	}
	if err = cursor.All(context.TODO(), &updateManyData); err != nil {
		t.Error(err)
	}

	assert.Equal(t,
		[]float64{512, 512},
		funk.Map(updateManyData, func(x map[string]interface{}) float64 {
			return x["price"].(float64)
		}),
	)
}

// 局部更新多个文档（ID）格式转换错误
func TestUpdateManyByIdFormatError(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"name": "agent",
			},
		},
		Format: map[string]interface{}{
			"name": "object_id",
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query := req.URL.Query()
	for _, x := range updateManyData {
		query.Add("id", x["_id"].(primitive.ObjectID).Hex())
	}
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 局部更新多个文档（ID）
func TestUpdateManyById(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"price": 1024.00,
			},
		},
	})
	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query := req.URL.Query()
	for _, x := range updateManyData {
		query.Add("id", x["_id"].(primitive.ObjectID).Hex())
	}
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	cursor, err := db.Collection("services").Find(context.TODO(), bson.M{
		"number": bson.M{"$in": bson.A{"66502334", "43678700"}},
	})
	if err != nil {
		t.Error(err)
	}
	var data []map[string]interface{}
	if err = cursor.All(context.TODO(), &data); err != nil {
		t.Error(err)
	}

	assert.Equal(t,
		[]float64{1024, 1024},
		funk.Map(data, func(x map[string]interface{}) float64 {
			return x["price"].(float64)
		}),
	)
}

var updateOneData map[string]interface{}

// 局部更新多个文档格式化错误
func TestUpdateOneFormatError(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"number": "38438365",
	})
	if err != nil {
		t.Error(err)
	}
	body, err := jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"name": "agent",
			},
		},
		Format: map[string]interface{}{
			"name": "object_id",
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query := req.URL.Query()
	query.Add("where", string(where))
	query.Add("single", "true")
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 局部更新多个文档引用错误
func TestUpdateOneRefError(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"number": "38438365",
	})
	if err != nil {
		t.Error(err)
	}
	body, err := jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"tag": []string{"a1", "a2"},
			},
		},
		Ref: []string{"tag"},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query := req.URL.Query()
	query.Add("where", string(where))
	query.Add("single", "true")
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 局部更新单个文档（过滤）
func TestUpdateOne(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"number": "38438365",
	})
	if err != nil {
		t.Error(err)
	}
	body, err := jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"price": 512.00,
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query := req.URL.Query()
	query.Add("where", string(where))
	query.Add("single", "true")
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	if err = db.Collection("services").
		FindOne(context.TODO(), bson.M{"number": "38438365"}).
		Decode(&updateOneData); err != nil {
		t.Error(err)
	}

	assert.Equal(t, float64(512), updateOneData["price"])
}

// 股部更新单个文档（ID）,非 object_id
func TestUpdateOneByIdNotObjectId(t *testing.T) {
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, "abc")
	req, _ := http.NewRequest("PATCH", url, nil)

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 局部更新单个文档空条件
func TestUpdateOneByIdEmptyWhere(t *testing.T) {
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, primitive.NewObjectID().Hex())
	req, _ := http.NewRequest("PATCH", url, nil)
	query := req.URL.Query()
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 局部更新单个文档（ID）
func TestUpdateOneByIdFormatError(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"name": "agent",
			},
		},
		Format: map[string]interface{}{
			"name": "object_id",
		},
	})
	if err != nil {
		t.Error(err)
	}
	url := fmt.Sprintf(`/services/%s`, updateOneData["_id"].(primitive.ObjectID).Hex())
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 局部更新单个文档（ID）
func TestUpdateOneById(t *testing.T) {
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"price": 1024.00,
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
	url := fmt.Sprintf(`/services/%s`, updateOneData["_id"].(primitive.ObjectID).Hex())
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var data map[string]interface{}
	if err = db.Collection("services").
		FindOne(context.TODO(), bson.M{"number": "38438365"}).
		Decode(&data); err != nil {
		t.Error(err)
	}

	assert.Equal(t, float64(1024), data["price"])
}

// 更新文档，非 object_id 返回错误
func TestReplaceOneNotObjectId(t *testing.T) {
	// 不合规的 object_id
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, "abc")
	req, _ := http.NewRequest("PUT", url, nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 更新文档不合规的请求内容
func TestReplaceOneBodyError(t *testing.T) {
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, primitive.NewObjectID().Hex())
	req, _ := http.NewRequest("PUT", url, nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 更新文档不合规的格式化处理
func TestReplaceOneFormatError(t *testing.T) {
	body, err := jsoniter.Marshal(ReplaceOneBody{
		Doc: map[string]interface{}{
			"name": "abc",
		},
		Format: map[string]interface{}{
			"name": "object_id",
		},
	})
	if err != nil {
		t.Error(err)
	}
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, primitive.NewObjectID().Hex())
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 更新文档，不合规的引用
func TestReplaceOneRefError(t *testing.T) {
	body, err := jsoniter.Marshal(ReplaceOneBody{
		Doc: map[string]interface{}{
			"tag": []string{"a1", "a2"},
		},
		Ref: []string{"tag"},
	})
	if err != nil {
		t.Error(err)
	}
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, primitive.NewObjectID().Hex())
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 更新文档
func TestReplaceOne(t *testing.T) {
	var doc map[string]interface{}
	if err := db.Collection("services").FindOne(context.TODO(), bson.M{
		"number": "62787493",
	}).Decode(&doc); err != nil {
		return
	}

	id := doc["_id"].(primitive.ObjectID).Hex()
	delete(doc, "_id")
	delete(doc, "create_time")
	delete(doc, "update_time")
	doc["price"] = 777.00

	body, err := jsoniter.Marshal(ReplaceOneBody{
		Doc: doc,
	})

	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, id)
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var update map[string]interface{}
	if err = db.Collection("services").FindOne(context.TODO(), bson.M{
		"number": "62787493",
	}).Decode(&update); err != nil {
		t.Error(err)
	}
	assert.Equal(t, float64(777), doc["price"])
}

// 删除文档，非 object_id 返回错误
func TestDeleteOneNotObjectId(t *testing.T) {
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, "abc")
	req, _ := http.NewRequest("DELETE", url, nil)

	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
}

// 删除文档
func TestDeleteOne(t *testing.T) {
	var doc map[string]interface{}
	if err := db.Collection("services").FindOne(context.TODO(), bson.M{
		"number": "35433318",
	}).Decode(&doc); err != nil {
		t.Error(err)
	}
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, doc["_id"].(primitive.ObjectID).Hex())
	req, _ := http.NewRequest("DELETE", url, nil)

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	count, err := db.Collection("services").CountDocuments(context.TODO(), bson.M{
		"number": "35433318",
	})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(0), count)
}

// 模型预定义测试
func TestPredefinedModel(t *testing.T) {
	var data map[string]interface{}
	if err := db.Collection("services").FindOne(context.TODO(), bson.M{
		"number": "55826199",
	}).Decode(&data); err != nil {
		t.Error(err)
	}
	res := httptest.NewRecorder()
	id := data["_id"].(primitive.ObjectID).Hex()
	req, _ := http.NewRequest("GET", fmt.Sprintf(`/svc/%s`, id), nil)
	r.ServeHTTP(res, req)
	var result map[string]interface{}
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &result); err != nil {
		t.Error(err)
	}
	assert.Equal(t, data["name"], result["name"])
}

// 获取多个文档，固定字段测试
func TestFindForStaticProjection(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var data []map[string]interface{}
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Error(err)
	}

	assert.Nil(t, data[0]["password"])
}

// 获取单个文档，固定字段测试
func TestFindOneForStaticProjection(t *testing.T) {
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"name": "admin",
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("GET", "/users", nil)
	query := req.URL.Query()
	query.Add("where", string(where))
	query.Add("single", "true")
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	var data map[string]interface{}
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Error(err)
	}

	assert.Nil(t, data["password"])
}

// 创建文档，队列事件测试
func TestCreateForStaticEvent(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	subj := "test.events.pages"
	queue := "test:events:pages"
	sub, err := js.QueueSubscribe(subj, queue, func(msg *nats.Msg) {
		assert.NotEmpty(t, msg.Data)
		wg.Done()
	})
	if err != nil {
		t.Error(err)
	}
	defer sub.Unsubscribe()
	res := httptest.NewRecorder()
	body, err := jsoniter.Marshal(CreateBody{
		Doc: map[string]interface{}{
			"name": "首页",
		},
	})
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("POST", "/pages", bytes.NewBuffer(body))

	r.ServeHTTP(res, req)
	assert.Equal(t, 201, res.Code)
	wg.Wait()
}

func TestUpdateForStaticEvent(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	subj := "test.events.pages"
	queue := "test:events:pages"
	sub, err := js.QueueSubscribe(subj, queue, func(msg *nats.Msg) {
		assert.NotEmpty(t, msg.Data)
		wg.Done()
	})
	if err != nil {
		t.Error(err)
	}
	defer sub.Unsubscribe()
	res := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"name": "首页",
	})
	if err != nil {
		t.Error(err)
	}
	body, err := jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"sort": 1,
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
	req, _ := http.NewRequest("PATCH", "/pages", bytes.NewBuffer(body))
	query := req.URL.Query()
	query.Add("where", string(where))
	query.Add("single", "true")
	req.URL.RawQuery = query.Encode()

	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)

	wg.Wait()
}
