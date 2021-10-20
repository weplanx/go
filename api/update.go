package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateBody Update resource request body
type UpdateBody struct {
	Where  `json:"where" binding:"required"`
	Update `json:"update" binding:"required"`
}

// Update resources
func (x *API) Update(c *gin.Context) interface{} {
	if err := x.setCollection(c); err != nil {
		return err
	}
	h := x.getHook(c)
	if h.body == nil {
		var updateBody UpdateBody
		if err := c.ShouldBindJSON(&updateBody); err != nil {
			return err
		}
		h.SetBody(updateBody)
	}
	body := h.body.(interface {
		GetWhere() *primitive.M
		GetUpdate() *primitive.M
	})
	if err := x.format(body.GetWhere()); err != nil {
		return err
	}
	result, err := x.collection(c).UpdateOne(c, body.GetWhere(), body.GetUpdate())
	if err != nil {
		return err
	}
	return result
}
