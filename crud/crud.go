package crud

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"reflect"
)

type Crud struct {
	Db    *gorm.DB
	Model interface{}
}

// New 创建控制器通用资源操作
//	参数:
//	 tx: gorm
//	 model: 模型名称
func New(tx *gorm.DB, model interface{}) *Crud {
	return &Crud{
		Db:    tx,
		Model: model,
	}
}

// Conditions 条件数组
type Conditions [][3]interface{}

func (c Conditions) GetConditions() Conditions {
	return c
}

// Orders 排序对象
type Orders map[string]string

func (c Orders) GetOrders() Orders {
	return c
}

func (x *Crud) where(tx *gorm.DB, conds Conditions) *gorm.DB {
	for _, v := range conds {
		tx = tx.Where(gorm.Expr(v[0].(string)+" "+v[1].(string)+" ?", v[2]))
	}
	return tx
}

func (x *Crud) orderBy(tx *gorm.DB, orders Orders) *gorm.DB {
	for k, v := range orders {
		tx = tx.Order(k + " " + v)
	}
	return tx
}

// Set 设置混合操作
func (x *Crud) setMix(c *gin.Context, operator ...Operator) *mix {
	var v *mix
	if value, exists := c.Get(MixStart); exists {
		v = value.(*mix)
	} else {
		v = &mix{}
	}
	for _, operator := range operator {
		operator(v)
	}
	c.Set(MixComplete, v)
	return v
}

// GetMixVar 获取混合变量
func (x *Crud) GetMixVar(c *gin.Context) *mix {
	value, _ := c.Get(MixComplete)
	return value.(*mix)
}

// GetBody 获取单条资源请求体
type GetBody struct {
	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// Get 获取单条资源
func (x *Crud) Get(c *gin.Context) interface{} {
	v := x.setMix(c,
		SetBody(&GetBody{}),
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

// OriginListsBody 获取原始列表资源请求体
type OriginListsBody struct {
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// OriginLists 获取原始列表资源
func (x *Crud) OriginLists(c *gin.Context) interface{} {
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

// ListsBody 获取分页列表资源请求体
type ListsBody struct {
	Pagination `json:"page" binding:"required"`
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// Lists 获取分页列表资源
func (x *Crud) Lists(c *gin.Context) interface{} {
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

// Add 创建资源
func (x *Crud) Add(c *gin.Context) interface{} {
	v := x.setMix(c)
	data := v.data
	if data == nil {
		v.Body = reflect.New(reflect.TypeOf(x.Model)).Interface()
		if err := c.ShouldBindJSON(v.Body); err != nil {
			return err
		}
		data = v.Body
	}
	if err := x.Db.WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
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
	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
}

// Edit 更新资源
func (x *Crud) Edit(c *gin.Context) interface{} {
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
		conds = body.(EditBody).GetConditions()
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

// DeleteBody 删除资源请求体
type DeleteBody struct {
	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
}

// Delete 删除资源
func (x *Crud) Delete(c *gin.Context) interface{} {
	v := x.setMix(c,
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
