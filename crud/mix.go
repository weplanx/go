package crud

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	variables = "mix.vars"
)

type mixed struct {
	Body   interface{}
	data   interface{}
	query  func(tx *gorm.DB) *gorm.DB
	txNext func(tx *gorm.DB, args ...interface{}) error
}

// Set default initial mix
func (x *API) mixed(c *gin.Context, operator ...Operator) *mixed {
	v := new(mixed)
	for _, operator := range operator {
		operator(v)
	}
	if value, exists := c.Get(variables); exists {
		mix := value.(*mixed)
		if mix.Body != nil {
			v.Body = mix.Body
		}
		if mix.data != nil {
			v.data = mix.data
		}
		if mix.query != nil {
			v.query = mix.query
		}
		if mix.txNext != nil {
			v.txNext = mix.txNext
		}
	}
	return v
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
//  description:
// 	 OriginLists,Lists,Get: Basic condition
func Query(fn func(tx *gorm.DB) *gorm.DB) Operator {
	return func(c *mixed) {
		c.query = fn
	}
}

// TxNext set the data operations included in the transaction
//  description:
// 	 Add,Edit,Delete
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
	c.Set(variables, x)
}
