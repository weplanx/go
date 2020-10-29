package curd

import (
	"github.com/kainonly/iris-helper/res"
	"gorm.io/gorm"
)

type EditBody struct {
	Id     interface{}
	Where  Conditions
	Switch bool
}

type editModel struct {
	common
	model      interface{}
	body       EditBody
	conditions Conditions
	query      Query
	after      func(tx *gorm.DB) error
}

func (c *editModel) Where(conditions Conditions) *editModel {
	c.conditions = conditions
	return c
}

func (c *editModel) Query(query Query) *editModel {
	c.query = query
	return c
}

func (c *editModel) After(hook func(tx *gorm.DB) error) *editModel {
	c.after = hook
	return c
}

func (c *editModel) Exec(value interface{}) interface{} {
	query := c.db.Model(c.model)
	if c.body.Id != nil {
		query = query.Where("`id` = ?", c.body.Id)
	} else {
		conditions := append(c.conditions, c.body.Where...)
		for _, condition := range conditions {
			query = query.Where("`"+condition[0].(string)+"` "+condition[1].(string)+" ?", condition[2])
		}
	}
	if c.query != nil {
		query = c.query(query)
	}
	if c.after == nil {
		if err := query.Updates(value).Error; err != nil {
			return err
		}
	} else {
		err := query.Transaction(func(tx *gorm.DB) error {
			if err := tx.Updates(value).Error; err != nil {
				return err
			}
			if err := c.after(tx); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return res.Error(err)
		}
	}
	return res.Ok()
}
