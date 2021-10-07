package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FindOneBody Get a single resource request body
type FindOneBody struct {
	//Conditions `json:"where" binding:"gte=0,dive,len=3,dive,required"`
	//Orders     `json:"order" binding:"omitempty,gte=0,dive,keys,endkeys,oneof=asc desc,required"`
	Where bson.M `json:"where"`
}

// FindOne Get a single resource
func (x *API) FindOne(c *gin.Context) interface{} {
	uri, err := x.getUri(c)
	if err != nil {
		return err
	}
	var body FindOneBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	data := make(map[string]interface{})
	if body.Where["_id"] != nil {
		body.Where["_id"], err = primitive.ObjectIDFromHex(body.Where["_id"].(string))
		if err != nil {
			return err
		}
	}
	if err := x.Db.
		Collection(uri.Collection).
		FindOne(c, body.Where).
		Decode(&data); err != nil {
		return err
	}
	return data
}
