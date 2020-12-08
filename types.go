package curd

type originListsAPI interface {
	getConditions() Conditions
	getOrders() Orders
}

// General definition of origin list request body
type OriginLists struct {
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

type listsAPI interface {
	getConditions() Conditions
	getOrders() Orders
	getPagination() Pagination
}

// General definition of list request body
type Lists struct {

	// Condition array
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`

	// Order by
	Orders `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`

	// Page definition
	Pagination `json:"page" binding:"required"`
}

type getAPI interface {
	getId() interface{}
	getConditions() Conditions
	getOrders() Orders
}

// General definition of get request body
type Get struct {

	// Primary key
	Id interface{} `json:"id" binding:"required_without=Conditions"`

	// Condition array
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`

	// Order by
	Orders `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

func (c Get) getId() interface{} {
	return c.Id
}

type editAPI interface {
	getId() interface{}
	isSwitch() bool
	getConditions() Conditions
}

// General definition of edit request body, choose one of primary key or condition array
type Edit struct {

	// Primary key
	Id interface{} `json:"id" binding:"required_without=Conditions"`

	// Only the status field is updated
	Switch bool `json:"switch"`

	// Condition array
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
}

func (c Edit) getId() interface{} {
	return c.Id
}

func (c Edit) isSwitch() bool {
	return c.Switch
}

type deleteAPI interface {
	getId() interface{}
	getConditions() Conditions
}

// General definition of delete request body, choose one of primary key or condition array
type Delete struct {

	// Primary key
	Id interface{} `json:"id" binding:"required_without=Conditions"`

	// Condition array
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
}

func (c Delete) getId() interface{} {
	return c.Id
}

// Array condition definition
type Conditions [][]interface{}

func (c Conditions) getConditions() Conditions {
	return c
}

// Order definition
type Orders map[string]string

func (c Orders) getOrders() Orders {
	return c
}

// Paging request field definition.
type Pagination struct {

	// the paging index
	Index int64 `json:"index" binding:"gt=0,number,required"`

	// the page size
	Limit int64 `json:"limit" binding:"gt=0,number,required"`
}

func (c Pagination) getPagination() Pagination {
	return c
}
