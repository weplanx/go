package operates

import (
	"github.com/kainonly/gin-curd/typ"
)

type ListsBody struct {
	Where typ.Conditions
	Order typ.Orders
	Page  Pagination
}

type Pagination struct {
	Index int64
	Limit int64
}

type ListsModel struct {
	typ.Common
	Model      interface{}
	Body       ListsBody
	conditions typ.Conditions
	query      typ.Query
	orders     typ.Orders
	field      []string
}

func (c *ListsModel) Where(conditions typ.Conditions) *ListsModel {
	c.conditions = conditions
	return c
}

func (c *ListsModel) Query(query typ.Query) *ListsModel {
	c.query = query
	return c
}

func (c *ListsModel) OrderBy(orders typ.Orders) *ListsModel {
	c.orders = orders
	return c
}

func (c *ListsModel) Field(field []string) *ListsModel {
	c.field = field
	return c
}

func (c *ListsModel) Exec() interface{} {
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
	page := c.Body.Page
	if page != (Pagination{}) {
		query = query.
			Limit(int(page.Limit)).
			Offset(int((page.Index - 1) * page.Limit))
	}
	var total int64
	query.Count(&total).Find(&lists)
	return typ.JSON{
		"error": 0,
		"data": typ.JSON{
			"lists": lists,
			"total": total,
		},
	}
}
