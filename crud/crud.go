package crud

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
)

type Crud struct {
	Db *gorm.DB
}

func New(db *gorm.DB) *Crud {
	return &Crud{db}
}

// Make 创建控制器通用资源操作
//	参数:
//	 model: 模型名称
//	 options: 配置
func (x *Crud) Make(model interface{}, options ...Option) *Resource {
	c := &Resource{
		Db:    x.Db,
		Model: model,
	}
	for _, apply := range options {
		apply(c)
	}
	return c
}

type Resource struct {
	Db     *gorm.DB
	Model  interface{}
	orders Orders
}

type Option func(*Resource)

// SetOrders 设置默认排序
//	参数:
//	 orders: map[string]string
func SetOrders(orders Orders) Option {
	return func(option *Resource) {
		option.orders = orders
	}
}

// Conditions 条件数组
type Conditions [][]interface{}

func (c Conditions) GetConditions() Conditions {
	return c
}

// Orders 排序对象
type Orders map[string]string

func (c Orders) GetOrders() Orders {
	return c
}

// Set 设置 ID 或 条件数组
func (x *Resource) setIdOrConditions(tx *gorm.DB, id interface{}, value Conditions) *gorm.DB {
	if id != nil {
		return tx.Where("id = ?", id)
	} else {
		return x.setConditions(tx, value)
	}
}

// Set 设置条件数组
func (x *Resource) setConditions(tx *gorm.DB, conditions Conditions) *gorm.DB {
	for _, condition := range conditions {
		tx = tx.Where(
			"? "+condition[1].(string)+" ?",
			clause.Column{Name: condition[0].(string)},
			condition[2],
		)
	}
	return tx
}

// Set 设置排序
func (x *Resource) setOrders(tx *gorm.DB, orders Orders) *gorm.DB {
	for field, order := range orders {
		tx = tx.Order(field + " " + order)
	}
	for field, order := range x.orders {
		tx = tx.Order(field + " " + order)
	}
	return tx
}

// Set 设置混合操作
func (x *Resource) setMix(c *gin.Context, operator ...Operator) *mix {
	var v *mix
	if value, exists := c.Get(mixStart); exists {
		v = value.(*mix)
	} else {
		v = &mix{}
	}
	for _, operator := range operator {
		operator(v)
	}
	c.Set(mixComplete, v)
	return v
}

// GetMixVar 获取混合变量
func (x *Resource) GetMixVar(c *gin.Context) *mix {
	value, _ := c.Get(mixComplete)
	return value.(*mix)
}

// OriginListsBody 获取原始列表资源请求体
type OriginListsBody struct {
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// OriginLists 获取原始列表资源
func (x *Resource) OriginLists(c *gin.Context) interface{} {
	v := x.setMix(c,
		SetBody(&OriginListsBody{}),
		SetData(reflect.New(reflect.SliceOf(reflect.TypeOf(x.Model))).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetConditions() Conditions
		GetOrders() Orders
	})
	tx := x.Db.Model(x.Model)
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

// ListsBody 获取分页列表资源请求体
type ListsBody struct {
	Pagination `json:"page" binding:"required"`
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// Lists 获取分页列表资源
func (x *Resource) Lists(c *gin.Context) interface{} {
	v := x.setMix(c,
		SetBody(&ListsBody{}),
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
	tx := x.Db.Model(x.Model)
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

// GetBody 获取单条资源请求体
type GetBody struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

func (x *GetBody) GetId() interface{} {
	return x.Id
}

// Get 获取单条资源
func (x *Resource) Get(c *gin.Context) interface{} {
	v := x.setMix(c,
		SetBody(&GetBody{}),
		SetData(reflect.New(reflect.TypeOf(x.Model)).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetId() interface{}
		GetConditions() Conditions
		GetOrders() Orders
	})
	tx := x.Db.Model(x.Model)
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

// Add 创建资源
func (x *Resource) Add(c *gin.Context) interface{} {
	v := x.setMix(c)
	data := v.data
	if data == nil {
		v.Body = reflect.New(reflect.TypeOf(x.Model)).Interface()
		if err := c.ShouldBindJSON(v.Body); err != nil {
			return err
		}
		data = v.Body
	}
	if err := x.Db.Transaction(func(tx *gorm.DB) (err error) {
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

// EditBody 更新资源请求体
type EditBody struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
	Switch     *bool `json:"switch"`
}

func (x *EditBody) GetId() interface{} {
	return x.Id
}

// Edit 更新资源
func (x *Resource) Edit(c *gin.Context) interface{} {
	v := x.setMix(c)
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
				Type:      reflect.TypeOf(x.Model),
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
	tx := x.Db.Model(x.Model)
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

// DeleteBody 删除资源请求体
type DeleteBody struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
}

func (x *DeleteBody) GetId() interface{} {
	return x.Id
}

// Delete 删除资源
func (x *Resource) Delete(c *gin.Context) interface{} {
	v := x.setMix(c,
		SetBody(&DeleteBody{}),
		SetData(reflect.New(reflect.TypeOf(x.Model)).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetId() interface{}
		GetConditions() Conditions
	})
	tx := x.Db.Model(x.Model)
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
