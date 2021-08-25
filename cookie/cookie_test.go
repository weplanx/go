package cookie

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var x *Cookie
var r *gin.Engine

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	x = New(Option{
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
	}, http.SameSiteLaxMode)
	r = gin.Default()
	r.GET("/", func(c *gin.Context) {
		value, _ := x.Get(c, "name")
		c.String(200, value)
	})
	r.POST("/", func(c *gin.Context) {
		x.Set(c, "name", "kain")
		c.Status(http.StatusNoContent)
	})
	r.DELETE("/", func(c *gin.Context) {
		x.Del(c, "name")
		value, _ := x.Get(c, "name")
		c.String(200, value)
	})
}

var cookie string

func TestCookie_Set(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", nil)
	r.ServeHTTP(res, req)
	assert.NotEmpty(t, res.Result().Cookies())
	c := res.Result().Cookies()[0]
	assert.Equal(t, c.Name, "name")
	assert.Equal(t, c.Value, "kain")
	assert.Equal(t, c.Path, "/")
	assert.Equal(t, c.MaxAge, 3600)
	assert.Equal(t, c.Secure, true)
	assert.Equal(t, c.HttpOnly, true)
	cookie = c.Raw
}

func TestCookie_Get(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(res, req)
	assert.Equal(t, res.Body.String(), "kain")
}

func TestCookie_Del(t *testing.T) {
	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("DELETE", "/", nil)
	req1.Header.Set("Cookie", cookie)
	r.ServeHTTP(res1, req1)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/", nil)
	req2.Header.Set("Cookie", res1.Header().Get("Set-Cookie"))
	r.ServeHTTP(res2, req2)
	assert.Equal(t, res2.Body.String(), "")
}
