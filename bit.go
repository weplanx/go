package bit

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

// Bit a tools
type Bit struct {
	tx *gorm.DB
}

// Initialize initialize tool set
//	Params:
//	 db: *gorm.DB
func Initialize(db *gorm.DB) *Bit {
	return &Bit{tx: db}
}

// CrudOption crud option
type CrudOption func(*Crud)

// SetOrders set global ordering
//	Params:
//	 orders: map[string]string
func SetOrders(orders Orders) CrudOption {
	return func(option *Crud) {
		option.orders = orders
	}
}

// Crud create CRUD functions
//	Params:
//	 model: model name
//	 options: crud option
func (x *Bit) Crud(model interface{}, options ...CrudOption) *Crud {
	crud := &Crud{
		tx:    x.tx,
		model: model,
	}
	for _, apply := range options {
		apply(crud)
	}
	return crud
}

// Bind The binding routing function returns the result uniformly
//	Params:
//	 handlerFn: func(c *gin.Context) interface{}
//	Mapped return:
//	 (string) => 200 {"error":0,"msg":<this string value>}
//	 (error) => 200 {"error":1,"msg":<this error msg, for example err.Error()>}
//	 (default) => 200 {"error":0,"data":<interface{}>}
//	Custom error code: c.Set("code", 1000)
func Bind(handlerFn interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if fn, ok := handlerFn.(func(c *gin.Context) interface{}); ok {
			switch result := fn(c).(type) {
			case string:
				c.JSON(http.StatusOK, gin.H{
					"error": 0,
					"msg":   result,
				})
				break
			case error:
				code, exists := c.Get("code")
				if !exists {
					code = 1
				}
				c.JSON(http.StatusOK, gin.H{
					"error": code,
					"msg":   result.Error(),
				})
				break
			default:
				if result != nil {
					c.JSON(http.StatusOK, gin.H{
						"error": 0,
						"data":  result,
					})
				} else {
					c.Status(http.StatusNotFound)
				}
			}
		}
	}
}
