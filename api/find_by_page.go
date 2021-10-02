package api

import "github.com/gin-gonic/gin"

type Pagination struct {
	Index int `json:"index" binding:"gt=0,number,required"`
	Limit int `json:"limit" binding:"gt=0,number,required"`
}

// FindByPageBody Get the request body of the paged list resource
type FindByPageBody struct {
	Pagination `json:"page" binding:"required"`
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// FindByPage Get paging list resources
func (x *API) FindByPage(c *gin.Context) interface{} {
	uri, err := x.getUri(c)
	if err != nil {
		return err
	}
	var body FindByPageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	// TODO: Load schema cache
	tx := x.Db.WithContext(c).Table(uri.Model)
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
