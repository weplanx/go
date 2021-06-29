package bit

import "gorm.io/gorm"

type Bit struct {
	db *gorm.DB
}

func Initialize(db *gorm.DB) *Bit {
	return &Bit{db: db}
}

type CrudOption func(*Crud)

func SetOrderBy(orders []string) CrudOption {
	return func(option *Crud) {
		option.orderBy = orders
	}
}

func (x *Bit) Crud(model interface{}, options ...CrudOption) *Crud {
	crud := &Crud{
		db:    x.db,
		model: model,
	}
	for _, apply := range options {
		apply(crud)
	}
	return crud
}
