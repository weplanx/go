package crud

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"reflect"
)

type Crud struct {
	Db    *gorm.DB
	Model interface{}
}

// New create controller general resource operation
//	 model: GORM Models
func New(tx *gorm.DB, model interface{}) *Crud {
	return &Crud{
		Db:    tx,
		Model: model,
	}
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
func (x *Crud) where(tx *gorm.DB, conds Conditions) *gorm.DB {
	for _, v := range conds {
		tx = tx.Where(gorm.Expr(v[0].(string)+" "+v[1].(string)+" ?", v[2]))
	}
	return tx
}

// orderBy generate request ordering rules
func (x *Crud) orderBy(tx *gorm.DB, orders Orders) *gorm.DB {
	for k, v := range orders {
		tx = tx.Order(k + " " + v)
	}
	return tx
}

// Set default initial mix
func (x *Crud) mixed(c *gin.Context, operator ...Operator) *mixed {
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

// FindOneBody Get a single resource request body
type FindOneBody struct {
	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// FindOne Get a single resource
func (x *Crud) FindOne(c *gin.Context) interface{} {
	v := x.mixed(c,
		SetBody(&FindOneBody{}),
		SetData(reflect.New(reflect.TypeOf(x.Model)).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetConditions() Conditions
		GetOrders() Orders
	})
	tx := x.Db.WithContext(c).Model(x.Model)
	if v.query != nil {
		tx = v.query(tx)
	}
	tx = x.where(tx, body.GetConditions())
	tx = x.orderBy(tx, body.GetOrders())
	if err := tx.First(v.data).Error; err != nil {
		return err
	}
	return v.data
}

// FindManyBody Get the original list resource request body
type FindManyBody struct {
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// FindMany Get the original list resource
func (x *Crud) FindMany(c *gin.Context) interface{} {
	v := x.mixed(c,
		SetBody(&FindManyBody{}),
		SetData(reflect.New(reflect.SliceOf(reflect.TypeOf(x.Model))).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetConditions() Conditions
		GetOrders() Orders
	})
	tx := x.Db.WithContext(c).Model(x.Model)
	if v.query != nil {
		tx = v.query(tx)
	}
	tx = x.where(tx, body.GetConditions())
	tx = x.orderBy(tx, body.GetOrders())
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

// FindPageBody Get the request body of the paged list resource
type FindPageBody struct {
	Pagination `json:"page" binding:"required"`
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// FindPage Get paging list resources
func (x *Crud) FindPage(c *gin.Context) interface{} {
	v := x.mixed(c,
		SetBody(&FindPageBody{}),
		SetData(reflect.New(reflect.SliceOf(reflect.TypeOf(x.Model))).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetPagination() Pagination
		GetConditions() Conditions
		GetOrders() Orders
	})
	tx := x.Db.WithContext(c).Model(x.Model)
	if v.query != nil {
		tx = v.query(tx)
	}
	tx = x.where(tx, body.GetConditions())
	tx = x.orderBy(tx, body.GetOrders())
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

// Create resources
func (x *Crud) Create(c *gin.Context) interface{} {
	v := x.mixed(c)
	data := v.data
	if data == nil {
		v.Body = reflect.New(reflect.TypeOf(x.Model))
		if err := c.ShouldBindJSON(v.Body.(reflect.Value).Interface()); err != nil {
			return err
		}
		data = v.Body.(reflect.Value).Elem().Interface()
	}
	if err := x.Db.WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(data).Error; err != nil {
			log.Println(err)
			return
		}
		if v.txNext != nil {
			if err = v.txNext(tx, data); err != nil {
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
}

// Update resources
func (x *Crud) Update(c *gin.Context) interface{} {
	v := x.mixed(c)
	var body interface{}
	data := v.data
	if data == nil {
		v.Body = reflect.New(reflect.StructOf([]reflect.StructField{
			{
				Name:      "EditBody",
				Type:      reflect.TypeOf(UpdateBody{}),
				Anonymous: true,
			},
			{
				Name: "Updates",
				Type: reflect.TypeOf(x.Model),
				Tag:  `json:"updates"`,
			},
		})).Interface()
		if err := c.ShouldBindJSON(v.Body); err != nil {
			return err
		}
		elem := reflect.ValueOf(v.Body).Elem()
		body = elem.FieldByName("EditBody").Interface()
		data = elem.FieldByName("Updates").Interface()
	}
	tx := x.Db.WithContext(c).Model(x.Model)
	if v.query != nil {
		tx = v.query(tx)
	}
	var conds Conditions
	if body != nil {
		conds = body.(UpdateBody).GetConditions()
	} else {
		conds = body.(interface{ GetConditions() Conditions }).GetConditions()
	}
	tx = x.where(tx, conds)
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

// DeleteBody Delete resource request body
type DeleteBody struct {
	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
}

// Delete resource
func (x *Crud) Delete(c *gin.Context) interface{} {
	v := x.mixed(c,
		SetBody(&DeleteBody{}),
		SetData(reflect.New(reflect.TypeOf(x.Model)).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetConditions() Conditions
	})
	tx := x.Db.WithContext(c).Model(x.Model)
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
