package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type API struct {
	Mongo      *mongo.Client
	Db         *mongo.Database
	Collection string
}

type OptionFunc func(*API)

func SetCollection(name string) OptionFunc {
	return func(api *API) {
		api.Collection = name
	}
}

func New(client *mongo.Client, db *mongo.Database, options ...OptionFunc) *API {
	api := new(API)
	api.Mongo = client
	api.Db = db
	for _, option := range options {
		option(api)
	}
	return api
}

func (x *API) where(input *bson.M) (err error) {
	if (*input)["_id"] != nil {
		switch value := (*input)["_id"].(type) {
		case string:
			var id primitive.ObjectID
			if id, err = primitive.ObjectIDFromHex(value); err != nil {
				return
			}
			(*input)["_id"] = id
			break
		case map[string]interface{}:
			values := value["$in"].([]interface{})
			ids := make([]primitive.ObjectID, len(values))
			for k, v := range values {
				if ids[k], err = primitive.ObjectIDFromHex(v.(string)); err != nil {
					return
				}
			}
			(*input)["_id"].(map[string]interface{})["$in"] = ids
			break
		}
	}
	return
}

type Uri struct {
	Collection string `uri:"collection" binding:"required"`
}

func (x *API) getUri(c *gin.Context) (uri Uri, err error) {
	if x.Collection != "" {
		uri.Collection = x.Collection
		return
	}
	if err = c.ShouldBindUri(&uri); err != nil {
		return
	}
	return
}
