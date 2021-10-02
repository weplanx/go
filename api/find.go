package api

import "github.com/gin-gonic/gin"

// FindBody Get the original list resource request body
type FindBody struct {
	Conditions `json:"where" binding:"omitempty,gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// Find Get the original list resource
func (x *API) Find(c *gin.Context) interface{} {
	uri, err := x.getUri(c)
	if err != nil {
		return err
	}
	var body FindBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	// TODO: Load schema cache
	tx := x.Db.WithContext(c).Table(uri.Model)
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
