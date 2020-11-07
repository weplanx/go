package operates

import (
	"github.com/kainonly/gin-curd/typ"
)

type OriginListsBody struct {
	Where typ.Conditions
	Order typ.Orders
}

type OriginListsModel struct {
	typ.Common
	Model      interface{}
	Body       OriginListsBody
	conditions typ.Conditions
	query      typ.Query
	orders     typ.Orders
	field      []string
}

func (c *OriginListsModel) Where(conditions typ.Conditions) *OriginListsModel {
	c.conditions = conditions
	return c
}

func (c *OriginListsModel) Query(query typ.Query) *OriginListsModel {
	c.query = query
	return c
}

func (c *OriginListsModel) OrderBy(orders typ.Orders) *OriginListsModel {
	c.orders = orders
	return c
}

func (c *OriginListsModel) Field(field []string) *OriginListsModel {
	c.field = field
	return c
}

func (c *OriginListsModel) Exec() interface{} {
	var lists []map[string]interface{}
	query := c.Db.Model(c.Model)
	conditions := append(c.conditions, c.Body.Where...)
	for _, condition := range conditions {
		query = query.Where("`"+condition[0].(string)+"` "+condition[1].(string)+" ?", condition[2])
	}
	if c.query != nil {
		query = c.query(query)
	}
	for filed, sort := range c.Body.Order {
		c.orders[filed] = sort
	}
	for filed, sort := range c.orders {
		query = query.Order(filed + " " + sort)
	}
	if len(c.field) != 0 {
		query = query.Select(c.field)
	}
	query.Find(&lists)
	return typ.JSON{
		"error": 0,
		"data":  lists,
	}
}
