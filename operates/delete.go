package operates

import (
	"github.com/kainonly/gin-curd/typ"
	"gorm.io/gorm"
)

type DeleteBody struct {
	Id    interface{}
	Where typ.Conditions
}

type DeleteModel struct {
	typ.Common
	Model      interface{}
	Body       DeleteBody
	conditions typ.Conditions
	query      typ.Query
	prep       func(tx *gorm.DB) error
	after      func(tx *gorm.DB) error
}

func (c *DeleteModel) Where(conditions typ.Conditions) *DeleteModel {
	c.conditions = conditions
	return c
}

func (c *DeleteModel) Query(query typ.Query) *DeleteModel {
	c.query = query
	return c
}

func (c *DeleteModel) Prep(hook func(tx *gorm.DB) error) *DeleteModel {
	c.prep = hook
	return c
}

func (c *DeleteModel) After(hook func(tx *gorm.DB) error) *DeleteModel {
	c.after = hook
	return c
}

func (c *DeleteModel) Exec() interface{} {
	query := c.Db
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
	if c.after == nil && c.prep == nil {
		if err := query.Delete(c.Model).Error; err != nil {
			return err
		}
	} else {
		if err := query.Transaction(func(tx *gorm.DB) error {
			if c.prep != nil {
				if err := c.prep(tx); err != nil {
					return err
				}
			}
			if err := tx.Delete(c.Model).Error; err != nil {
				return err
			}
			if c.after != nil {
				if err := c.after(tx); err != nil {
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
