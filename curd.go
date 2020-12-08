package curd

import (
	"gorm.io/gorm"
)

type Curd struct {
	tx *gorm.DB
	*Option
}

type Option struct {
	Orders       Orders
	UpdateStatus string
	UpdateOmit   []string
}

func Initialize(tx *gorm.DB) *Curd {
	return &Curd{
		tx: tx,
		Option: &Option{
			Orders:       Orders{"id": "desc"},
			UpdateStatus: "status",
			UpdateOmit:   []string{"id", "create_time"},
		},
	}
}

func (c *Curd) Set(option Option) {
	c.Option = &option
}

func (c *Curd) Operates(operates ...Operator) *execute {
	exec := &execute{
		tx:               c.tx,
		opt:              c.Option,
		planOperator:     new(planOperator),
		whereOperator:    new(whereOperator),
		subQueryOperator: new(subQueryOperator),
		orderByOperator:  new(orderByOperator),
		fieldOperator:    new(fieldOperator),
		updateOperator:   new(updateOperator),
		afterHook:        new(afterHook),
		prepHook:         new(prepHook),
	}
	for _, operator := range operates {
		operator.apply(exec)
	}
	return exec
}
