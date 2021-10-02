package api

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

type API struct {
	Db    *gorm.DB
	Model string
}

type OptionFunc func(*API)

func SetModel(name string) OptionFunc {
	return func(api *API) {
		api.Model = name
	}
}

func New(db *gorm.DB, options ...OptionFunc) *API {
	api := new(API)
	api.Db = db
	for _, option := range options {
		option(api)
	}
	return api
}

// Conditions 条件数组
type Conditions [][3]interface{}

// Orders 排序字段
type Orders map[string]string

func (x *API) where(tx *gorm.DB, conds Conditions) *gorm.DB {
	for _, v := range conds {
		tx = tx.Where(gorm.Expr(v[0].(string)+" "+v[1].(string)+" ?", v[2]))
	}
	return tx
}

// orderBy sort fields initial
func (x *API) orderBy(tx *gorm.DB, orders Orders) *gorm.DB {
	for k, v := range orders {
		tx = tx.Order(k + " " + v)
	}
	return tx
}

func (x *API) toJSON(rows *sql.Rows, value *map[string]interface{}) (err error) {
	typs, err := rows.ColumnTypes()
	if err != nil {
		return
	}
	for _, typ := range typs {
		switch typ.DatabaseTypeName() {
		case "ARRAY":
		case "JSON":
		case "JSONB":
			var JSON json.RawMessage
			if err = jsoniter.Unmarshal([]byte((*value)[typ.Name()].(string)), &JSON); err != nil {
				return
			}
			(*value)[typ.Name()] = &JSON
			break
		}
	}
	return
}

type Uri struct {
	Model string `uri:"model" binding:"required"`
}

func (x *API) getUri(c *gin.Context) (uri Uri, err error) {
	if x.Model != "" {
		uri.Model = x.Model
		return
	}
	if err = c.ShouldBindUri(&uri); err != nil {
		return
	}
	return
}
