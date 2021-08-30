package crud

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type UserMixController struct {
	*Crud
}

func (x *UserMixController) Get(c *gin.Context) interface{} {
	var body struct {
		GetBody
		Name string `json:"name"`
	}
	Mix(c,
		SetBody(&body),
		Query(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("name = ?", body.Name)
			return tx
		}),
	)
	return x.Crud.Get(c)
}

func TestMixGet(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"name": "Marcia",
	})
	req, _ := http.NewRequest("POST", "/user-mix/get", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"data":{"id":8,"path":"Marcia@VX.com","name":"Marcia","age":37,"gender":"Female","department":"Support"},"error":0}`,
	)
}

func (x *UserMixController) OriginLists(c *gin.Context) interface{} {
	Mix(c,
		Query(func(tx *gorm.DB) *gorm.DB {
			tx.Where("id in ?", []uint64{5, 6})
			return tx
		}),
	)
	return x.Crud.OriginLists(c)
}

func TestMixOriginLists(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"order": Orders{
			"id": "desc",
		},
	})
	req, _ := http.NewRequest("POST", "/user-mix/originLists", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"data":[{"id":6,"path":"Max@VX.com","name":"Max","age":28,"gender":"Female","department":"Designer"},{"id":5,"path":"Vivianne@VX.com","name":"Vivianne","age":36,"gender":"Male","department":"Sale"}],"error":0}`,
	)
}

func (x *UserMixController) Add(c *gin.Context) interface{} {
	Mix(c,
		TxNext(func(tx *gorm.DB, args ...interface{}) error {
			log.Println(args[0].(*User))
			return errors.New("an abnormal rollback occurred")
		}),
	)
	return x.Crud.Add(c)
}

func TestMixAdd(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&User{
		Email:      "Zhang@VX.com",
		Name:       "Zhang",
		Age:        27,
		Gender:     "Male",
		Department: "IT",
	})
	req, _ := http.NewRequest("POST", "/user-mix/add", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"error":1,"msg":"an abnormal rollback occurred"}`,
	)
	var count int64
	err = db.Model(&User{}).Where("name = ?", "Zhang").Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, count, int64(0))
}

func (x *UserMixController) Edit(c *gin.Context) interface{} {
	var body struct {
		EditBody
		Name string `json:"name"`
	}
	Mix(c,
		SetBody(&body),
		Query(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("name = ?", body.Name)
			return tx
		}),
		TxNext(func(tx *gorm.DB, args ...interface{}) error {
			log.Println(args[0].(*User))
			return errors.New("an abnormal rollback occurred")
		}),
	)
	return x.Crud.Edit(c)
}

func TestMixEdit(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(map[string]interface{}{
		"name":    "Stuart",
		"updates": User{Age: 25},
	})
	req, _ := http.NewRequest("POST", "/user-mix/edit", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"error":1,"msg":"an abnormal rollback occurred"}`,
	)
	var data User
	err = db.Where("name = ?", "Stuart").First(&data).Error
	assert.Nil(t, err)
	assert.Equal(t, data.Age, 27)
}

func (x *UserMixController) Delete(c *gin.Context) interface{} {
	var body struct {
		DeleteBody
		Name string `json:"name"`
	}
	Mix(c,
		SetBody(&body),
		Query(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("name = ?", body.Name)
			return tx
		}),
		TxNext(func(tx *gorm.DB, args ...interface{}) error {
			log.Println(args[0])
			return errors.New("an abnormal rollback occurred")
		}),
	)
	return x.Crud.Delete(c)
}

func TestMixDelete(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(map[string]interface{}{
		"name": "Joanna",
	})
	req, _ := http.NewRequest("POST", "/user-mix/delete", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"error":1,"msg":"an abnormal rollback occurred"}`,
	)
	var count int64
	err = db.Model(&User{}).Where("name = ?", "Joanna").Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, count, int64(1))
}
