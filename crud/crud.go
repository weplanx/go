package crud

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"reflect"
)

type Crud struct {
	Db *gorm.DB
}

func New(db *gorm.DB) *Crud {
	return &Crud{Db: db}
}

// API create resource operation
func (x *Crud) API(model interface{}) *API {
	return &API{
		Model: model,
		Tx:    x.Db.Model(model),
	}
}

type API struct {
	Model interface{}
	Tx    *gorm.DB
	Operations
}

type Operations interface {
	First(c *gin.Context) interface{}
	Find(c *gin.Context) interface{}
	Page(c *gin.Context) interface{}
	Create(c *gin.Context) interface{}
	Update(c *gin.Context) interface{}
	Delete(c *gin.Context) interface{}
}

// Conditions conditions array
type Conditions [][3]interface{}

func (c Conditions) GetConditions() Conditions {
	return c
}

// Orders sort fields
type Orders map[string]string

func (c Orders) GetOrders() Orders {
	return c
}

// where generate request query conditions
func (x *API) where(tx *gorm.DB, conds Conditions) *gorm.DB {
	for _, v := range conds {
		tx = tx.Where(gorm.Expr(v[0].(string)+" "+v[1].(string)+" ?", v[2]))
	}
	return tx
}

// orderBy generate request ordering rules
func (x *API) orderBy(tx *gorm.DB, orders Orders) *gorm.DB {
	for k, v := range orders {
		tx = tx.Order(k + " " + v)
	}
	return tx
}

// FindBody Get the original list resource request body
type FindBody struct {
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// Find Get the original list resource
func (x *API) Find(c *gin.Context) interface{} {
	var body FindBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	v := x.mixed(c,
		SetData(reflect.New(reflect.SliceOf(reflect.TypeOf(x.Model))).Interface()),
	)
	tx := x.Tx.WithContext(c)
	if v.query != nil {
		tx = v.query(tx)
	}
	tx = x.where(tx, body.Conditions)
	tx = x.orderBy(tx, body.Orders)
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

// PageBody Get the request body of the paged list resource
type PageBody struct {
	Pagination `json:"page" binding:"required"`
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// Page Get paging list resources
func (x *API) Page(c *gin.Context) interface{} {
	var body PageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	v := x.mixed(c,
		SetData(reflect.New(reflect.SliceOf(reflect.TypeOf(x.Model))).Interface()),
	)
	tx := x.Tx.WithContext(c)
	if v.query != nil {
		tx = v.query(tx)
	}
	tx = x.where(tx, body.Conditions)
	tx = x.orderBy(tx, body.Orders)
	var total int64
	tx.Count(&total)
	p := body.Pagination
	tx = tx.Limit(p.Limit).Offset((p.Index - 1) * p.Limit)
	if err := tx.Find(v.data).Error; err != nil {
		return err
	}
	return gin.H{
		"lists": v.data,
		"total": total,
	}
}

// Create resources
func (x *API) Create(c *gin.Context) interface{} {
	v := x.mixed(c,
		SetBody(reflect.New(reflect.TypeOf(x.Model))),
	)
	if err := c.ShouldBindJSON(&v.Body); err != nil {
		return err
	}
	if err := x.Tx.WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(v.Body).Error; err != nil {
			return
		}
		if v.txNext != nil {
			if err = v.txNext(tx, v.Body); err != nil {
				return
			}
		}
		return
	}); err != nil {
		return err
	}
	return "ok"
}

// UpdateBody Update resource request body
type UpdateBody struct {
	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
	Updates    interface{} `json:"updates" binding:"required"`
}

// Update resources
func (x *API) Update(c *gin.Context) interface{} {
	var body UpdateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	v := x.mixed(c)
	tx := x.Tx.WithContext(c)
	if v.query != nil {
		tx = v.query(tx)
	}
	tx = x.where(tx, body.Conditions)
	if err := tx.Transaction(func(txx *gorm.DB) (err error) {
		if err = txx.Updates(body.Updates).Error; err != nil {
			return
		}
		if v.txNext != nil {
			if err = v.txNext(txx, body.Updates); err != nil {
				return
			}
		}
		return
	}); err != nil {
		return err
	}
	return "ok"
}

// DeleteBody Delete resource request body
type DeleteBody struct {
	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
}

// Delete resource
func (x *API) Delete(c *gin.Context) interface{} {
	var body DeleteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	v := x.mixed(c)
	tx := x.Tx.WithContext(c)
	if v.query != nil {
		tx = v.query(tx)
	}
	tx = x.where(tx, body.GetConditions())
	if err := tx.Transaction(func(txx *gorm.DB) (err error) {
		if err = txx.Delete(v.data).Error; err != nil {
			return
		}
		if v.txNext != nil {
			if err = v.txNext(txx, body.GetConditions()); err != nil {
				return
			}
		}
		return
	}); err != nil {
		return err
	}
	return "ok"
}
