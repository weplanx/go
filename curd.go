package curd

import (
	"github.com/kainonly/gin-curd/api"
	"github.com/kainonly/gin-curd/typ"
	"gorm.io/gorm"
)

var (
	Orders       = typ.Orders{"id": "desc"}
	UpdateStatus = "status"
	UpdateOmit   = []string{"id", "create_time"}
)

type Curd struct {
	tx *gorm.DB
}

func Initialize(tx *gorm.DB) *Curd {
	return &Curd{tx: tx}
}

func (c *Curd) Operates(options ...Option) *operator {
	op := new(operator)
	op.tx = c.tx
	op.planOption = new(planOption)
	op.conditionsOption = new(conditionsOption)
	op.queryOption = new(queryOption)
	op.orderOption = new(orderOption)
	op.fieldOption = new(fieldOption)
	op.statusOption = new(statusOption)
	op.afterHook = new(afterHook)
	op.prepHook = new(prepHook)
	for _, option := range options {
		option.apply(op)
	}
	return op
}

// Option
type Option interface {
	apply(operator *operator)
}

func Plan(model interface{}, body interface{}) Option {
	return &planOption{
		model: model,
		body:  body,
	}
}

type planOption struct {
	model interface{}
	body  interface{}
}

func (c *planOption) apply(operator *operator) {
	operator.planOption = c
}

func Conditions(value typ.Conditions) Option {
	return &conditionsOption{conditions: value}
}

type conditionsOption struct {
	conditions typ.Conditions
}

func (c *conditionsOption) apply(operator *operator) {
	operator.conditionsOption = c
}

func Query(value typ.Query) Option {
	return &queryOption{query: value}
}

type queryOption struct {
	query typ.Query
}

func (c *queryOption) apply(operator *operator) {
	operator.queryOption = c
}

func OrderBy(value typ.Orders) Option {
	return &orderOption{orders: value}
}

type orderOption struct {
	orders typ.Orders
}

func (c *orderOption) apply(operator *operator) {
	operator.orderOption = c
}

func Field(value []string, exclude bool) Option {
	return &fieldOption{fields: value, exclude: exclude}
}

type fieldOption struct {
	exclude bool
	fields  []string
}

func (c *fieldOption) apply(operator *operator) {
	operator.fieldOption = c
}

func Update(status string) Option {
	return &statusOption{status: status}
}

type statusOption struct {
	status string
}

func (c *statusOption) apply(operator *operator) {
	operator.statusOption = c
}

func After(fn typ.Hook) Option {
	return &afterHook{after: fn}
}

type afterHook struct {
	after typ.Hook
}

func (c *afterHook) apply(operator *operator) {
	operator.afterHook = c
}

func Prep(fn typ.Hook) Option {
	return &prepHook{prep: fn}
}

type prepHook struct {
	prep typ.Hook
}

func (c *prepHook) apply(operator *operator) {
	operator.prepHook = c
}

// Operator
type operator struct {
	tx *gorm.DB
	*planOption
	*conditionsOption
	*queryOption
	*orderOption
	*fieldOption
	*statusOption
	*afterHook
	*prepHook
}

func (c *operator) setIdOrConditions(tx *gorm.DB, id interface{}, value typ.Conditions) *gorm.DB {
	if id != nil {
		tx = tx.Where("id = ?", id)
	} else {
		tx = c.setConditions(tx, value)
	}
	return tx
}

func (c *operator) setConditions(tx *gorm.DB, value typ.Conditions) *gorm.DB {
	conditions := append(c.conditions, value...)
	for _, condition := range conditions {
		query := condition[0].(string) + " " + condition[1].(string) + " ?"
		tx = tx.Where(query, condition[2])
	}
	return tx
}

func (c *operator) setOrders(tx *gorm.DB, value typ.Orders) *gorm.DB {
	if len(c.orders) == 0 {
		c.orders = Orders
	}
	for filed, sort := range value {
		c.orders[filed] = sort
	}
	for filed, sort := range c.orders {
		query := filed + " " + sort
		tx = tx.Order(query)
	}
	return tx
}

func (c *operator) setFields(tx *gorm.DB) *gorm.DB {
	if len(c.fields) != 0 {
		if !c.fieldOption.exclude {
			tx = tx.Select(c.fields)
		} else {
			tx = tx.Omit(c.fields...)
		}
	}
	return tx
}

func (c *operator) Originlists() interface{} {
	var lists []map[string]interface{}
	tx := c.tx.Model(c.model)
	apis := c.body.(api.OriginLists)
	tx = c.setConditions(tx, apis.GetConditions())
	if c.query != nil {
		tx = c.query(tx)
	}
	tx = c.setOrders(tx, apis.GetOrders())
	tx = c.setFields(tx)
	tx.Find(&lists)
	return lists
}

func (c *operator) Lists() interface{} {
	var lists []map[string]interface{}
	tx := c.tx.Model(c.model)
	apis := c.body.(api.Lists)
	tx = c.setConditions(tx, apis.GetConditions())
	if c.query != nil {
		tx = c.query(tx)
	}
	tx = c.setOrders(tx, apis.GetOrders())
	tx = c.setFields(tx)
	page := apis.GetPagination()
	tx = tx.Limit(int(page.Limit)).Offset(int((page.Index - 1) * page.Limit))
	var total int64
	tx.Count(&total).Find(&lists)
	return map[string]interface{}{
		"lists": lists,
		"total": total,
	}
}

func (c *operator) Get() interface{} {
	data := make(map[string]interface{})
	tx := c.tx.Model(c.model)
	apis := c.body.(api.Get)
	tx = c.setIdOrConditions(tx, apis.GetId(), apis.GetConditions())
	if c.query != nil {
		tx = c.query(tx)
	}
	tx = c.setOrders(tx, apis.GetOrders())
	tx = c.setFields(tx)
	tx.First(&data)
	return data
}

func (c *operator) Add(value interface{}) interface{} {
	tx := c.tx
	if c.after == nil {
		if err := tx.Create(value).Error; err != nil {
			return err
		}
	} else {
		if err := tx.Transaction(func(ttx *gorm.DB) error {
			if err := ttx.Create(value).Error; err != nil {
				return err
			}
			if err := c.after(ttx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return true
}

func (c *operator) Edit(value interface{}) interface{} {
	tx := c.tx.Model(c.model)
	apis := c.body.(api.Edit)
	tx = c.setIdOrConditions(tx, apis.GetId(), apis.GetConditions())
	if c.query != nil {
		tx = c.query(tx)
	}
	if apis.IsSwitch() {
		status := UpdateStatus
		if c.status != "" {
			status = c.status
		}
		tx = tx.Select(status)
	} else {
		if len(c.fields) != 0 {
			if !c.exclude {
				tx = tx.Select(c.fields)
			} else {
				tx = tx.Select("*").Omit(c.fields...)
			}
		} else {
			tx = tx.Select("*").Omit(UpdateOmit...)
		}
	}
	if c.after == nil {
		if err := tx.Updates(value).Error; err != nil {
			return err
		}
	} else {
		if err := tx.Transaction(func(txx *gorm.DB) error {
			if err := txx.Updates(value).Error; err != nil {
				return err
			}
			if err := c.after(txx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return true
}

func (c *operator) Delete() interface{} {
	tx := c.tx
	apis := c.body.(api.Delete)
	tx = c.setIdOrConditions(tx, apis.GetId(), apis.GetConditions())
	if c.query != nil {
		tx = c.query(tx)
	}
	if c.after == nil && c.prep == nil {
		if err := tx.Delete(c.model).Error; err != nil {
			return err
		}
	} else {
		if err := tx.Transaction(func(ttx *gorm.DB) error {
			if c.prep != nil {
				if err := c.prep(ttx); err != nil {
					return err
				}
			}
			if err := ttx.Delete(c.model).Error; err != nil {
				return err
			}
			if c.after != nil {
				if err := c.after(ttx); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return true
}
