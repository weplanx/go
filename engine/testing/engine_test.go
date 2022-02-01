package testing

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
	"github.com/weplanx/go/engine"
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
	"strings"
	"testing"
)

var r *gin.Engine
var db *mongo.Database

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	r = gin.Default()
	x := engine.New(
		engine.SetApp("testing"),
	)
	client, err := mongo.Connect(
		context.TODO(),
		options.Client().ApplyURI(os.Getenv("TEST_DB")),
	)
	if err != nil {
		panic(err)
	}
	db = client.Database("example")
	service := engine.Service{
		Engine: x,
		Db:     db,
	}
	controller := engine.Controller{
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

func TestController_Create(t *testing.T) {
	if err := db.Drop(context.TODO()); err != nil {
		panic(err)
	}
	// 创建文档
	res1 := httptest.NewRecorder()
	body1, err := jsoniter.Marshal(engine.CreateBody{
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
	body2, err := jsoniter.Marshal(engine.CreateBody{
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
	body3, err := jsoniter.Marshal(engine.CreateBody{
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
	body4, err := jsoniter.Marshal(engine.CreateBody{
		Docs: []map[string]interface{}{
			{
				"name": "Genevieve50",
			},
			{
				"name": "Jaden32",
			},
			{
				"name": "Ottilie90",
			},
		},
	})
	if err != nil {
		panic(err)
	}
	req4, _ := http.NewRequest("POST", "/members", bytes.NewBuffer(body4))
	r.ServeHTTP(res4, req4)
	assert.Equal(t, 201, res4.Code)
	count, err := db.Collection("members").
		CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(3), count)
}

func TestController_Find(t *testing.T) {
	// 获取文档
	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/members", nil)
	r.ServeHTTP(res1, req1)
	var body []map[string]interface{}
	if err := jsoniter.Unmarshal(res1.Body.Bytes(), &body); err != nil {
		t.Error(err)
	}
	assert.Equal(t,
		[]string{"Ottilie90", "Jaden32", "Genevieve50"},
		funk.Map(body, func(x map[string]interface{}) string {
			return x["name"].(string)
		}),
	)

	// 多个文档过滤
	res2 := httptest.NewRecorder()
	where, err := jsoniter.Marshal(map[string]interface{}{
		"name": map[string]interface{}{"$in": []string{"Ottilie90", "Jaden32"}},
	})
	if err != nil {
		t.Error(err)
	}
	req2, _ := http.NewRequest("GET", fmt.Sprintf(`/members?where=%s`, where), nil)
	r.ServeHTTP(res2, req2)
	if err := jsoniter.Unmarshal(res2.Body.Bytes(), &body); err != nil {
		t.Error(err)
	}
	assert.Equal(t,
		[]string{"Ottilie90", "Jaden32"},
		funk.Map(body, func(x map[string]interface{}) string {
			return x["name"].(string)
		}),
	)

	// 多个文档ID
	res3 := httptest.NewRecorder()
	var ids []string
	for _, x := range body {
		ids = append(ids, fmt.Sprintf(`id=%s`, x["_id"].(primitive.ObjectID).Hex()))
	}
	idsQuery := strings.Join(ids, "&")
	req3, _ := http.NewRequest("GET", fmt.Sprintf(`/members?%s`, idsQuery), nil)
	r.ServeHTTP(res3, req3)
	if err := jsoniter.Unmarshal(res3.Body.Bytes(), &body); err != nil {
		t.Error(err)
	}
	assert.Equal(t,
		[]string{"Ottilie90", "Jaden32"},
		funk.Map(body, func(x map[string]interface{}) string {
			return x["name"].(string)
		}),
	)
}
