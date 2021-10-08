package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// UpdateBody Update resource request body
type UpdateBody struct {
	Where  bson.M `json:"where" binding:"required"`
	Update bson.M `json:"update" binding:"required"`
}

// Update resources
func (x *API) Update(c *gin.Context) interface{} {
	uri, err := x.getUri(c)
	if err != nil {
		return err
	}
	var body UpdateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.format(&body.Where); err != nil {
		return err
	}
	result, err := x.Db.
		Collection(uri.Collection).
		UpdateOne(c, body.Where, body.Update)
	if err != nil {
		return err
	}
	return result
}
