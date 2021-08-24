package crud

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	MixStart    = "mix.start"
	MixComplete = "mix.complete"
)

type mixed struct {
	Body   interface{}
	data   interface{}
	query  func(tx *gorm.DB) *gorm.DB
	txNext func(tx *gorm.DB, args ...interface{}) error
}

type Operator func(*mixed)

// SetBody custom request body
//	description:
//	 OriginLists,Lists,Get,Delete: Need embedded structure OriginListsBody, ListsBody, GetBody, DeleteBody
func SetBody(body interface{}) Operator {
	return func(c *mixed) {
		if c.Body == nil {
			c.Body = body
		}
	}
}

// SetData custom data
//	description:
//	 OriginLists,Lists,Get: Set gorm model for query and final data return
//	 Add: Custom create data
//	 Edit: Custom update data
func SetData(data interface{}) Operator {
	return func(c *mixed) {
		if c.data == nil {
			c.data = data
		}
	}
}

// Query custom query
func Query(fn func(tx *gorm.DB) *gorm.DB) Operator {
	return func(c *mixed) {
		c.query = fn
	}
}

// TxNext set the data operations included in the transaction
func TxNext(fn func(tx *gorm.DB, args ...interface{}) error) Operator {
	return func(c *mixed) {
		c.txNext = fn
	}
}

// Mix define mixed operations
func Mix(c *gin.Context, operator ...Operator) {
	x := new(mixed)
	for _, operator := range operator {
		operator(x)
	}
	c.Set(MixStart, x)
}
