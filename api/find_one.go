package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindOneBody Get a single resource request body
type FindOneBody struct {
	Where bson.M `json:"where"`
}

// FindOne Get a single resource
func (x *API) FindOne(c *gin.Context) interface{} {
	if err := x.setCollection(c); err != nil {
		return err
	}
	var body FindOneBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	data := make(map[string]interface{})
	if err := x.format(&body.Where); err != nil {
		return err
	}
	name := x.getName(c)
	opts := options.FindOne()
	projection, err := x.getProjection(c)
	if err != nil {
		return err
	}
	opts.SetProjection(projection)
	if err := x.Db.Collection(name).FindOne(c, body.Where, opts).Decode(&data); err != nil {
		return err
	}
	return data
}
