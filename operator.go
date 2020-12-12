package curd

import "gorm.io/gorm"

type Operator interface {
	apply(exec *execute)
}

// Plan a model expression
//	@param `model` interface{} GORM defined model
//	@param `body` interface{} Request body
//	@return Operator
func Plan(model interface{}, body interface{}) Operator {
	return &planOperator{
		model: model,
		body:  body,
	}
}

type planOperator struct {
	model interface{}
	body  interface{}
}

func (c *planOperator) apply(exec *execute) {
	exec.planOperator = c
}

// Set condition array
//	@param `value` Conditions Condition array
//	@return Operator
func Where(value Conditions) Operator {
	return &whereOperator{conditions: value}
}

type whereOperator struct {
	conditions Conditions
}

func (c *whereOperator) apply(exec *execute) {
	exec.whereOperator = c
}

// Set sub query
//	@param `fn` func(tx *gorm.DB) *gorm.DB
//	@return Operator
func SubQuery(fn func(tx *gorm.DB) *gorm.DB) Operator {
	return &subQueryOperator{query: fn}
}

type subQueryOperator struct {
	query func(tx *gorm.DB) *gorm.DB
}

func (c *subQueryOperator) apply(exec *execute) {
	exec.subQueryOperator = c
}

// Set order by
//	@param `value` Orders
//	@return Operator
func OrderBy(value Orders) Operator {
	return &orderByOperator{orders: value}
}

type orderByOperator struct {
	orders Orders
}

func (c *orderByOperator) apply(exec *execute) {
	exec.orderByOperator = c
}

// Set selecting specific fields
//	@param `value` []string
//	@param `exclude` bool
//	@return Operator
func Field(value []string, exclude bool) Operator {
	return &fieldOperator{fields: value, exclude: exclude}
}

type fieldOperator struct {
	exclude bool
	fields  []string
}

func (c *fieldOperator) apply(exec *execute) {
	exec.fieldOperator = c
}

// Set update
//	@param `status` string When switch is true, update the status field
//	@return Operator
func Update(status string) Operator {
	return &updateOperator{status: status}
}

type updateOperator struct {
	status string
}

func (c *updateOperator) apply(exec *execute) {
	exec.updateOperator = c
}

// After hook, when the return is error, the transaction will be rolled back
//	@param `fn` func(tx *gorm.DB) error
//	@return Operator
func After(fn func(tx *gorm.DB) error) Operator {
	return &afterHook{after: fn}
}

type afterHook struct {
	after func(tx *gorm.DB) error
}

func (c *afterHook) apply(exec *execute) {
	exec.afterHook = c
}

// Preparation hook, the transaction will be terminated early when the return is error
//	@param `fn` func(tx *gorm.DB) error
//	@return Operator
func Prep(fn func(tx *gorm.DB) error) Operator {
	return &prepHook{prep: fn}
}

type prepHook struct {
	prep func(tx *gorm.DB) error
}

func (c *prepHook) apply(exec *execute) {
	exec.prepHook = c
}
