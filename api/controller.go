package api

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type Controller struct {
	API *API
}

func AutoController(api *API) *Controller {
	return &Controller{api}
}

func (x *Controller) setCollectionName(c *fiber.Ctx) {
	c.SetUserContext(
		context.WithValue(context.Background(), "collection", c.Params("collection")),
	)
}

// Create resources
func (x *Controller) Create(c *fiber.Ctx) interface{} {
	var body bson.M
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	x.setCollectionName(c)
	result, err := x.API.Create(c.UserContext(), &body)
	if err != nil {
		return err
	}
	return result
}

// FindOne Get a single resource
func (x *Controller) FindOne(c *fiber.Ctx) interface{} {
	var body FindOneDto
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	if err := validator.New().Struct(body); err != nil {
		return err
	}
	x.setCollectionName(c)
	data := make(map[string]interface{})
	if err := x.API.FindOne(c.UserContext(), &body, &data); err != nil {
		return err
	}
	return data
}

func (x *Controller) Find(c *fiber.Ctx) interface{} {
	var body FindDto
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	if err := validator.New().Struct(body); err != nil {
		return err
	}
	x.setCollectionName(c)
	data := make([]map[string]interface{}, 0)
	if err := x.API.Find(c.UserContext(), &body, &data); err != nil {
		return err
	}
	return data
}

// FindByPage Get paging list resources
func (x *Controller) FindByPage(c *fiber.Ctx) interface{} {
	var body FindByPageDto
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	if err := validator.New().Struct(body); err != nil {
		return err
	}
	x.setCollectionName(c)
	result, err := x.API.FindByPage(c.UserContext(), &body)
	if err != nil {
		return err
	}
	return result
}

// Update resources
func (x *Controller) Update(c *fiber.Ctx) interface{} {
	var body UpdateDto
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	if err := validator.New().Struct(body); err != nil {
		return err
	}
	x.setCollectionName(c)
	result, err := x.API.Update(c.UserContext(), &body)
	if err != nil {
		return err
	}
	return result
}

// Delete resource
func (x *Controller) Delete(c *fiber.Ctx) interface{} {
	var body DeleteDto
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	if err := validator.New().Struct(body); err != nil {
		return err
	}
	x.setCollectionName(c)
	result, err := x.API.Delete(c.UserContext(), &body)
	if err != nil {
		return err
	}
	return result
}
