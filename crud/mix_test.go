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

func (x *UserMixController) Add(c *gin.Context) interface{} {
	Mix(c,
		TxNext(func(tx *gorm.DB, args ...interface{}) error {
			log.Println(args[0].(*User))
			return errors.New("an abnormal rollback occurred")
		}),
	)
	return x.Crud.Add(c)
}

func TestTxNext(t *testing.T) {
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

func (x *UserMixController) OriginLists(c *gin.Context) interface{} {
	Mix(c,
		Query(func(tx *gorm.DB) *gorm.DB {
			tx.Where("id in ?", []uint64{5, 6})
			return tx
		}),
	)
	return x.Crud.OriginLists(c)
}

func TestQuery(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/user-mix/originLists", bytes.NewBuffer([]byte(`{}`)))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"data":[{"id":5,"path":"Vivianne@VX.com","name":"Vivianne","age":36,"gender":"Male","department":"Sale"},{"id":6,"path":"Max@VX.com","name":"Max","age":28,"gender":"Female","department":"Designer"}],"error":0}`,
	)
}
