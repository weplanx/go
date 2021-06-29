package planx

import (
	"gorm.io/gorm"
)

type Planx struct {
	db *gorm.DB
}

func Initialize(db *gorm.DB) *Planx {
	return &Planx{
		db: db,
	}
}

func (x *Planx) Make() *Crud {
	return &Crud{}
}
