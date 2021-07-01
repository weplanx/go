package bit

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
)

type Crud struct {
	tx     *gorm.DB
	model  interface{}
	orders Orders
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

func (x *Crud) setIdOrConditions(tx *gorm.DB, id interface{}, value Conditions) *gorm.DB {
	if id != nil {
		return tx.Where("id = ?", id)
	} else {
		return x.setConditions(tx, value)
	}
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

func (x *Crud) setOrders(tx *gorm.DB, orders Orders) *gorm.DB {
	for field, order := range x.orders {
		tx = tx.Order(field + " " + order)
	}
	for field, order := range orders {
		tx = tx.Order(field + " " + order)
	}
	return tx
}

func (x *Crud) setComplexVar(c *gin.Context, operator ...Operator) *complexVar {
	var v *complexVar
	if value, exists := c.Get(complexStart); exists {
		v = value.(*complexVar)
	} else {
		v = &complexVar{}
	}
	for _, operator := range operator {
		operator(v)
	}
	c.Set(complexComplete, v)
	return v
}

func (x *Crud) GetComplexVar(c *gin.Context) *complexVar {
	value, _ := c.Get(complexComplete)
	return value.(*complexVar)
}

type OriginListsBody struct {
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

func (x *Crud) OriginLists(c *gin.Context) interface{} {
	v := x.setComplexVar(c,
		SetBody(&OriginListsBody{}),
		SetData(reflect.New(reflect.SliceOf(reflect.TypeOf(x.model))).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetConditions() Conditions
		GetOrders() Orders
	})
	tx := x.tx.Model(x.model)
	if v.query != nil {
		tx = v.query(tx)
	}
	tx = x.setConditions(tx, body.GetConditions())
	tx = x.setOrders(tx, body.GetOrders())
	if err := tx.Find(v.data).Error; err != nil {
		return err
	}
	return v.data
}

type Pagination struct {
	Index int `json:"index" binding:"gt=0,number,required"`
	Limit int `json:"limit" binding:"gt=0,number,required"`
}

func (x Pagination) GetPagination() Pagination {
	return x
}

type ListsBody struct {
	Pagination `json:"page" binding:"required"`
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

func (x *Crud) Lists(c *gin.Context) interface{} {
	v := x.setComplexVar(c,
		SetBody(&ListsBody{}),
		SetData(reflect.New(reflect.SliceOf(reflect.TypeOf(x.model))).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetPagination() Pagination
		GetConditions() Conditions
		GetOrders() Orders
	})
	tx := x.tx.Model(x.model)
	if v.query != nil {
		tx = v.query(tx)
	}
	tx = x.setConditions(tx, body.GetConditions())
	tx = x.setOrders(tx, body.GetOrders())
	var total int64
	tx.Count(&total)
	page := body.GetPagination()
	tx = tx.Limit(page.Limit).Offset((page.Index - 1) * page.Limit)
	if err := tx.Find(v.data).Error; err != nil {
		return err
	}
	return gin.H{
		"lists": v.data,
		"total": total,
	}
}

type GetBody struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

func (x *GetBody) GetId() interface{} {
	return x.Id
}

func (x *Crud) Get(c *gin.Context) interface{} {
	v := x.setComplexVar(c,
		SetBody(&GetBody{}),
		SetData(reflect.New(reflect.TypeOf(x.model)).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetId() interface{}
		GetConditions() Conditions
		GetOrders() Orders
	})
	tx := x.tx.Model(x.model)
	if v.query != nil {
		tx = v.query(tx)
	}
	tx = x.setIdOrConditions(tx, body.GetId(), body.GetConditions())
	tx = x.setOrders(tx, body.GetOrders())
	if err := tx.First(v.data).Error; err != nil {
		return err
	}
	return v.data
}

func (x *Crud) Add(c *gin.Context) interface{} {
	v := x.setComplexVar(c)
	data := v.data
	if data == nil {
		v.Body = reflect.New(reflect.TypeOf(x.model)).Interface()
		if err := c.ShouldBindJSON(v.Body); err != nil {
			return err
		}
		data = v.Body
	}
	if err := x.tx.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(data).Error; err != nil {
			return
		}
		if v.txNext != nil {
			ID := reflect.ValueOf(data).Elem().FieldByName("ID").Interface()
			if err = v.txNext(tx, ID); err != nil {
				return
			}
		}
		return
	}); err != nil {
		return err
	}
	return "ok"
}

type EditBody struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
	Switch     *bool `json:"switch"`
}

func (x *EditBody) GetId() interface{} {
	return x.Id
}

func (x *Crud) Edit(c *gin.Context) interface{} {
	v := x.setComplexVar(c)
	var body interface{}
	data := v.data
	if data == nil {
		v.Body = reflect.New(reflect.StructOf([]reflect.StructField{
			{
				Name:      "EditBody",
				Type:      reflect.TypeOf(EditBody{}),
				Anonymous: true,
			},
			{
				Name:      "Updates",
				Type:      reflect.TypeOf(x.model),
				Anonymous: true,
			},
		})).Interface()
		if err := c.ShouldBindJSON(v.Body); err != nil {
			return err
		}
		elem := reflect.ValueOf(v.Body).Elem()
		body = elem.FieldByName("EditBody").Interface()
		data = elem.FieldByName("Updates").Interface()
	}
	tx := x.tx.Model(x.model)
	if v.query != nil {
		tx = v.query(tx)
	}
	if body != nil {
		b := body.(EditBody)
		tx = x.setIdOrConditions(tx, b.Id, b.Conditions)
	} else {
		b := v.Body.(interface {
			GetId() interface{}
			GetConditions() Conditions
		})
		tx = x.setIdOrConditions(tx, b.GetId(), b.GetConditions())
	}
	if err := tx.Transaction(func(txx *gorm.DB) (err error) {
		if err = txx.Updates(data).Error; err != nil {
			return
		}
		if v.txNext != nil {
			if err = v.txNext(txx, data); err != nil {
				return
			}
		}
		return
	}); err != nil {
		return err
	}
	return "ok"
}
