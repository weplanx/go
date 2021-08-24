package mvc

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

func (x *Example) Error(c *gin.Context) interface{} {
	return errors.New("this is a test")
}

func (x *Example) ErrorCustomCode(c *gin.Context) interface{} {
	c.Set("code", 100)
	return errors.New("this is a test")
}

func (x *Example) Default(c *gin.Context) interface{} {
	return gin.H{
		"status": "ok",
	}
}

func (x *Example) DefaultBool(c *gin.Context) interface{} {
	return true
}

func (x *Example) Empty(c *gin.Context) interface{} {
	return nil
}

var r *gin.Engine

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	r = gin.Default()
	example := new(Example)
	r.GET("/string", Bind(example.String))
	r.GET("/error", Bind(example.Error))
	r.GET("/error-custom", Bind(example.ErrorCustomCode))
	r.GET("/default", Bind(example.Default))
	r.GET("/default-bool", Bind(example.DefaultBool))
	r.GET("/empty", Bind(example.Empty))
	os.Exit(m.Run())
}

func TestBindString(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/string", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, res.Body.String(), `{"error":0,"msg":"Hi there"}`)
}

func TestBindError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, res.Body.String(), `{"error":1,"msg":"this is a test"}`)
}

func TestBindCustomCode(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error-custom", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, res.Body.String(), `{"error":100,"msg":"this is a test"}`)
}

func TestBindDefault(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/default", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, res.Body.String(), `{"data":{"status":"ok"},"error":0}`)
}

func TestBindDefaultBool(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/default-bool", nil)
	r.ServeHTTP(res, req)
	t.Log(res.Body.String())
	assert.Equal(t, res.Body.String(), `{"data":true,"error":0}`)
}

func TestBindEmpty(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/empty", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, res.Code, 404)
}
