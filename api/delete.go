package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteBody Delete resource request body
type DeleteBody struct {
	Where `json:"where" binding:"required"`
}

// Delete resource
func (x *API) Delete(c *gin.Context) interface{} {
	if err := x.setCollection(c); err != nil {
		return err
	}
	h := x.getHook(c)
	if h.body == nil {
		var deleteBody DeleteBody
		if err := c.ShouldBindJSON(&deleteBody); err != nil {
			return err
		}
		h.SetBody(deleteBody)
	}
	body := h.body.(interface {
		GetWhere() *primitive.M
	})
	if err := x.format(body.GetWhere()); err != nil {
		return err
	}
	result, err := x.collection(c).DeleteMany(c, body.GetWhere())
	if err != nil {
		return err
	}
	return result
}
