package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindBody Get the original list resource request body
type FindBody struct {
	Id    []*primitive.ObjectID `json:"id" validate:"omitempty,gt=0"`
	Where bson.M                `json:"where"`
	Sort  bson.M                `json:"sort" validate:"omitempty"`
}

// Find Get the original list resource
func (x *API) Find(c *fiber.Ctx) interface{} {
	ctx := c.UserContext()
	var body FindBody
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	if err := validator.New().Struct(body); err != nil {
		return err
	}
	name := x.collectionName(c)
	var filter bson.M
	if len(body.Id) != 0 {
		filter = bson.M{"_id": bson.M{"$in": body.Id}}
	} else {
		filter = body.Where
	}
	opts := options.Find()
	if len(body.Sort) != 0 {
		var sorts bson.D
		for k, v := range body.Sort {
			sorts = append(sorts, bson.E{Key: k, Value: v})
		}
		opts.SetSort(sorts)
		opts.SetAllowDiskUse(true)
	} else {
		opts.SetSort(bson.M{"_id": -1})
	}
	cursor, err := x.Db.Collection(name).Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	var data []map[string]interface{}
	if err = cursor.All(ctx, &data); err != nil {
		return err
	}
	return data
}
