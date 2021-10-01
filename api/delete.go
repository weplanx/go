package api

import "github.com/gin-gonic/gin"

// DeleteBody Delete resource request body
type DeleteBody struct {
	Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
}

// Delete resource
func (x *API) Delete(c *gin.Context) interface{} {
	var uri Uri
	if err := c.ShouldBindUri(&uri); err != nil {
		return err
	}
	var body DeleteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	tx := x.Db.WithContext(c).Table(uri.Model)
	tx = x.where(tx, body.Conditions)
	if err := tx.Delete(nil).Error; err != nil {
		return err
	}
	return "ok"
}
