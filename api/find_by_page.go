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
	Pagination `json:"page" binding:"required"`
	Where      bson.M `json:"where"`
}

// FindByPage Get paging list resources
func (x *API) FindByPage(c *gin.Context) interface{} {
	uri, err := x.getUri(c)
	if err != nil {
		return err
	}
	var body FindByPageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.where(&body.Where); err != nil {
		return err
	}
	total, err := x.Db.
		Collection(uri.Collection).
		CountDocuments(c, body.Where)
	if err != nil {
		return err
	}
	opts := options.Find()
	page := body.Pagination
	opts.SetLimit(page.Limit)
	opts.SetSkip((page.Index - 1) * page.Limit)
	cursor, err := x.Db.
		Collection(uri.Collection).
		Find(c, body.Where, opts)
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
