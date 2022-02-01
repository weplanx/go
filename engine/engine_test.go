package engine

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
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
	"testing"
)

var r *gin.Engine
var db *mongo.Database

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	r = gin.Default()
	x := New(
		SetApp("testing"),
		UseStaticOptions(map[string]Option{
			"users": {
				Projection: map[string]interface{}{
					"password": 0,
				},
			},
		}),
	)
	client, err := mongo.Connect(
		context.TODO(),
		options.Client().ApplyURI(os.Getenv("TEST_DB")),
	)
	if err != nil {
		panic(err)
	}
	db = client.Database("example")
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

func TestCreate(t *testing.T) {
	if err := db.Drop(context.TODO()); err != nil {
		panic(err)
	}
	// 创建文档
	res1 := httptest.NewRecorder()
	body1, err := jsoniter.Marshal(CreateBody{
		Doc: map[string]interface{}{
			"name": "agent",
		},
	})
	if err != nil {
		panic(err)
	}
	req1, _ := http.NewRequest("POST", "/privileges", bytes.NewBuffer(body1))

	r.ServeHTTP(res1, req1)
	assert.Equal(t, 201, res1.Code)

	var result1 map[string]interface{}
	if err := jsoniter.Unmarshal(res1.Body.Bytes(), &result1); err != nil {
		panic(err)
	}
	exists, err := db.Collection("privileges").CountDocuments(context.TODO(), bson.M{
		"name": "agent",
	})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(1), exists)

	// 包含引用
	res2 := httptest.NewRecorder()
	insertedID := result1["InsertedID"].(string)
	body2, err := jsoniter.Marshal(CreateBody{
		Doc: map[string]interface{}{
			"privileges": []string{insertedID},
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
	req2, _ := http.NewRequest("POST", "/example", bytes.NewBuffer(body2))

	r.ServeHTTP(res2, req2)
	assert.Equal(t, 201, res2.Code)

	var result2 map[string]interface{}
	oid, _ := primitive.ObjectIDFromHex(insertedID)
	if err = db.Collection("example").FindOne(context.TODO(), bson.M{
		"privileges": bson.M{"$in": bson.A{oid}},
	}).Decode(&result2); err != nil {
		t.Error(err)
	}
	assert.Equal(t, "Kenny Boyer", result2["name"])
	t.Log(res2.Body.String())

	// 格式转换
	res3 := httptest.NewRecorder()
	body3, err := jsoniter.Marshal(CreateBody{
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
	req3, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body3))

	r.ServeHTTP(res3, req3)
	assert.Equal(t, 201, res3.Code)

	var result3 map[string]interface{}
	if err = db.Collection("users").FindOne(context.TODO(), bson.M{
		"name": "admin",
	}).Decode(&result3); err != nil {
		t.Error(err)
	}
	assert.Nil(t, password.Verify("adx8090", result3["password"].(string)))
	assert.True(t, true, result3["alias"].(primitive.ObjectID))

	// 批量创建
	res4 := httptest.NewRecorder()
	body4, err := jsoniter.Marshal(CreateBody{Docs: mock})
	if err != nil {
		panic(err)
	}
	req4, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(body4))
	r.ServeHTTP(res4, req4)
	assert.Equal(t, 201, res4.Code)
	count, err := db.Collection("services").
		CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(50), count)
}

func TestFindOne(t *testing.T) {
	var body map[string]interface{}
	// 通过条件筛选
	res1 := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"number": "55826199",
	})
	if err != nil {
		t.Error(err)
	}
	req1, _ := http.NewRequest("GET", "/services", nil)
	query1 := req1.URL.Query()
	query1.Add("where", string(where))
	query1.Add("single", "true")
	req1.URL.RawQuery = query1.Encode()

	r.ServeHTTP(res1, req1)
	assert.Equal(t, 200, res1.Code)

	if err = jsoniter.Unmarshal(res1.Body.Bytes(), &body); err != nil {
		t.Error(err)
	}
	assert.Equal(t, "Handmade Soft Salad", body["name"])

	// 通过ID获取
	res2 := httptest.NewRecorder()
	url2 := fmt.Sprintf(`/services/%s`, body["_id"].(string))
	req2, _ := http.NewRequest("GET", url2, nil)

	r.ServeHTTP(res2, req2)
	assert.Equal(t, 200, res2.Code)

	if err = jsoniter.Unmarshal(res2.Body.Bytes(), &body); err != nil {
		t.Error(err)
	}
	assert.Equal(t, float64(727), body["price"])
}

func TestFind(t *testing.T) {
	var body []map[string]interface{}
	// 获取文档
	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/services", nil)

	r.ServeHTTP(res1, req1)
	assert.Equal(t, 200, res1.Code)

	if err := jsoniter.Unmarshal(res1.Body.Bytes(), &body); err != nil {
		t.Error(err)
	}
	a1, b1 := funk.Difference(
		funk.Map(mock, func(x map[string]interface{}) string {
			return x["name"].(string)
		}),
		funk.Map(body, func(x map[string]interface{}) string {
			return x["name"].(string)
		}),
	)
	assert.Empty(t, a1)
	assert.Empty(t, b1)

	// 多个文档过滤
	res2 := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"number": map[string]interface{}{"$in": []string{"55826199", "57277117"}},
	})
	if err != nil {
		t.Error(err)
	}
	req2, _ := http.NewRequest("GET", "/services", nil)
	query2 := req2.URL.Query()
	query2.Add("where", string(where))
	req2.URL.RawQuery = query2.Encode()

	r.ServeHTTP(res2, req2)
	assert.Equal(t, 200, res2.Code)

	if err := jsoniter.Unmarshal(res2.Body.Bytes(), &body); err != nil {
		t.Error(err)
	}
	a2, b2 := funk.Difference(
		[]string{"Handmade Soft Salad", "Intelligent Fresh Shoes"},
		funk.Map(body, func(x map[string]interface{}) string {
			return x["name"].(string)
		}),
	)
	assert.Empty(t, a2)
	assert.Empty(t, b2)

	// 多个文档ID
	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/services", nil)
	query3 := req3.URL.Query()
	for _, x := range body {
		query3.Add("id", x["_id"].(string))
	}
	req3.URL.RawQuery = query3.Encode()

	r.ServeHTTP(res3, req3)
	assert.Equal(t, 200, res3.Code)

	if err := jsoniter.Unmarshal(res3.Body.Bytes(), &body); err != nil {
		t.Error(err)
	}
	a3, b3 := funk.Difference(
		[]float64{float64(727), float64(47)},
		funk.Map(body, func(x map[string]interface{}) float64 {
			return x["price"].(float64)
		}),
	)
	assert.Empty(t, a3)
	assert.Empty(t, b3)
}

func TestFindPage(t *testing.T) {
	var body []map[string]interface{}
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/services", nil)
	req.Header.Set("x-page-size", "5")
	req.Header.Set("x-page", "1")
	query := req.URL.Query()
	query.Add("sort", "price.1")
	req.URL.RawQuery = query.Encode()
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	if err := jsoniter.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Error(err)
	}

	assert.Equal(t, 5, len(body))
	assert.Equal(t,
		[]string{"44243580", "57277117", "87239908", "60599531", "38438365"},
		funk.Map(body, func(x map[string]interface{}) string {
			return x["number"].(string)
		}),
	)
	assert.Equal(t, "50", res.Header().Get("X-Page-Total"))
}

func TestUpdateOne(t *testing.T) {
	res1 := httptest.NewRecorder()
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
	req1, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query1 := req1.URL.Query()
	query1.Add("where", string(where))
	query1.Add("single", "true")
	req1.URL.RawQuery = query1.Encode()

	r.ServeHTTP(res1, req1)
	assert.Equal(t, 200, res1.Code)

	var data map[string]interface{}
	if err = db.Collection("services").
		FindOne(context.TODO(), bson.M{"number": "38438365"}).
		Decode(&data); err != nil {
		t.Error(err)
	}

	assert.Equal(t, float64(512), data["price"])

	// 使用 ID 更新
	res2 := httptest.NewRecorder()
	body, err = jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"price": 1024.00,
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
	url2 := fmt.Sprintf(`/services/%s`, data["_id"].(primitive.ObjectID).Hex())
	req2, _ := http.NewRequest("PATCH", url2, bytes.NewBuffer(body))

	r.ServeHTTP(res2, req2)
	assert.Equal(t, 200, res2.Code)

	if err = db.Collection("services").
		FindOne(context.TODO(), bson.M{"number": "38438365"}).
		Decode(&data); err != nil {
		t.Error(err)
	}

	assert.Equal(t, float64(1024), data["price"])
}

func TestUpdateMany(t *testing.T) {
	res1 := httptest.NewRecorder()
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
	req1, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query1 := req1.URL.Query()
	query1.Add("where", string(where))
	req1.URL.RawQuery = query1.Encode()

	r.ServeHTTP(res1, req1)
	assert.Equal(t, 200, res1.Code)

	var data []map[string]interface{}
	cursor, err := db.Collection("services").Find(context.TODO(), bson.M{
		"number": bson.M{"$in": bson.A{"66502334", "43678700"}},
	})
	if err != nil {
		t.Error(err)
	}
	if err = cursor.All(context.TODO(), &data); err != nil {
		t.Error(err)
	}

	assert.Equal(t,
		[]float64{512, 512},
		funk.Map(data, func(x map[string]interface{}) float64 {
			return x["price"].(float64)
		}),
	)

	res2 := httptest.NewRecorder()
	body, err = jsoniter.Marshal(UpdateBody{
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"price": 1024.00,
			},
		},
	})
	req2, _ := http.NewRequest("PATCH", "/services", bytes.NewBuffer(body))
	query2 := req2.URL.Query()
	for _, x := range data {
		query2.Add("id", x["_id"].(primitive.ObjectID).Hex())
	}
	req2.URL.RawQuery = query2.Encode()

	r.ServeHTTP(res2, req2)
	assert.Equal(t, 200, res2.Code)

	cursor, err = db.Collection("services").Find(context.TODO(), bson.M{
		"number": bson.M{"$in": bson.A{"66502334", "43678700"}},
	})
	if err != nil {
		t.Error(err)
	}
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

func TestReplaceOne(t *testing.T) {
	var data map[string]interface{}
	if err := db.Collection("services").FindOne(context.TODO(), bson.M{
		"number": "62787493",
	}).Decode(&data); err != nil {
		t.Error(err)
	}
	id := data["_id"].(primitive.ObjectID).Hex()
	delete(data, "_id")
	delete(data, "create_time")
	delete(data, "update_time")
	data["price"] = 777.00
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, id)
	body, err := jsoniter.Marshal(ReplaceOneBody{
		Doc: data,
	})
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
	assert.Equal(t, float64(777), data["price"])
}

func TestDeleteOne(t *testing.T) {
	var data map[string]interface{}
	if err := db.Collection("services").FindOne(context.TODO(), bson.M{
		"number": "35433318",
	}).Decode(&data); err != nil {
		t.Error(err)
	}
	res := httptest.NewRecorder()
	url := fmt.Sprintf(`/services/%s`, data["_id"].(primitive.ObjectID).Hex())
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

func TestFindUsers(t *testing.T) {
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
