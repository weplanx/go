package bit

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const complexStart = "complex.start"
const complexComplete = "complex.complete"

type complexVar struct {
	Body   interface{}
	data   interface{}
	query  func(tx *gorm.DB) *gorm.DB
	txNext func(tx *gorm.DB, args ...interface{}) error
}

type Operator func(*complexVar)

func SetBody(body interface{}) Operator {
	return func(c *complexVar) {
		if c.Body == nil {
			c.Body = body
		}
	}
}

func SetData(data interface{}) Operator {
	return func(c *complexVar) {
		if c.data == nil {
			c.data = data
		}
	}
}

func Query(fn func(tx *gorm.DB) *gorm.DB) Operator {
	return func(c *complexVar) {
		c.query = fn
	}
}

func TxNext(fn func(tx *gorm.DB, args ...interface{}) error) Operator {
	return func(c *complexVar) {
		c.txNext = fn
	}
}

func Complex(c *gin.Context, operator ...Operator) {
	v := new(complexVar)
	for _, operator := range operator {
		operator(v)
	}
	c.Set(complexStart, v)
}
