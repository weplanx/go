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

// GetConditions get array condition definition
func (c Conditions) GetConditions() Conditions {
	return c
}

// Orders order definition
type Orders map[string]string

// GetOrders get order definition
func (c Orders) GetOrders() Orders {
	return c
}

// Set ID or array conditions
func (x *Crud) setIdOrConditions(tx *gorm.DB, id interface{}, value Conditions) *gorm.DB {
	if id != nil {
		return tx.Where("id = ?", id)
	} else {
		return x.setConditions(tx, value)
	}
}

// Set array conditions
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

// Set orders
func (x *Crud) setOrders(tx *gorm.DB, orders Orders) *gorm.DB {
	for field, order := range x.orders {
		tx = tx.Order(field + " " + order)
	}
	for field, order := range orders {
		tx = tx.Order(field + " " + order)
	}
	return tx
}

// Set mixed
func (x *Crud) setMixed(c *gin.Context, operator ...Operator) *mixed {
	var v *mixed
	if value, exists := c.Get(mixedStart); exists {
		v = value.(*mixed)
	} else {
		v = &mixed{}
	}
	for _, operator := range operator {
		operator(v)
	}
	c.Set(mixedComplete, v)
	return v
}

// GetMixed get mixed vars
func (x *Crud) GetMixed(c *gin.Context) *mixed {
	value, _ := c.Get(mixedComplete)
	return value.(*mixed)
}

// OriginListsBody general definition of list body
type OriginListsBody struct {
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// OriginLists general definition of list request
func (x *Crud) OriginLists(c *gin.Context) interface{} {
	v := x.setMixed(c,
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

// Pagination page properties
//	Index: page number
//	Limit: number of pages
type Pagination struct {
	Index int `json:"index" binding:"gt=0,number,required"`
	Limit int `json:"limit" binding:"gt=0,number,required"`
}

// GetPagination get page properties
func (x Pagination) GetPagination() Pagination {
	return x
}

// ListsBody general definition of page body
type ListsBody struct {
	Pagination `json:"page" binding:"required"`
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// Lists general definition of page data
func (x *Crud) Lists(c *gin.Context) interface{} {
	v := x.setMixed(c,
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

// GetBody general definition of get body
type GetBody struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// GetId get id
func (x *GetBody) GetId() interface{} {
	return x.Id
}

// Get general definition of get data request
func (x *Crud) Get(c *gin.Context) interface{} {
	v := x.setMixed(c,
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

// Add general definition of create data request
func (x *Crud) Add(c *gin.Context) interface{} {
	v := x.setMixed(c)
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

// EditBody general definition of edit body
type EditBody struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
	Switch     *bool `json:"switch"`
}

// GetId get id
func (x *EditBody) GetId() interface{} {
	return x.Id
}

// Edit general definition of edit data request
func (x *Crud) Edit(c *gin.Context) interface{} {
	v := x.setMixed(c)
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

// DeleteBody general definition of delete body
type DeleteBody struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
}

// GetId get id
func (x *DeleteBody) GetId() interface{} {
	return x.Id
}

// Delete general definition of delete data request
func (x *Crud) Delete(c *gin.Context) interface{} {
	v := x.setMixed(c,
		SetBody(&DeleteBody{}),
		SetData(reflect.New(reflect.TypeOf(x.model)).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetId() interface{}
		GetConditions() Conditions
	})
	tx := x.tx.Model(x.model)
	if v.query != nil {
		tx = v.query(tx)
	}
	id := body.GetId()
	if id != nil {
		tx = tx.Where("id in (?)", id)
	} else {
		tx = x.setConditions(tx, body.GetConditions())
	}
	if err := tx.Transaction(func(txx *gorm.DB) (err error) {
		if err = txx.Delete(v.data).Error; err != nil {
			return
		}
		if v.txNext != nil {
			if err = v.txNext(txx); err != nil {
				return
			}
		}
		return
	}); err != nil {
		return err
	}
	return "ok"
}
