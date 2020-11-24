package operates

import (
	"github.com/kainonly/gin-curd/typ"
)

type GetBody struct {
	Id    interface{}
	Where typ.Conditions
	Order typ.Orders
}

type GetModel struct {
	typ.Common
	Model      interface{}
	Body       GetBody
	conditions typ.Conditions
	query      typ.Query
	orders     typ.Orders
	field      []string
}

func (c *GetModel) Where(conditions typ.Conditions) *GetModel {
	c.conditions = conditions
	return c
}

func (c *GetModel) Query(query typ.Query) *GetModel {
	c.query = query
	return c
}

func (c *GetModel) OrderBy(orders typ.Orders) *GetModel {
	c.orders = orders
	return c
}

func (c *GetModel) Field(field []string) *GetModel {
	c.field = field
	return c
}

func (c *GetModel) Exec() interface{} {
	data := make(map[string]interface{})
	query := c.Db.Model(c.Model)
	if c.Body.Id != nil {
		query = query.Where("`id` = ?", c.Body.Id)
	} else {
		conditions := append(c.conditions, c.Body.Where...)
		for _, condition := range conditions {
			query = query.Where("`"+condition[0].(string)+"` "+condition[1].(string)+" ?", condition[2])
		}
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
	query.First(&data)
	return data
}
