package bit

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

type Crud struct {
	tx      *gorm.DB
	model   interface{}
	orderBy []string
}

// Conditions array condition definition
type Conditions [][]interface{}

// Orders definition
type Orders map[string]string

type GetAPI struct {
	Id         interface{} `json:"id" binding:"required_without=Conditions"`
	Conditions `json:"where" binding:"required_without=Id,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

func (x *Crud) setIdOrConditions(tx *gorm.DB, id interface{}, value Conditions) *gorm.DB {
	if id != nil {
		tx = tx.Where("id = ?", id)
	} else {
		tx = x.setConditions(tx, value)
	}
	return tx
}

func (x *Crud) setConditions(tx *gorm.DB, conditions Conditions) *gorm.DB {
	for _, condition := range conditions {
		if !(strings.Contains(condition[0].(string), "->") && tx.Name() == "mysql") {
			tx = tx.Where(
				"? "+condition[1].(string)+" ?",
				clause.Column{Name: condition[0].(string)},
				condition[2],
			)
		} else {
			column := strings.Split(condition[0].(string), "->")
			tx = tx.Where(
				"json_extract(?,?) "+condition[1].(string)+" ?",
				clause.Table{Name: column[0]},
				"$."+strings.Join(column[1:], "."),
				condition[2],
			)
		}
	}
	return tx
}

func (x *Crud) Get(c *gin.Context) interface{} {
	var body GetAPI
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	data := make(map[string]interface{})
	tx := x.tx.Model(x.model)
	tx = x.setIdOrConditions(tx, body.Id, body.Conditions)
	tx.First(&data)
	return gin.H(data)
}
