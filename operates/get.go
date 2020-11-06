package operates

import (
	"github.com/kainonly/gin-curd/typ"
)

type GetBody struct {
	Id    interface{}
	Where typ.Conditions
	Order []string
}

type GetModel struct {
	typ.Common
	Model      interface{}
	Body       GetBody
	conditions typ.Conditions
	query      typ.Query
	orders     []string
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

func (c *GetModel) OrderBy(orders []string) *GetModel {
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
	orders := append(c.orders, c.Body.Order...)
	for _, order := range orders {
		query = query.Order(order)
	}
	if len(c.field) != 0 {
		query = query.Select(c.field)
	}
	query.First(&data)
	return typ.JSON{
		"error": 0,
		"data":  data,
	}
}
