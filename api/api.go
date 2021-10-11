package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type API struct {
	Mongo          *mongo.Client
	Db             *mongo.Database
	Collection     *mongo.Collection
	CollectionName string
}

type Where bson.M

func (x Where) Filter() *primitive.M {
	value := primitive.M(x)
	return &value
}

type Update bson.M

func (x Update) Update() bson.M {
	return bson.M(x)
}

type OptionFunc func(*API)

func SetCollection(name string) OptionFunc {
	return func(api *API) {
		api.CollectionName = name
		api.Collection = api.Db.Collection(name)
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

func (x *API) format(input *primitive.M) (err error) {
	p := *input
	if p["_id"] != nil {
		switch value := p["_id"].(type) {
		case string:
			var id primitive.ObjectID
			if id, err = primitive.ObjectIDFromHex(value); err != nil {
				return
			}
			p["_id"] = id
			break
		case map[string]interface{}:
			values := value["$in"].([]interface{})
			ids := make([]primitive.ObjectID, len(values))
			for k, v := range values {
				if ids[k], err = primitive.ObjectIDFromHex(v.(string)); err != nil {
					return
				}
			}
			p["_id"].(map[string]interface{})["$in"] = ids
			break
		}
	}
	return
}

func (x *API) setCollection(c *gin.Context) error {
	if x.CollectionName != "" {
		return nil
	}
	var uri struct {
		Collection string `uri:"collection" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		return err
	}
	x.CollectionName = uri.Collection
	x.Collection = x.Db.Collection(uri.Collection)
	return nil
}

func (x *API) getHook(c *gin.Context) *Hook {
	if value, exists := c.Get("hook"); exists {
		return value.(*Hook)
	}
	return &Hook{}
}
