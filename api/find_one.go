package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FindOneDto Get a single resource request body
type FindOneDto struct {
	Id    *primitive.ObjectID `json:"id" validate:"required_without=Where"`
	Where bson.M              `json:"where" validate:"required_without=Id,excluded_with=Id"`
}

// FindOne Get a single resource
func (x *API) FindOne(c *fiber.Ctx) interface{} {
	ctx := c.UserContext()
	var body FindOneDto
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	if err := validator.New().Struct(body); err != nil {
		return err
	}
	data := make(map[string]interface{})
	name := x.collectionName(c)
	var filter bson.M
	if body.Id != nil {
		filter = bson.M{"_id": body.Id}
	} else {
		filter = body.Where
	}
	if err := x.Db.Collection(name).FindOne(ctx, filter).Decode(&data); err != nil {
		return err
	}
	return data
}
