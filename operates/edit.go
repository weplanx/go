package operates

import (
	"github.com/kainonly/gin-curd/typ"
	"gorm.io/gorm"
)

type EditBody struct {
	Id     interface{}
	Where  typ.Conditions
	Switch bool
}

type EditModel struct {
	typ.Common
	Model      interface{}
	Body       EditBody
	conditions typ.Conditions
	query      typ.Query
	after      func(tx *gorm.DB) error
}

func (c *EditModel) Where(conditions typ.Conditions) *EditModel {
	c.conditions = conditions
	return c
}

func (c *EditModel) Query(query typ.Query) *EditModel {
	c.query = query
	return c
}

func (c *EditModel) After(hook func(tx *gorm.DB) error) *EditModel {
	c.after = hook
	return c
}

func (c *EditModel) Exec(value interface{}) interface{} {
	query := c.Db.Model(c.Model)
	if c.Body.Id != nil {
		query = query.Where("`id` = ?", c.Body.Id)
	} else {
		conditions := append(c.conditions, c.Body.Where...)
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
		if err := query.Transaction(func(tx *gorm.DB) error {
			if err := tx.Updates(value).Error; err != nil {
				return err
			}
			if err := c.after(tx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return true
}
