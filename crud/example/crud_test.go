package example

import (
	"bytes"
	jsoniter "github.com/json-iterator/go"
	"github.com/kainonly/go-bit/crud"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type UserController struct {
	*crud.Crud
}

func TestCrud_FindOne(t *testing.T) {
	res1 := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"where": crud.Conditions{
			{"name", "=", "Stuart"},
		},
	})
	req1, _ := http.NewRequest("POST", "/user/r/find/one", bytes.NewBuffer(body))
	r.ServeHTTP(res1, req1)
	assert.Equal(t,
		res1.Body.String(),
		`{"data":{"id":4,"path":"Stuart@VX.com","name":"Stuart","age":27,"gender":"Female","department":"Sale"},"error":0}`,
	)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/user/r/find/one", bytes.NewBuffer([]byte("text/plain")))
	r.ServeHTTP(res2, req2)
	assert.Equal(t,
		res2.Body.String(),
		`{"error":1,"msg":"invalid character 'e' in literal true (expecting 'r')"}`,
	)

	res3 := httptest.NewRecorder()
	body, _ = jsoniter.Marshal(&map[string]interface{}{
		"where": crud.Conditions{
			{"number", "=", 100},
		},
	})
	req3, _ := http.NewRequest("POST", "/user/r/find/one", bytes.NewBuffer(body))
	r.ServeHTTP(res3, req3)
	assert.Equal(t,
		res3.Body.String(),
		`{"error":1,"msg":"ERROR: column \"number\" does not exist (SQLSTATE 42703)"}`,
	)
}

func TestCrud_FindMany(t *testing.T) {
	res1 := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"where": crud.Conditions{
			{"id", "in", []int{1, 2, 3}},
		},
		"order": crud.Orders{
			"id": "desc",
		},
	})
	req1, _ := http.NewRequest("POST", "/user/r/find/many", bytes.NewBuffer(body))
	r.ServeHTTP(res1, req1)
	assert.Equal(t,
		res1.Body.String(),
		`{"data":[{"id":3,"path":"Simone@VX.com","name":"Simone","age":23,"gender":"Male","department":"IT"},{"id":2,"path":"Questa@VX.com","name":"Questa","age":21,"gender":"Female","department":"IT"},{"id":1,"path":"Vandal@VX.com","name":"Vandal","age":25,"gender":"Male","department":"IT"}],"error":0}`,
	)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/user/r/find/many", bytes.NewBuffer([]byte("text/plain")))
	r.ServeHTTP(res2, req2)
	assert.Equal(t,
		res2.Body.String(),
		`{"error":1,"msg":"invalid character 'e' in literal true (expecting 'r')"}`,
	)

	res3 := httptest.NewRecorder()
	body, _ = jsoniter.Marshal(&map[string]interface{}{
		"where": crud.Conditions{
			{"number", "=", 100},
		},
	})
	req3, _ := http.NewRequest("POST", "/user/r/find/many", bytes.NewBuffer(body))
	r.ServeHTTP(res3, req3)
	assert.Equal(t,
		res3.Body.String(),
		`{"error":1,"msg":"ERROR: column \"number\" does not exist (SQLSTATE 42703)"}`,
	)
}

func TestCrud_FindPage(t *testing.T) {
	res1 := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"page": crud.Pagination{
			Index: 2,
			Limit: 5,
		},
	})
	req1, _ := http.NewRequest("POST", "/user/r/find/page", bytes.NewBuffer(body))
	r.ServeHTTP(res1, req1)
	assert.Equal(t,
		res1.Body.String(),
		`{"data":{"lists":[{"id":6,"path":"Max@VX.com","name":"Max","age":28,"gender":"Female","department":"Designer"},{"id":7,"path":"Eagle-Eyed@VX.com","name":"Eagle-Eyed","age":31,"gender":"Male","department":"Support"},{"id":8,"path":"Marcia@VX.com","name":"Marcia","age":37,"gender":"Female","department":"Support"},{"id":9,"path":"Joanna@VX.com","name":"Joanna","age":40,"gender":"Male","department":"Manager"},{"id":10,"path":"Judy@VX.com","name":"Judy","age":50,"gender":"Female","department":"Manager"}],"total":20},"error":0}`,
	)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/user/r/find/page", bytes.NewBuffer([]byte("text/plain")))
	r.ServeHTTP(res2, req2)
	assert.Equal(t,
		res2.Body.String(),
		`{"error":1,"msg":"invalid character 'e' in literal true (expecting 'r')"}`,
	)

	res3 := httptest.NewRecorder()
	body, _ = jsoniter.Marshal(&map[string]interface{}{
		"page": crud.Pagination{
			Index: 2,
			Limit: 5,
		},
		"where": crud.Conditions{
			{"number", "=", 100},
		},
	})
	req3, _ := http.NewRequest("POST", "/user/r/find/page", bytes.NewBuffer(body))
	r.ServeHTTP(res3, req3)
	assert.Equal(t,
		res3.Body.String(),
		`{"error":1,"msg":"ERROR: column \"number\" does not exist (SQLSTATE 42703)"}`,
	)
}

func TestCrud_Create(t *testing.T) {
	res1 := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&Example{
		Email:      "Kain@VX.com",
		Name:       "Kain",
		Age:        27,
		Gender:     "Male",
		Department: "IT",
	})
	req1, _ := http.NewRequest("POST", "/user/w/create", bytes.NewBuffer(body))
	r.ServeHTTP(res1, req1)
	assert.Equal(t,
		res1.Body.String(),
		`{"error":0,"msg":"ok"}`,
	)
	var count int64
	err = db.Model(&Example{}).Where("name = ?", "Kain").Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, count, int64(1))

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/user/w/create", bytes.NewBuffer([]byte("text/plain")))
	r.ServeHTTP(res2, req2)
	assert.Equal(t,
		res2.Body.String(),
		`{"error":1,"msg":"invalid character 'e' in literal true (expecting 'r')"}`,
	)

	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/user/w/create", bytes.NewBuffer(body))
	r.ServeHTTP(res3, req3)
	assert.Equal(t,
		res3.Body.String(),
		`{"error":1,"msg":"ERROR: duplicate key value violates unique constraint \"examples_email_key\" (SQLSTATE 23505)"}`,
	)
}

func TestCrud_Update(t *testing.T) {
	res1 := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"where": crud.Conditions{
			{"name", "=", "Questa"},
		},
		"updates": Example{Age: 25},
	})
	req1, _ := http.NewRequest("POST", "/user/w/update", bytes.NewBuffer(body))
	r.ServeHTTP(res1, req1)
	assert.Equal(t,
		res1.Body.String(),
		`{"error":0,"msg":"ok"}`,
	)
	var user Example
	err = db.Where("name = ?", "Questa").First(&user).Error
	assert.Nil(t, err)
	assert.Equal(t, user.Age, 25)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/user/w/update", bytes.NewBuffer([]byte("text/plain")))
	r.ServeHTTP(res2, req2)
	assert.Equal(t,
		res2.Body.String(),
		`{"error":1,"msg":"invalid character 'e' in literal true (expecting 'r')"}`,
	)

	res3 := httptest.NewRecorder()
	body, _ = jsoniter.Marshal(&map[string]interface{}{
		"where": crud.Conditions{
			{"name", "=", "Questa"},
		},
		"updates": Example{Email: "Vandal@VX.com"},
	})
	req3, _ := http.NewRequest("POST", "/user/w/update", bytes.NewBuffer(body))
	r.ServeHTTP(res3, req3)
	assert.Equal(t,
		res3.Body.String(),
		`{"error":1,"msg":"ERROR: duplicate key value violates unique constraint \"examples_email_key\" (SQLSTATE 23505)"}`,
	)
}

func TestCrud_Delete(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"where": crud.Conditions{
			{"name", "=", "Questa"},
		},
	})
	req, _ := http.NewRequest("POST", "/user/w/delete", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"error":0,"msg":"ok"}`,
	)
	var count int64
	err = db.Model(&Example{}).Where("name = ?", "Questa").Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, count, int64(0))
}
