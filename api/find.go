package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindBody Get the original list resource request body
type FindBody struct {
	Where bson.M `json:"where"`
	Sort  bson.M `json:"sort"`
}

// Find Get the original list resource
func (x *API) Find(c *gin.Context) interface{} {
	if err := x.setCollection(c); err != nil {
		return err
	}
	var body FindBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.format(&body.Where); err != nil {
		return err
	}
	name := x.getName(c)
	opts := options.Find()
	opts.Sort = body.Sort
	projection, err := x.getProjection(c)
	if err != nil {
		return err
	}
	opts.SetProjection(projection)
	cursor, err := x.Db.Collection(name).Find(c, body.Where, opts)
	if err != nil {
		return err
	}
	var data []map[string]interface{}
	if err = cursor.All(c, &data); err != nil {
		return err
	}
	return data
}
