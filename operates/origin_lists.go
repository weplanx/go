package operates

import (
	"github.com/kainonly/gin-curd/typ"
	"github.com/kainonly/gin-helper/res"
)

type OriginListsBody struct {
	Where typ.Conditions
	Order []string
}

type OriginListsModel struct {
	typ.Common
	Model      interface{}
	Body       OriginListsBody
	conditions typ.Conditions
	query      typ.Query
	orders     []string
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

func (c *OriginListsModel) OrderBy(orders []string) *OriginListsModel {
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
	orders := append(c.orders, c.Body.Order...)
	for _, order := range orders {
		query = query.Order(order)
	}
	if len(c.field) != 0 {
		query = query.Select(c.field)
	}
	query.Find(&lists)
	return res.Data(lists)
}
