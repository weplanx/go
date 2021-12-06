package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindByPageBody Get the request body of the paged list resource
type FindByPageBody struct {
	Where      bson.M     `json:"where"`
	Sort       bson.M     `json:"sort"`
	Pagination Pagination `json:"page" validate:"required"`
}

type Pagination struct {
	Index int64 `json:"index" validate:"required,gt=0,number"`
	Size  int64 `json:"size" validate:"required,oneof=10 20 50 100"`
}

// FindByPage Get paging list resources
func (x *API) FindByPage(c *fiber.Ctx) interface{} {
	ctx := c.UserContext()
	var body FindByPageBody
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	if err := validator.New().Struct(body); err != nil {
		return err
	}
	name := x.collectionName(c)
	var total int64
	var err error
	if len(body.Where) != 0 {
		if total, err = x.Db.Collection(name).CountDocuments(ctx, body.Where); err != nil {
			return err
		}
	} else {
		if total, err = x.Db.Collection(name).EstimatedDocumentCount(ctx); err != nil {
			return err
		}
	}
	opts := options.Find()
	page := body.Pagination
	if len(body.Sort) != 0 {
		var sorts bson.D
		for k, v := range body.Sort {
			sorts = append(sorts, bson.E{Key: k, Value: v})
		}
		opts.SetSort(sorts)
		opts.SetAllowDiskUse(true)
	}
	opts.SetLimit(page.Size)
	opts.SetSkip((page.Index - 1) * page.Size)
	cursor, err := x.Db.Collection(name).Find(ctx, body.Where, opts)
	if err != nil {
		return err
	}
	value := make([]map[string]interface{}, page.Size)
	if err := cursor.All(ctx, &value); err != nil {
		return err
	}
	return fiber.Map{
		"value": value,
		"total": total,
	}
}
