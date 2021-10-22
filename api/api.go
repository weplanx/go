package api

import (
	"github.com/gin-gonic/gin"
	"github.com/weplanx/support/basic"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type API struct {
	Mongo          *mongo.Client
	Db             *mongo.Database
	CollectionName string
	ProjectionNone bool
}

type Where bson.M

func (x Where) GetWhere() *primitive.M {
	value := primitive.M(x)
	return &value
}

type Update bson.M

func (x Update) GetUpdate() *primitive.M {
	value := primitive.M(x)
	return &value
}

type OptionFunc func(*API)

func SetCollection(name string) OptionFunc {
	return func(api *API) {
		api.CollectionName = name
	}
}

func ProjectionNone() OptionFunc {
	return func(api *API) {
		api.ProjectionNone = true
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
		c.Set("collection", x.CollectionName)
		return nil
	}
	var uri struct {
		Collection string `uri:"collection" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		return err
	}
	c.Set("collection", uri.Collection)
	return nil
}

func (x *API) getName(c *gin.Context) string {
	name, _ := c.Get("collection")
	return name.(string)
}

func (x *API) getProjection(c *gin.Context) (projection bson.M, err error) {
	if x.ProjectionNone {
		return
	}
	var schema basic.Schema
	name := x.getName(c)
	if err = x.Db.Collection("schema").FindOne(c, bson.M{
		"key": name,
	}).Decode(&schema); err != nil {
		return
	}
	projection = make(bson.M)
	for _, x := range schema.Fields {
		log.Println(x.Key, x.Private)
		//if x.Private == true {
		//	projection[x.Key] = 0
		//}
	}
	return
}

func (x *API) getHook(c *gin.Context) *Hook {
	if value, exists := c.Get("hook"); exists {
		return value.(*Hook)
	}
	return &Hook{}
}
