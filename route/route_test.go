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
	r.GET("/string", Returns(example.String))
	r.GET("/error", Returns(example.Error))
	r.GET("/error-custom", Returns(example.ErrorCustomCode))
	r.GET("/default", Returns(example.Default))
	r.GET("/default-bool", Returns(example.DefaultBool))
	r.GET("/empty", Returns(example.Empty))
	os.Exit(m.Run())
}

func TestReturnsString(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/string", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, res.Body.String(), `{"code":0,"message":"Hi there"}`)
}

func TestReturnsError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, res.Body.String(), `{"code":1,"message":"this is a test"}`)
}

func TestReturnsCustomCode(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error-custom", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, res.Body.String(), `{"code":100,"message":"this is a test"}`)
}

func TestReturnsDefault(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/default", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, res.Body.String(), `{"code":0,"data":{"status":"ok"},"message":"ok"}`)
}

func TestReturnsDefaultBool(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/default-bool", nil)
	r.ServeHTTP(res, req)
	t.Log(res.Body.String())
	assert.Equal(t, res.Body.String(), `{"code":0,"data":true,"message":"ok"}`)
}

func TestReturnsEmpty(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/empty", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, res.Code, 404)
}
