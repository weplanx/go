package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// Create resources
func (x *API) Create(c *gin.Context) interface{} {
	if err := x.setCollection(c); err != nil {
		return err
	}
	var body bson.M
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	name := x.getName(c)
	result, err := x.Db.Collection(name).InsertOne(c, body)
	if err != nil {
		return err
	}
	return result
}
