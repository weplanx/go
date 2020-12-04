package typ

import "gorm.io/gorm"

type OriginLists struct {
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

type Lists struct {
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
	Pagination `json:"page" binding:"required"`
}

type Get struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

func (c Get) GetId() interface{} {
	return c.Id
}

type Edit struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Switch     bool        `json:"switch"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
}

func (c Edit) GetId() interface{} {
	return c.Id
}

func (c Edit) IsSwitch() bool {
	return c.Switch
}

type Delete struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
}

func (c Delete) GetId() interface{} {
	return c.Id
}

type Conditions [][]interface{}

func (c Conditions) GetConditions() Conditions {
	return c
}

type Query func(tx *gorm.DB) *gorm.DB

type Orders map[string]string

func (c Orders) GetOrders() Orders {
	return c
}

type Pagination struct {
	Index int64 `json:"index" binding:"gt=0,number,required"`
	Limit int64 `json:"limit" binding:"gt=0,number,required"`
}

func (c Pagination) GetPagination() Pagination {
	return c
}

type Hook func(tx *gorm.DB) error
