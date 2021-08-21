package crud

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
)

// GetBody 获取单条资源请求体
type GetBody struct {
	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// Get 获取单条资源
func (x *Resource) Get(c *gin.Context) interface{} {
	v := x.setMix(c,
		SetBody(&GetBody{}),
		SetData(reflect.New(reflect.TypeOf(x.Model)).Interface()),
	)
	if err := c.ShouldBindJSON(v.Body); err != nil {
		return err
	}
	body := v.Body.(interface {
		GetConditions() Conditions
		GetOrders() Orders
	})
	tx := x.Db.WithContext(c).Model(x.Model)
	if v.query != nil {
		tx = v.query(tx)
	}
	tx = tx.Clauses(
		x.toClauseWhere(body.GetConditions()),
		x.toClauseOrderBy(body.GetOrders()),
	)
	if err := tx.First(v.data).Error; err != nil {
		return err
	}
	return v.data
}

func (x *Resource) toClauseWhere(conds Conditions) clause.Where {
	exprs := make([]clause.Expression, 0)
	for _, v := range conds {
		exprs = append(exprs, gorm.Expr(v[0].(string)+" "+v[1].(string)+" ?", v[2]))
	}
	return clause.Where{
		Exprs: exprs,
	}
}

func (x *Resource) toClauseOrderBy(orders Orders) clause.OrderBy {
	columns := make([]clause.OrderByColumn, 0)
	for k, v := range orders {
		columns = append(columns, clause.OrderByColumn{
			Column: clause.Column{Name: k + " " + v, Raw: true},
		})
	}
	return clause.OrderBy{Columns: columns}
}
