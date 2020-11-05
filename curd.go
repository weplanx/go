package curd

import (
	"github.com/kainonly/gin-curd/operates"
	"github.com/kainonly/gin-curd/typ"
	"gorm.io/gorm"
)

type Curd struct {
	typ.Common
}

func Initialize(db *gorm.DB) *Curd {
	c := new(Curd)
	c.Db = db
	return c
}

func (c *Curd) Originlists(model interface{}, body operates.OriginListsBody) *operates.OriginListsModel {
	m := new(operates.OriginListsModel)
	m.Common = c.Common
	m.Model = model
	m.Body = body
	return m
}

func (c *Curd) Lists(model interface{}, body operates.ListsBody) *operates.ListsModel {
	m := new(operates.ListsModel)
	m.Common = c.Common
	m.Model = model
	m.Body = body
	return m
}

func (c *Curd) Get(model interface{}, body operates.GetBody) *operates.GetModel {
	m := new(operates.GetModel)
	m.Common = c.Common
	m.Model = model
	m.Body = body
	return m
}

func (c *Curd) Add() *operates.AddModel {
	m := new(operates.AddModel)
	m.Common = c.Common
	return m
}

func (c *Curd) Edit(model interface{}, body operates.EditBody) *operates.EditModel {
	m := new(operates.EditModel)
	m.Common = c.Common
	m.Model = model
	m.Body = body
	return m
}

func (c *Curd) Delete(model interface{}, body operates.DeleteBody) *operates.DeleteModel {
	m := new(operates.DeleteModel)
	m.Common = c.Common
	m.Model = model
	m.Body = body
	return m
}
