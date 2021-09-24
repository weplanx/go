package api

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

type API struct {
	Db *gorm.DB
}

func InitializeAPI(tx *gorm.DB) *API {
	return &API{
		Db: tx,
	}
}

// Conditions conditions array
type Conditions [][3]interface{}

// Orders sort fields
type Orders map[string]string

// where conditional array initial
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

// FindOneBody Get a single resource request body
type FindOneBody struct {
	Model string `json:"model" binding:"required"`

	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// FindOne Get a single resource
func (x *API) FindOne(c *gin.Context) interface{} {
	var body FindOneBody
	if err := c.ShouldBind(&body); err != nil {
		return err
	}
	// TODO: Load schema cache
	tx := x.Db.WithContext(c).Table(body.Model)
	tx = x.where(tx, body.Conditions)
	tx = x.orderBy(tx, body.Orders)
	data := make(map[string]interface{})
	rows, err := tx.Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := tx.ScanRows(rows, &data); err != nil {
			return err
		}
		if err := x.toJSON(rows, &data); err != nil {
			return err
		}
	}
	return data
}

// FindBody Get the original list resource request body
type FindBody struct {
	Model string `json:"model" binding:"required"`

	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// Find Get the original list resource
func (x *API) Find(c *gin.Context) interface{} {
	var body FindBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	// TODO: Load schema cache
	tx := x.Db.WithContext(c).Table(body.Model)
	tx = x.where(tx, body.Conditions)
	tx = x.orderBy(tx, body.Orders)
	var data []map[string]interface{}
	rows, err := tx.Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		value := make(map[string]interface{})
		if err := tx.ScanRows(rows, &value); err != nil {
			return err
		}
		if err := x.toJSON(rows, &value); err != nil {
			return err
		}
		data = append(data, value)
	}
	return data
}

type Pagination struct {
	Index int `json:"index" binding:"gt=0,number,required"`
	Limit int `json:"limit" binding:"gt=0,number,required"`
}

// FindPageBody Get the request body of the paged list resource
type FindPageBody struct {
	Model string `json:"model" binding:"required"`

	Pagination `json:"page" binding:"required"`
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// Page Get paging list resources
func (x *API) Page(c *gin.Context) interface{} {
	var body FindPageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	// TODO: Load schema cache
	tx := x.Db.WithContext(c).Table(body.Model)
	tx = x.where(tx, body.Conditions)
	tx = x.orderBy(tx, body.Orders)
	var total int64
	tx.Count(&total)
	page := body.Pagination
	tx = tx.Limit(page.Limit).Offset((page.Index - 1) * page.Limit)
	var data []map[string]interface{}
	rows, err := tx.Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		value := make(map[string]interface{})
		if err := tx.ScanRows(rows, &value); err != nil {
			return err
		}
		if err := x.toJSON(rows, &value); err != nil {
			return err
		}
		data = append(data, value)
	}
	return gin.H{
		"lists": data,
		"total": total,
	}
}

type CreateBody struct {
	Model string                 `json:"model" binding:"required"`
	Data  map[string]interface{} `json:"data" binding:"required"`
}

// Create resources
func (x *API) Create(c *gin.Context) interface{} {
	var body CreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	// TODO: Load schema cache
	if err := x.Db.WithContext(c).
		Table(body.Model).
		Create(body.Data).Error; err != nil {
		return err
	}
	return "ok"
}

// UpdateBody Update resource request body
type UpdateBody struct {
	Model string                 `json:"model" binding:"required"`
	Data  map[string]interface{} `json:"data" binding:"required"`

	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
}

// Update resources
func (x *API) Update(c *gin.Context) interface{} {
	var body UpdateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	tx := x.Db.WithContext(c).Table(body.Model)
	tx = x.where(tx, body.Conditions)
	if err := tx.Updates(body.Data).Error; err != nil {
		return err
	}
	return "ok"
}

// DeleteBody Delete resource request body
type DeleteBody struct {
	Model string `json:"model" binding:"required"`

	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
}

// Delete resource
func (x *API) Delete(c *gin.Context) interface{} {
	var body DeleteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	tx := x.Db.WithContext(c).Table(body.Model)
	tx = x.where(tx, body.Conditions)
	if err := tx.Delete(nil).Error; err != nil {
		return err
	}
	return "ok"
}
