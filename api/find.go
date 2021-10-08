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
	uri, err := x.getUri(c)
	if err != nil {
		return err
	}
	var body FindBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.format(&body.Where); err != nil {
		return err
	}
	opts := options.Find()
	opts.Sort = body.Sort
	cursor, err := x.Db.
		Collection(uri.Collection).
		Find(c, body.Where, opts)
	if err != nil {
		return err
	}
	var data []map[string]interface{}
	if err = cursor.All(c, &data); err != nil {
		return err
	}
	return data
}
