package bit

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Mixed context keys
const mixedStart = "mixed.start"
const mixedComplete = "mixed.complete"

// Mixed Vars
type mixed struct {
	Body   interface{}
	data   interface{}
	query  func(tx *gorm.DB) *gorm.DB
	txNext func(tx *gorm.DB, args ...interface{}) error
}

// Operator mixed operations
type Operator func(*mixed)

// SetBody define the custom request body
//	Scenes:
//	 OriginLists,Lists,Get,Delete: they need to embed OriginListsBody, ListsBody, GetBody, DeleteBody
//	 Add,Edit: useless
func SetBody(body interface{}) Operator {
	return func(c *mixed) {
		if c.Body == nil {
			c.Body = body
		}
	}
}

// SetData data with different meanings
//	Scenes:
//	 OriginLists,Lists,Get: the specified gorm model, used to query and return data
//	 Add: used to create custom data
//	 Edit: used to customize update data
//	 Delete: useless
func SetData(data interface{}) Operator {
	return func(c *mixed) {
		if c.data == nil {
			c.data = data
		}
	}
}

// Query query custom extension
//	Params:
//	 fn: func(tx *gorm.DB) *gorm.DB
func Query(fn func(tx *gorm.DB) *gorm.DB) Operator {
	return func(c *mixed) {
		c.query = fn
	}
}

// TxNext transaction custom extension
//	Params:
//	 fn: func(tx *gorm.DB, args ...interface{}) error
func TxNext(fn func(tx *gorm.DB, args ...interface{}) error) Operator {
	return func(c *mixed) {
		c.txNext = fn
	}
}

// Mixed define mixed operations
//	Params:
//	 c: *gin.Context
//	 operator: mixed operations
func Mixed(c *gin.Context, operator ...Operator) {
	v := new(mixed)
	for _, operator := range operator {
		operator(v)
	}
	c.Set(mixedStart, v)
}
