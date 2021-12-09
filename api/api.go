package api

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type API struct {
	MongoClient *mongo.Client
	Db          *mongo.Database
}

func New(client *mongo.Client, db *mongo.Database) *API {
	x := new(API)
	x.MongoClient = client
	x.Db = db
	return x
}

func (x *API) collectionName(c *fiber.Ctx) string {
	return c.Params("collection")
}
