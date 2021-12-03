package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateBody Update resource request body
type UpdateBody struct {
	Id     *primitive.ObjectID `json:"id" validate:"required_without=Where"`
	Where  bson.M              `json:"where" validate:"required_without=Id"`
	Update bson.M              `json:"update" validate:"required"`
}

// Update resources
func (x *API) Update(c *fiber.Ctx) interface{} {
	ctx := c.UserContext()
	var body UpdateBody
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	if err := validator.New().Struct(body); err != nil {
		return err
	}
	name := x.collectionName(c)
	filter := body.Where
	if body.Id != nil {
		filter = bson.M{"_id": body.Id}
	}
	result, err := x.Db.Collection(name).UpdateOne(ctx, filter, body.Update)
	if err != nil {
		return err
	}
	return result
}
