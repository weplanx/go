package api

import (
	"github.com/gin-gonic/gin"
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

type Uri struct {
	Name string `json:"name" binding:"required"`
}

func (x *Controller) setCollectionName(c *gin.Context) (err error) {
	if x.PATH != "" {
		c.Set("collection", x.PATH)
		return
	}
	var uri Uri
	if err = c.ShouldBindUri(&uri); err != nil {
		return
	}
	c.Set("collection", uri.Name)
	return
}

// Create resources
func (x *Controller) Create(c *gin.Context) interface{} {
	var body bson.M
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.setCollectionName(c); err != nil {
		return err
	}
	result, err := x.API.Create(c.Request.Context(), &body)
	if err != nil {
		return err
	}
	return result
}

// FindOneDto Get a single resource request body
type FindOneDto struct {
	Id    primitive.ObjectID `json:"id" binding:"required_without=Where"`
	Where bson.M             `json:"where" binding:"required_without=Id,excluded_with=Id"`
}

// FindOne Get a single resource
func (x *Controller) FindOne(c *gin.Context) interface{} {
	var body FindOneDto
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.setCollectionName(c); err != nil {
		return err
	}
	data := make(map[string]interface{})
	if err := x.API.FindOne(c.Request.Context(), &body, &data); err != nil {
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

func (x *Controller) Find(c *gin.Context) interface{} {
	var body FindDto
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.setCollectionName(c); err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0)
	if err := x.API.Find(c.Request.Context(), &body, &data); err != nil {
		return err
	}
	return gin.H{
		"data": data,
	}
}

// FindByPageDto Get the request body of the paged list resource
type FindByPageDto struct {
	Where      bson.M     `json:"where"`
	Sort       bson.M     `json:"sort"`
	Pagination Pagination `json:"page" binding:"required"`
}

type Pagination struct {
	Index int64 `json:"index" binding:"required,gt=0,number"`
	Size  int64 `json:"size" binding:"required,oneof=10 20 50 100"`
}

type FindByPageResult struct {
	Value []map[string]interface{} `json:"value"`
	Total int64                    `json:"total"`
}

// FindByPage Get paging list resources
func (x *Controller) FindByPage(c *gin.Context) interface{} {
	var body FindByPageDto
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.setCollectionName(c); err != nil {
		return err
	}
	result, err := x.API.FindByPage(c.Request.Context(), &body)
	if err != nil {
		return err
	}
	return result
}

// UpdateDto Update resource request body
type UpdateDto struct {
	Id     primitive.ObjectID `json:"id" binding:"required_without=Where"`
	Where  bson.M             `json:"where" binding:"required_without=Id"`
	Update bson.M             `json:"update" binding:"required"`
	Refs   []string           `json:"refs"`
}

// Update resources
func (x *Controller) Update(c *gin.Context) interface{} {
	var body UpdateDto
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.setCollectionName(c); err != nil {
		return err
	}
	result, err := x.API.Update(c.Request.Context(), &body)
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
func (x *Controller) Delete(c *gin.Context) interface{} {
	var body DeleteDto
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.setCollectionName(c); err != nil {
		return err
	}
	result, err := x.API.Delete(c.Request.Context(), &body)
	if err != nil {
		return err
	}
	return result
}
