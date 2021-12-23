package api

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Controller struct {
	API  *API
	PATH string
}

func AutoController(api *API) *Controller {
	return &Controller{api, ""}
}

func SetController(api *API, path string) *Controller {
	return &Controller{api, path}
}

func (x *Controller) setCollectionName(c *fiber.Ctx) {
	name := c.Params("collection")
	if x.PATH != "" {
		name = x.PATH
	}
	c.SetUserContext(
		context.WithValue(context.Background(), "collection", name),
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

// FindOneDto Get a single resource request body
type FindOneDto struct {
	Id    primitive.ObjectID `json:"id" validate:"required_without=Where"`
	Where bson.M             `json:"where" validate:"required_without=Id,excluded_with=Id"`
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

// FindDto Get the original list resource request body
type FindDto struct {
	Id    []primitive.ObjectID `json:"id" validate:"omitempty,gt=0"`
	Where bson.M               `json:"where"`
	Sort  [][]interface{}      `json:"sort" validate:"omitempty"`
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
	return fiber.Map{
		"value": data,
	}
}

// FindByPageDto Get the request body of the paged list resource
type FindByPageDto struct {
	Where      bson.M     `json:"where"`
	Sort       bson.M     `json:"sort"`
	Pagination Pagination `json:"page" validate:"required"`
}

type Pagination struct {
	Index int64 `json:"index" validate:"required,gt=0,number"`
	Size  int64 `json:"size" validate:"required,oneof=10 20 50 100"`
}

type FindByPageResult struct {
	Value []map[string]interface{} `json:"value"`
	Total int64                    `json:"total"`
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

// UpdateDto Update resource request body
type UpdateDto struct {
	Id     primitive.ObjectID `json:"id" validate:"required_without=Where"`
	Where  bson.M             `json:"where" validate:"required_without=Id"`
	Update bson.M             `json:"update" validate:"required"`
	Refs   []string           `json:"refs"`
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

// DeleteDto Delete resource request body
type DeleteDto struct {
	Id    []primitive.ObjectID `json:"id" validate:"required_without=Where,omitempty,gt=0"`
	Where bson.M               `json:"where" validate:"required_without=Id,excluded_with=Id"`
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
