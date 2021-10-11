package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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
		Filter() *primitive.M
		Update() bson.M
	})
	if err := x.format(body.Filter()); err != nil {
		return err
	}
	result, err := x.Collection.UpdateOne(c, body.Filter(), body.Update())
	if err != nil {
		return err
	}
	return result
}
