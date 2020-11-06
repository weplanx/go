package operates

import (
	"github.com/kainonly/gin-curd/typ"
	"gorm.io/gorm"
)

type AddModel struct {
	typ.Common
	after func(tx *gorm.DB) error
}

func (c *AddModel) After(hook func(tx *gorm.DB) error) *AddModel {
	c.after = hook
	return c
}

func (c *AddModel) Exec(value interface{}) interface{} {
	query := c.Db
	if c.after == nil {
		if err := query.Create(value).Error; err != nil {
			return typ.JSON{
				"error": 1,
				"msg":   err.Error(),
			}
		}
	} else {
		if err := query.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(value).Error; err != nil {
				return err
			}
			if err := c.after(tx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return typ.JSON{
				"error": 1,
				"msg":   err.Error(),
			}
		}
	}
	return typ.JSON{
		"error": 0,
		"msg":   "ok",
	}
}
