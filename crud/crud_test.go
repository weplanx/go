package crud

import (
	"bytes"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type UserController struct {
	*Crud
}

func TestCrud_Get(t *testing.T) {
	res1 := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"where": Conditions{
			{"name", "=", "Stuart"},
		},
	})
	req1, _ := http.NewRequest("POST", "/user/get", bytes.NewBuffer(body))
	r.ServeHTTP(res1, req1)
	assert.Equal(t,
		res1.Body.String(),
		`{"data":{"id":4,"path":"Stuart@VX.com","name":"Stuart","age":27,"gender":"Female","department":"Sale"},"error":0}`,
	)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/user/get", bytes.NewBuffer([]byte("text/plain")))
	r.ServeHTTP(res2, req2)
	assert.Equal(t,
		res2.Body.String(),
		`{"error":1,"msg":"invalid character 'e' in literal true (expecting 'r')"}`,
	)

	res3 := httptest.NewRecorder()
	body, _ = jsoniter.Marshal(&map[string]interface{}{
		"where": Conditions{
			{"number", "=", 100},
		},
	})
	req3, _ := http.NewRequest("POST", "/user/get", bytes.NewBuffer(body))
	r.ServeHTTP(res3, req3)
	assert.Equal(t,
		res3.Body.String(),
		`{"error":1,"msg":"Error 1054: Unknown column 'number' in 'where clause'"}`,
	)
}

func TestCrud_OriginLists(t *testing.T) {
	res1 := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"where": Conditions{
			{"id", "in", []int{1, 2, 3}},
		},
		"order": Orders{
			"id": "desc",
		},
	})
	req1, _ := http.NewRequest("POST", "/user/originLists", bytes.NewBuffer(body))
	r.ServeHTTP(res1, req1)
	assert.Equal(t,
		res1.Body.String(),
		`{"data":[{"id":3,"path":"Simone@VX.com","name":"Simone","age":23,"gender":"Male","department":"IT"},{"id":2,"path":"Questa@VX.com","name":"Questa","age":21,"gender":"Female","department":"IT"},{"id":1,"path":"Vandal@VX.com","name":"Vandal","age":25,"gender":"Male","department":"IT"}],"error":0}`,
	)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/user/originLists", bytes.NewBuffer([]byte("text/plain")))
	r.ServeHTTP(res2, req2)
	assert.Equal(t,
		res2.Body.String(),
		`{"error":1,"msg":"invalid character 'e' in literal true (expecting 'r')"}`,
	)

	res3 := httptest.NewRecorder()
	body, _ = jsoniter.Marshal(&map[string]interface{}{
		"where": Conditions{
			{"number", "=", 100},
		},
	})
	req3, _ := http.NewRequest("POST", "/user/originLists", bytes.NewBuffer(body))
	r.ServeHTTP(res3, req3)
	assert.Equal(t,
		res3.Body.String(),
		`{"error":1,"msg":"Error 1054: Unknown column 'number' in 'where clause'"}`,
	)
}

func TestCrud_Lists(t *testing.T) {
	res1 := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"page": Pagination{
			Index: 2,
			Limit: 5,
		},
	})
	req1, _ := http.NewRequest("POST", "/user/lists", bytes.NewBuffer(body))
	r.ServeHTTP(res1, req1)
	assert.Equal(t,
		res1.Body.String(),
		`{"data":{"lists":[{"id":6,"path":"Max@VX.com","name":"Max","age":28,"gender":"Female","department":"Designer"},{"id":7,"path":"Eagle-Eyed@VX.com","name":"Eagle-Eyed","age":31,"gender":"Male","department":"Support"},{"id":8,"path":"Marcia@VX.com","name":"Marcia","age":37,"gender":"Female","department":"Support"},{"id":9,"path":"Joanna@VX.com","name":"Joanna","age":40,"gender":"Male","department":"Manager"},{"id":10,"path":"Judy@VX.com","name":"Judy","age":50,"gender":"Female","department":"Manager"}],"total":20},"error":0}`,
	)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/user/lists", bytes.NewBuffer([]byte("text/plain")))
	r.ServeHTTP(res2, req2)
	assert.Equal(t,
		res2.Body.String(),
		`{"error":1,"msg":"invalid character 'e' in literal true (expecting 'r')"}`,
	)

	res3 := httptest.NewRecorder()
	body, _ = jsoniter.Marshal(&map[string]interface{}{
		"page": Pagination{
			Index: 2,
			Limit: 5,
		},
		"where": Conditions{
			{"number", "=", 100},
		},
	})
	req3, _ := http.NewRequest("POST", "/user/lists", bytes.NewBuffer(body))
	r.ServeHTTP(res3, req3)
	assert.Equal(t,
		res3.Body.String(),
		`{"error":1,"msg":"Error 1054: Unknown column 'number' in 'where clause'"}`,
	)
}

func TestCrud_Add(t *testing.T) {
	res1 := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&User{
		Email:      "Kain@VX.com",
		Name:       "Kain",
		Age:        27,
		Gender:     "Male",
		Department: "IT",
	})
	req1, _ := http.NewRequest("POST", "/user/add", bytes.NewBuffer(body))
	r.ServeHTTP(res1, req1)
	assert.Equal(t,
		res1.Body.String(),
		`{"error":0,"msg":"ok"}`,
	)
	var count int64
	err = db.Model(&User{}).Where("name = ?", "Kain").Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, count, int64(1))

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/user/add", bytes.NewBuffer([]byte("text/plain")))
	r.ServeHTTP(res2, req2)
	assert.Equal(t,
		res2.Body.String(),
		`{"error":1,"msg":"invalid character 'e' in literal true (expecting 'r')"}`,
	)

	res3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/user/add", bytes.NewBuffer(body))
	r.ServeHTTP(res3, req3)
	assert.Equal(t,
		res3.Body.String(),
		`{"error":1,"msg":"Error 1062: Duplicate entry 'Kain@VX.com' for key 'email'"}`,
	)
}

func TestCrud_Edit(t *testing.T) {
	res1 := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"where": Conditions{
			{"name", "=", "Questa"},
		},
		"updates": User{Age: 25},
	})
	req1, _ := http.NewRequest("POST", "/user/edit", bytes.NewBuffer(body))
	r.ServeHTTP(res1, req1)
	assert.Equal(t,
		res1.Body.String(),
		`{"error":0,"msg":"ok"}`,
	)
	var user User
	err = db.Where("name = ?", "Questa").First(&user).Error
	assert.Nil(t, err)
	assert.Equal(t, user.Age, 25)

	res2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/user/edit", bytes.NewBuffer([]byte("text/plain")))
	r.ServeHTTP(res2, req2)
	assert.Equal(t,
		res2.Body.String(),
		`{"error":1,"msg":"invalid character 'e' in literal true (expecting 'r')"}`,
	)

	res3 := httptest.NewRecorder()
	body, _ = jsoniter.Marshal(&map[string]interface{}{
		"where": Conditions{
			{"name", "=", "Questa"},
		},
		"updates": User{Email: "Vandal@VX.com"},
	})
	req3, _ := http.NewRequest("POST", "/user/edit", bytes.NewBuffer(body))
	r.ServeHTTP(res3, req3)
	assert.Equal(t,
		res3.Body.String(),
		`{"error":1,"msg":"Error 1062: Duplicate entry 'Vandal@VX.com' for key 'email'"}`,
	)
}

func TestCrud_Delete(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"where": Conditions{
			{"name", "=", "Questa"},
		},
	})
	req, _ := http.NewRequest("POST", "/user/delete", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"error":0,"msg":"ok"}`,
	)
	var count int64
	err = db.Model(&User{}).Where("name = ?", "Questa").Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, count, int64(0))
}
