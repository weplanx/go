package crud

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"reflect"
)

// FirstBody Get a single resource request body
type FirstBody struct {
	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

type FirstOption struct {
	Query func(tx *gorm.DB) *gorm.DB
}

func NewFirstOption(c *gin.Context) *FirstOption {
	option := new(FirstOption)
	c.Set("option", option)
	return option
}

func (x *FirstOption) SetQuery(fn func(tx *gorm.DB) *gorm.DB) *FirstOption {
	x.Query = fn
	return x
}

// First Get a single resource
func (x *API) First(c *gin.Context) interface{} {
	var option FirstOption
	if value, exists := c.Get("option"); exists {
		option = value.(FirstOption)
	}
	body := FirstBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	data := reflect.New(reflect.TypeOf(x.Model)).Interface()
	tx := x.Tx.WithContext(c)
	if option.Query != nil {
		tx = option.Query(tx)
	}
	tx = x.where(tx, body.GetConditions())
	tx = x.orderBy(tx, body.GetOrders())
	if err := tx.First(data).Error; err != nil {
		return err
	}
	return data
}
