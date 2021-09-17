package crud

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// FirstBody Get a single resource request body
type FirstBody struct {
	Where bson.M `json:"where"`
}

// First Get a single resource
func (x *API) First(c *gin.Context) interface{} {
	body := FirstBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	var data bson.M
	if err := x.Collection.FindOne(c, body.Where).Decode(&data); err != nil {
		return err
	}
	return data
}
