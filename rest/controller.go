package rest

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/gookit/goutil/strutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Controller struct {
	*Service
}

type CreateDto struct {
	Collection string `path:"collection" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Data       M      `json:"data,required" vd:"len($)>0;msg:'document cannot be empty data'"`
	Xdata      M      `json:"xdata"`
	Txn        string `json:"txn"`
}

// Create
// @router /:collection/create [POST]
func (x *Controller) Create(ctx context.Context, c *app.RequestContext) {
	var dto CreateDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	if err := x.Service.Transform(dto.Data, dto.Xdata); err != nil {
		c.Error(errors.New(err, errors.ErrorTypePublic, nil))
		return
	}
	dto.Data["create_time"] = time.Now()
	dto.Data["update_time"] = time.Now()

	if dto.Txn != "" {
		if err := x.Service.Pending(ctx, dto.Txn, PendingDto{
			Action: "create",
			Name:   dto.Collection,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusNoContent)
		return
	}

	r, err := x.Service.Create(ctx, dto.Collection, dto.Data)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, r)
}

type BulkCreateDto struct {
	Collection string `path:"collection,required" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Data       []M    `json:"data,required" vd:"len($)>0 && range($,len(#v)>0);msg:'batch documents cannot have empty data'"`
	Xdata      M      `json:"xdata"`
	Txn        string `json:"txn"`
}

// BulkCreate
// @router /:collection/bulk_create [POST]
func (x *Controller) BulkCreate(ctx context.Context, c *app.RequestContext) {
	var dto BulkCreateDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	docs := make([]interface{}, len(dto.Data))
	for i, doc := range dto.Data {
		if err := x.Service.Transform(doc, dto.Xdata); err != nil {
			c.Error(errors.New(err, errors.ErrorTypePublic, nil))
			return
		}
		doc["create_time"] = time.Now()
		doc["update_time"] = time.Now()
		docs[i] = doc
	}

	if dto.Txn != "" {
		if err := x.Service.Pending(ctx, dto.Txn, PendingDto{
			Action: "bulk_create",
			Name:   dto.Collection,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusNoContent)
		return
	}

	r, err := x.Service.BulkCreate(ctx, dto.Collection, docs)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, r)
}

type SizeDto struct {
	Collection string `path:"collection,required" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Filter     M      `json:"filter,required"`
	Xfilter    M      `json:"xfilter"`
}

// Size
// @router /:collection/size [POST]
func (x *Controller) Size(ctx context.Context, c *app.RequestContext) {
	var dto SizeDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	if err := x.Service.Transform(dto.Filter, dto.Xfilter); err != nil {
		c.Error(errors.New(err, errors.ErrorTypePublic, nil))
		return
	}

	size, err := x.Service.Size(ctx, dto.Collection, dto.Filter)
	if err != nil {
		c.Error(err)
		return
	}

	c.Header("x-total", strconv.Itoa(int(size)))
	c.Status(http.StatusNoContent)
}

type FindDto struct {
	Collection string   `path:"collection,required" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Pagesize   int64    `header:"x-pagesize" vd:"$>=0 && $<=1000;msg:'the number of pages must be between 1 and 1000'"`
	Page       int64    `header:"x-page" vd:"$>=0;msg:'the page number must be greater than 0'"`
	Filter     M        `json:"filter,required"`
	Xfilter    M        `json:"xfilter"`
	Sort       []string `query:"sort" vd:"range($,regexp('^[a-z_]+:(-1|1)$',#v)));msg:'the collation is not standardized'"`
	Keys       []string `query:"keys" vd:"range($,regexp('^[a-z_]+$',#v));msg:'the projection rules are not standardized'"`
}

// Find
// @router /:collection/find [POST]
func (x *Controller) Find(ctx context.Context, c *app.RequestContext) {
	var dto FindDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	if err := x.Service.Transform(dto.Filter, dto.Xfilter); err != nil {
		c.Error(errors.New(err, errors.ErrorTypePublic, nil))
		return
	}

	size, err := x.Service.Size(ctx, dto.Collection, dto.Filter)
	if err != nil {
		c.Error(err)
		return
	}

	if dto.Pagesize == 0 {
		dto.Pagesize = 100
	}

	if dto.Page == 0 {
		dto.Page = 1
	}

	sort := make(bson.D, len(dto.Sort))
	for i, v := range dto.Sort {
		rule := strings.Split(v, ":")
		order, _ := strconv.Atoi(rule[1])
		sort[i] = bson.E{Key: rule[0], Value: order}
	}

	if len(sort) == 0 {
		sort = bson.D{{Key: "_id", Value: -1}}
	}

	option := options.Find().
		SetProjection(x.Service.Projection(dto.Collection, dto.Keys)).
		SetLimit(dto.Pagesize).
		SetSkip((dto.Page - 1) * dto.Pagesize).
		SetSort(sort).
		SetAllowDiskUse(true)

	data, err := x.Service.Find(ctx, dto.Collection, dto.Filter, option)
	if err != nil {
		c.Error(err)
		return
	}

	c.Header("x-total", strconv.Itoa(int(size)))
	c.JSON(http.StatusOK, data)
}

type FindOneDto struct {
	Collection string   `path:"collection,required" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Filter     M        `json:"filter,required" vd:"len($)>0;msg:'the filter cannot be empty'"`
	Xfilter    M        `json:"xfilter"`
	Keys       []string `query:"keys" vd:"range($,regexp('^[a-z_]+$',#v));msg:'the projection rules are not standardized'"`
}

// FindOne
// @router /:collection/fine_one [POST]
func (x *Controller) FindOne(ctx context.Context, c *app.RequestContext) {
	var dto FindOneDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	if err := x.Service.Transform(dto.Filter, dto.Xfilter); err != nil {
		c.Error(errors.New(err, errors.ErrorTypePublic, nil))
		return
	}

	option := options.FindOne().
		SetProjection(x.Service.Projection(dto.Collection, dto.Keys))

	data, err := x.Service.FindOne(ctx, dto.Collection, dto.Filter, option)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, data)
}

type FindByIdDto struct {
	Collection string   `path:"collection,required" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Id         string   `path:"id,required" vd:"mongoId($);msg:'the document id must be an ObjectId'"`
	Keys       []string `query:"keys" vd:"range($,regexp('^[a-z_]+$',#v));msg:'the projection rules are not standardized'"`
}

// FindById
// @router /:collection/:id [GET]
func (x *Controller) FindById(ctx context.Context, c *app.RequestContext) {
	var dto FindByIdDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	id, _ := primitive.ObjectIDFromHex(dto.Id)
	option := options.FindOne().
		SetProjection(x.Service.Projection(dto.Collection, dto.Keys))

	data, err := x.Service.FindOne(ctx, dto.Collection, M{"_id": id}, option)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, data)
}

type UpdateDto struct {
	Collection string `path:"collection" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Filter     M      `json:"filter,required" vd:"len($)>0;msg:'the filter cannot be empty'"`
	Xfilter    M      `json:"xfilter"`
	Data       M      `json:"data,required" vd:"len($)>0;msg:'the update cannot be empty'"`
	Xdata      M      `json:"xdata"`
	Txn        string `json:"txn"`
}

// Update
// @router /:collection/update [POST]
func (x *Controller) Update(ctx context.Context, c *app.RequestContext) {
	var dto UpdateDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	if err := x.Service.Transform(dto.Filter, dto.Xfilter); err != nil {
		c.Error(errors.New(err, errors.ErrorTypePublic, nil))
		return
	}
	if err := x.Service.Transform(dto.Data, dto.Xdata); err != nil {
		c.Error(errors.New(err, errors.ErrorTypePublic, nil))
		return
	}
	if _, ok := dto.Data["$set"]; !ok {
		dto.Data["$set"] = M{}
	}
	dto.Data["$set"].(M)["update_time"] = time.Now()

	if dto.Txn != "" {
		if err := x.Service.Pending(ctx, dto.Txn, PendingDto{
			Action: "update",
			Name:   dto.Collection,
			Filter: dto.Filter,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusNoContent)
		return
	}

	r, err := x.Service.Update(ctx, dto.Collection, dto.Filter, dto.Data)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, r)
}

type UpdateByIdDto struct {
	Collection string `path:"collection,required" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Id         string `path:"id,required" vd:"mongoId($);msg:'the document id must be an ObjectId'"`
	Data       M      `json:"data,required" vd:"len($)>0;msg:'the update cannot be empty'"`
	Xdata      M      `json:"xdata"`
	Txn        string `json:"txn"`
}

// UpdateById
// @router /:collection/:id [PATCH]
func (x *Controller) UpdateById(ctx context.Context, c *app.RequestContext) {
	var dto UpdateByIdDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	if err := x.Service.Transform(dto.Data, dto.Xdata); err != nil {
		c.Error(errors.New(err, errors.ErrorTypePublic, nil))
		return
	}
	if _, ok := dto.Data["$set"]; !ok {
		dto.Data["$set"] = M{}
	}
	dto.Data["$set"].(M)["update_time"] = time.Now()
	id, _ := primitive.ObjectIDFromHex(dto.Id)

	if dto.Txn != "" {
		if err := x.Service.Pending(ctx, dto.Txn, PendingDto{
			Action: "update_by_id",
			Name:   dto.Collection,
			Id:     id,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusNoContent)
		return
	}

	r, err := x.Service.UpdateById(ctx, dto.Collection, id, dto.Data)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, r)
}

type ReplaceDto struct {
	Collection string `path:"collection,required" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Id         string `path:"id,required" vd:"mongoId($);msg:'the document id must be an ObjectId'"`
	Data       M      `json:"data,required" vd:"len($)>0;msg:'document cannot be empty data'"`
	Xdata      M      `json:"xdata"`
	Txn        string `json:"txn"`
}

// Replace
// @router /:collection/:id [PUT]
func (x *Controller) Replace(ctx context.Context, c *app.RequestContext) {
	var dto ReplaceDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	if err := x.Service.Transform(dto.Data, dto.Xdata); err != nil {
		c.Error(errors.New(err, errors.ErrorTypePublic, nil))
		return
	}
	dto.Data["create_time"] = time.Now()
	dto.Data["update_time"] = time.Now()
	id, _ := primitive.ObjectIDFromHex(dto.Id)

	if dto.Txn != "" {
		if err := x.Service.Pending(ctx, dto.Txn, PendingDto{
			Action: "replace",
			Name:   dto.Collection,
			Id:     id,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusNoContent)
		return
	}

	r, err := x.Service.Replace(ctx, dto.Collection, id, dto.Data)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, r)
}

type DeleteDto struct {
	Collection string `path:"collection,required" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Id         string `path:"id,required" vd:"mongoId($);msg:'the document id must be an ObjectId'"`
	Txn        string `query:"txn"`
}

// Delete
// @router /:collection/:id [DELETE]
func (x *Controller) Delete(ctx context.Context, c *app.RequestContext) {
	var dto DeleteDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	id, _ := primitive.ObjectIDFromHex(dto.Id)

	if dto.Txn != "" {
		if err := x.Service.Pending(ctx, dto.Txn, PendingDto{
			Action: "delete",
			Name:   dto.Collection,
			Id:     id,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusNoContent)
		return
	}

	r, err := x.Service.Delete(ctx, dto.Collection, id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, r)
}

type BulkDeleteDto struct {
	Collection string `path:"collection,required" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Filter     M      `json:"filter,required" vd:"len($)>0;msg:'the filter cannot be empty'"`
	Xfilter    M      `json:"xfilter"`
	Txn        string `json:"txn"`
}

// BulkDelete
// @router /:collection/bulk_delete [POST]
func (x *Controller) BulkDelete(ctx context.Context, c *app.RequestContext) {
	var dto BulkDeleteDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	if err := x.Service.Transform(dto.Filter, dto.Xfilter); err != nil {
		c.Error(errors.New(err, errors.ErrorTypePublic, nil))
		return
	}

	if dto.Txn != "" {
		if err := x.Service.Pending(ctx, dto.Txn, PendingDto{
			Action: "bulk_delete",
			Name:   dto.Collection,
			Filter: dto.Filter,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusNoContent)
		return
	}

	r, err := x.Service.BulkDelete(ctx, dto.Collection, dto.Filter)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, r)
}

type SortDto struct {
	Collection string      `path:"collection,required" vd:"regexp('^[a-z_]+$');msg:'the collection name must be lowercase letters with underscores'"`
	Data       SortDtoData `json:"data,required"`
	Txn        string      `json:"txn"`
}

type SortDtoData struct {
	Key    string               `json:"key,required"`
	Values []primitive.ObjectID `json:"values,required" vd:"len($)>0;msg:'the submission data must be an array of ObjectId'"`
}

// Sort
// @router /:collection/sort [POST]
func (x *Controller) Sort(ctx context.Context, c *app.RequestContext) {
	var dto SortDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if x.IsForbid(dto.Collection) {
		c.Error(ErrCollectionForbidden)
		return
	}

	if dto.Txn != "" {
		if err := x.Service.Pending(ctx, dto.Txn, PendingDto{
			Action: "sort",
			Name:   dto.Collection,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusNoContent)
		return
	}

	_, err := x.Service.Sort(ctx, dto.Collection, dto.Data.Key, dto.Data.Values)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Transaction
// @router /transaction [POST]
func (x *Controller) Transaction(ctx context.Context, c *app.RequestContext) {
	txn := strutil.MicroTimeHexID()
	x.Service.Transaction(ctx, txn)
	c.JSON(http.StatusCreated, utils.H{
		"txn": txn,
	})
}

type CommitDto struct {
	Txn string `json:"txn,required"`
}

// Commit
// @router /commit [POST]
func (x *Controller) Commit(ctx context.Context, c *app.RequestContext) {
	var dto CommitDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	r, err := x.Service.Commit(ctx, dto.Txn)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, r)
}
