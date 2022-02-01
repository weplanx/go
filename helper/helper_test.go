package helper

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/route"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUuid(t *testing.T) {
	u, err := uuid.Parse(Uuid())
	if err != nil {
		t.Error(err)
	}
	t.Log(u)
}

type Example struct{}

func (x *Example) ObjectIdValidate(c *gin.Context) interface{} {
	var params struct {
		Id string `uri:"id" binding:"objectId"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		return err
	}
	return nil
}

func (x *Example) KeyValidate(c *gin.Context) interface{} {
	var query struct {
		Key string `form:"key" binding:"key"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		return err
	}
	return nil
}

func (x *Example) SortValidate(c *gin.Context) interface{} {
	var query struct {
		Sort string `form:"sort" binding:"sort"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		return err
	}
	return nil
}

var r *gin.Engine

func BeforeServe() {
	gin.SetMode(gin.TestMode)
	r = gin.Default()
	ExtendValidation()
	example := new(Example)
	r.GET("/xxx/:id", route.Use(example.ObjectIdValidate))
	r.GET("/xxx/exists", route.Use(example.KeyValidate))
	r.GET("/xxx", route.Use(example.SortValidate))
}

func TestObjectIdValidate(t *testing.T) {
	BeforeServe()
	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/xxx/abcd", nil)
	r.ServeHTTP(res1, req1)
	assert.Equal(t, 400, res1.Code)
	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/xxx/61f8c094ccb6315051abfba4", nil)
	r.ServeHTTP(res2, req2)
	assert.Equal(t, 204, res2.Code)
}

func TestKeyValidate(t *testing.T) {
	BeforeServe()
	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/xxx/exists?key=xas32", nil)
	r.ServeHTTP(res1, req1)
	assert.Equal(t, 400, res1.Code)
	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/xxx/exists?key=xasAA", nil)
	r.ServeHTTP(res2, req2)
	assert.Equal(t, 400, res2.Code)
	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/xxx/exists?key=xasaa", nil)
	r.ServeHTTP(res3, req3)
	assert.Equal(t, 204, res3.Code)
}

func TestSortValidate(t *testing.T) {
	BeforeServe()
	res1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/xxx?sort=age.1", nil)
	r.ServeHTTP(res1, req1)
	assert.Equal(t, 204, res1.Code)
	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/xxx?sort=age.-1", nil)
	r.ServeHTTP(res2, req2)
	assert.Equal(t, 204, res2.Code)
	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/xxx?sort=age_1", nil)
	r.ServeHTTP(res3, req3)
	assert.Equal(t, 400, res3.Code)
	res4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/xxx?sort=age.2", nil)
	r.ServeHTTP(res4, req4)
	assert.Equal(t, 400, res4.Code)
	res5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("GET", "/xxx?sort=age.x", nil)
	r.ServeHTTP(res5, req5)
	assert.Equal(t, 400, res5.Code)
}
