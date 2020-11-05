package operates

import (
	"github.com/kainonly/gin-curd/typ"
	"github.com/kainonly/gin-helper/res"
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
			return err
		}
	} else {
		err := query.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(value).Error; err != nil {
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
