package route

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type Example struct{}

func (x *Example) String(c *gin.Context) interface{} {
	return "Hi there"
}

func (x *Example) Default(c *gin.Context) interface{} {
	return gin.H{
		"msg": "你好",
	}
}

func (x *Example) Create(c *gin.Context) interface{} {
	c.Set("status_code", 201)
	return gin.H{
		"count":   1,
		"hook-id": "xxx-xxx-xxx",
	}
}

func (x *Example) MockError(c *gin.Context) interface{} {
	return errors.New("模拟一个错误")
}

func (x *Example) MockErrorCustom(c *gin.Context) interface{} {
	c.Set("status_code", 401)
	c.Set("code", "AUTH_FAILED")
	return errors.New("模拟一个自定义错误")
}

func (x *Example) Empty(c *gin.Context) interface{} {
	return nil
}

func (x *Example) ModelName(c *gin.Context) interface{} {
	name, _ := c.Get("model")
	return gin.H{
		"name": name,
	}
}

var r *gin.Engine

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	r = gin.Default()
	example := new(Example)
	r.GET("/", Use(example.Default))
	r.POST("/", Use(example.Create))
	r.GET("/empty", Use(example.Empty))
	r.GET("error", Use(example.MockError))
	r.GET("error-custom", Use(example.MockErrorCustom))
	r.GET("model", Use(example.ModelName, SetModel("tests")))
	os.Exit(m.Run())
}

func TestRouteDefault(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	assert.Equal(t, `{"msg":"你好"}`, res.Body.String())
}

func TestRouteCreate(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 201, res.Code)
	assert.EqualValues(t, `{"count":1,"hook-id":"xxx-xxx-xxx"}`, res.Body.String())
}

func TestRouteEmpty(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/empty", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 204, res.Code)
	assert.Equal(t, 0, res.Body.Len())
}

func TestRouteError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 400, res.Code)
	assert.Equal(t, `{"code":"INVALID","message":"模拟一个错误"}`, res.Body.String())
}

func TestRouteErrorCustom(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error-custom", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 401, res.Code)
	assert.Equal(t, `{"code":"AUTH_FAILED","message":"模拟一个自定义错误"}`, res.Body.String())
}

func TestRouteModel(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/model", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code)
	assert.Equal(t, `{"name":"tests"}`, res.Body.String())
	t.Log(res.Body)
}
