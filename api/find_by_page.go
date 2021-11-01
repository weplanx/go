package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Pagination struct {
	Index int64 `json:"index" binding:"gt=0,number,required"`
	Limit int64 `json:"limit" binding:"gt=0,number,required"`
}

// FindByPageBody Get the request body of the paged list resource
type FindByPageBody struct {
	Where      bson.M `json:"where"`
	Sort       bson.M `json:"sort"`
	Pagination `json:"page" binding:"required"`
}

// FindByPage Get paging list resources
func (x *API) FindByPage(c *gin.Context) interface{} {
	if err := x.setCollection(c); err != nil {
		return err
	}
	var body FindByPageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.format(&body.Where); err != nil {
		return err
	}
	name := x.getName(c)
	var total int64
	var err error
	if body.Where != nil {
		if total, err = x.Db.Collection(name).CountDocuments(c, body.Where); err != nil {
			return err
		}
	} else {
		if total, err = x.Db.Collection(name).EstimatedDocumentCount(c); err != nil {
			return err
		}
	}
	opts := options.Find()
	page := body.Pagination
	opts.SetLimit(page.Limit)
	opts.SetSkip((page.Index - 1) * page.Limit)
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
	if err := cursor.All(c, &data); err != nil {
		return err
	}
	return gin.H{
		"lists": data,
		"total": total,
	}
}
