package curd

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

type execute struct {
	tx  *gorm.DB
	opt *Option
	*planOperator
	*whereOperator
	*subQueryOperator
	*orderByOperator
	*fieldOperator
	*updateOperator
	*afterHook
	*prepHook
}

// Set primary key or condition array query
func (c *execute) setIdOrConditions(tx *gorm.DB, id interface{}, value Conditions) *gorm.DB {
	if id != nil {
		tx = tx.Where("id = ?", id)
	} else {
		tx = c.setConditions(tx, value)
	}
	return tx
}

// Set condition array query
func (c *execute) setConditions(tx *gorm.DB, value Conditions) *gorm.DB {
	conditions := append(c.conditions, value...)
	for _, condition := range conditions {
		if !(strings.Contains(condition[0].(string), "->") && tx.Name() == "mysql") {
			query := condition[0].(string) + " " + condition[1].(string) + " ?"
			tx = tx.Where(query, condition[2])
		} else {
			column := strings.Split(condition[0].(string), "->")
			tx = tx.Where(
				"json_extract(?,?) "+condition[1].(string)+" ?",
				clause.Table{Name: column[0]},
				"$."+strings.Join(column[1:], "."),
				condition[2],
			)
		}
	}
	return tx
}

// Set order by query
func (c *execute) setOrders(tx *gorm.DB, value Orders) *gorm.DB {
	if len(c.orders) == 0 {
		c.orders = c.opt.Orders
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

// Set field
func (c *execute) setFields(tx *gorm.DB) *gorm.DB {
	if len(c.fields) != 0 {
		if !c.fieldOperator.exclude {
			tx = tx.Select(c.fields)
		} else {
			tx = tx.Omit(c.fields...)
		}
	}
	return tx
}

// Execute origin lists
func (c *execute) Originlists() interface{} {
	var lists []map[string]interface{}
	tx := c.tx.Model(c.model)
	apis := c.body.(originListsAPI)
	tx = c.setConditions(tx, apis.getConditions())
	if c.query != nil {
		tx = c.query(tx)
	}
	tx = c.setOrders(tx, apis.getOrders())
	tx = c.setFields(tx)
	tx.Find(&lists)
	return lists
}

// Execute lists
func (c *execute) Lists() interface{} {
	var lists []map[string]interface{}
	tx := c.tx.Model(c.model)
	apis := c.body.(listsAPI)
	tx = c.setConditions(tx, apis.getConditions())
	if c.query != nil {
		tx = c.query(tx)
	}
	tx = c.setOrders(tx, apis.getOrders())
	tx = c.setFields(tx)
	var total int64
	tx.Count(&total)
	page := apis.getPagination()
	tx = tx.Limit(int(page.Limit)).Offset(int((page.Index - 1) * page.Limit))
	tx.Find(&lists)
	return map[string]interface{}{
		"lists": lists,
		"total": total,
	}
}

// Execute get
func (c *execute) Get() interface{} {
	data := make(map[string]interface{})
	tx := c.tx.Model(c.model)
	apis := c.body.(getAPI)
	tx = c.setIdOrConditions(tx, apis.getId(), apis.getConditions())
	if c.query != nil {
		tx = c.query(tx)
	}
	tx = c.setOrders(tx, apis.getOrders())
	tx = c.setFields(tx)
	tx.First(&data)
	return data
}

// Execute add
func (c *execute) Add(value interface{}) interface{} {
	tx := c.tx
	if c.after == nil {
		if err := tx.Create(value).Error; err != nil {
			return err
		}
	} else {
		// There is a hook to start the transaction
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

// Execute edit
func (c *execute) Edit(value interface{}) interface{} {
	tx := c.tx.Model(c.model)
	apis := c.body.(editAPI)
	tx = c.setIdOrConditions(tx, apis.getId(), apis.getConditions())
	if c.query != nil {
		tx = c.query(tx)
	}
	if apis.isSwitch() {
		status := c.opt.UpdateStatus
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
			tx = tx.Select("*").Omit(c.opt.UpdateOmit...)
		}
	}
	if c.after == nil {
		if err := tx.Updates(value).Error; err != nil {
			return err
		}
	} else {
		// There is a hook to start the transaction
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

// Execute delete
func (c *execute) Delete() interface{} {
	tx := c.tx
	apis := c.body.(deleteAPI)
	id := apis.getId()
	if id != nil {
		tx = tx.Where("id in ?", id)
	} else {
		tx = c.setConditions(tx, apis.getConditions())
	}
	if c.query != nil {
		tx = c.query(tx)
	}
	if c.after == nil && c.prep == nil {
		if err := tx.Delete(c.model).Error; err != nil {
			return err
		}
	} else {
		// There is a hook to start the transaction
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
