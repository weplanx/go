package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// Create resources
func (x *API) Create(c *gin.Context) interface{} {
	uri, err := x.getUri(c)
	if err != nil {
		return err
	}
	var body bson.M
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	result, err := x.Db.Collection(uri.Collection).InsertOne(c, body)
	if err != nil {
		return err
	}
	return result
}
