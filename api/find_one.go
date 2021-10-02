package api

import "github.com/gin-gonic/gin"

// FindOneBody Get a single resource request body
type FindOneBody struct {
	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
	Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
}

// FindOne Get a single resource
func (x *API) FindOne(c *gin.Context) interface{} {
	uri, err := x.getUri(c)
	if err != nil {
		return err
	}
	var body FindOneBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	// TODO: Load schema cache
	tx := x.Db.WithContext(c).Table(uri.Model)
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
