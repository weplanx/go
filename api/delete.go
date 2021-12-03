package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteBody Delete resource request body
type DeleteBody struct {
	Id    []*primitive.ObjectID `json:"id" validate:"required_without=Where,omitempty,gt=0"`
	Where bson.M                `json:"where" validate:"required_without=Id,excluded_with=Id"`
}

// Delete resource
func (x *API) Delete(c *fiber.Ctx) interface{} {
	ctx := c.UserContext()
	var body DeleteBody
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
	result, err := x.Db.Collection(name).DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	return result
}
