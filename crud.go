package bit

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
)

type Crud struct {
	tx      *gorm.DB
	model   interface{}
	orderBy []string
}

// Conditions array condition definition
type Conditions [][]interface{}

func (c Conditions) GetConditions() Conditions {
	return c
}

// Orders definition
type Orders map[string]string

func (c Orders) GetOrders() Orders {
	return c
}

type apis interface {
	GetId() interface{}
	GetConditions() Conditions
	GetOrders() Orders
}

type GetBody struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

func (x *GetBody) GetId() interface{} {
	return x.Id
}

func (x *Crud) setIdOrConditions(tx *gorm.DB, id interface{}, value Conditions) *gorm.DB {
	if id != nil {
		tx = tx.Where("id = ?", id)
	} else {
		tx = x.setConditions(tx, value)
	}
	return tx
}

func (x *Crud) setConditions(tx *gorm.DB, conditions Conditions) *gorm.DB {
	for _, condition := range conditions {
		tx = tx.Where(
			"? "+condition[1].(string)+" ?",
			clause.Column{Name: condition[0].(string)},
			condition[2],
		)
	}
	return tx
}

func (x *Crud) getComplexVar(c *gin.Context, body interface{}) *complexVar {
	var v *complexVar
	if value, exists := c.Get(context); exists {
		v = value.(*complexVar)
	} else {
		v = &complexVar{}
	}
	if v.data == nil {
		v.data = reflect.New(reflect.TypeOf(x.model)).Interface()
	}
	if v.body == nil {
		v.body = body
	}
	return v
}

func (x *Crud) Get(c *gin.Context) interface{} {
	v := x.getComplexVar(c, &GetBody{})
	if err := c.ShouldBindJSON(v.body); err != nil {
		return err
	}
	body := v.body.(apis)
	tx := x.tx.Model(x.model)
	tx = x.setIdOrConditions(tx, body.GetId(), body.GetConditions())
	tx.First(v.data)
	return v.data
}
