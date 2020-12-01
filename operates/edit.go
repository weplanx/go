package operates

import (
	"github.com/kainonly/gin-curd/typ"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EditBody struct {
	Id     interface{}    `json:"id"`
	Where  typ.Conditions `json:"where"`
	Switch bool           `json:"switch"`
}

type EditModel struct {
	typ.Common
	Model      interface{}
	Body       EditBody
	conditions typ.Conditions
	query      typ.Query
	status     string
	omit       []string
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

func (c *EditModel) Status(field string) *EditModel {
	c.status = field
	return c
}

func (c *EditModel) Omit(fields []string) *EditModel {
	c.omit = fields
	return c
}

func (c *EditModel) After(hook func(tx *gorm.DB) error) *EditModel {
	c.after = hook
	return c
}

func (c *EditModel) Exec(value interface{}) interface{} {
	query := c.Db.Debug().Model(c.Model)
	if c.Body.Id != nil {
		query = query.Where("id = ?", c.Body.Id)
	} else {
		conditions := append(c.conditions, c.Body.Where...)
		for _, condition := range conditions {
			query = query.Where("? ? ?",
				clause.Column{Name: condition[0].(string)}, gorm.Expr(condition[1].(string)), condition[2],
			)
		}
	}
	if c.query != nil {
		query = c.query(query)
	}
	if c.Body.Switch {
		status := "status"
		if c.status != "" {
			status = c.status
		}
		query = query.Select(status)
	} else {
		omit := []string{"id", "create_time"}
		if len(c.omit) != 0 {
			omit = c.omit
		}
		query = query.Select("*").Omit(omit...)
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
