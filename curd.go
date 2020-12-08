package curd

import (
	"gorm.io/gorm"
)

type Curd struct {
	tx *gorm.DB
	*Option
}

type Option struct {

	// Default order by, Orders{"id": "desc"}
	Orders Orders

	// Default updated status field, "status"
	UpdateStatus string

	// Default updated exclude fields, []string{"id", "create_time"}
	UpdateOmit []string
}

// Initialize the curd auxiliary library
//	@param `tx` *gorm.DB
//	@return Curd
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

// Set curd auxiliary configuration
//	@param `option` Option
func (c *Curd) Set(option Option) {
	c.Option = &option
}

// Start Curd operation process
//	@param `operates` ...Operator
//	@return execute
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
