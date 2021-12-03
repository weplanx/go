package api

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// Create resources
func (x *API) Create(c *fiber.Ctx) interface{} {
	ctx := c.UserContext()
	var body bson.M
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	name := x.collectionName(c)
	result, err := x.Db.Collection(name).InsertOne(ctx, body)
	if err != nil {
		return err
	}
	return result
}
