package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// DeleteBody Delete resource request body
type DeleteBody struct {
	Where bson.M `json:"where" binding:"required"`
}

// Delete resource
func (x *API) Delete(c *gin.Context) interface{} {
	uri, err := x.getUri(c)
	if err != nil {
		return err
	}
	var body DeleteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.format(&body.Where); err != nil {
		return err
	}
	result, err := x.Db.Collection(uri.Collection).DeleteMany(c, body.Where)
	if err != nil {
		return err
	}
	return result
}
