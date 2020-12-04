package api

import (
	"github.com/kainonly/gin-curd/typ"
)

type OriginLists interface {
	GetConditions() typ.Conditions
	GetOrders() typ.Orders
}

type Lists interface {
	GetConditions() typ.Conditions
	GetOrders() typ.Orders
	GetPagination() typ.Pagination
}

type Get interface {
	GetId() interface{}
	GetConditions() typ.Conditions
	GetOrders() typ.Orders
}

type Edit interface {
	GetId() interface{}
	IsSwitch() bool
	GetConditions() typ.Conditions
}

type Delete interface {
	GetId() interface{}
	GetConditions() typ.Conditions
}
