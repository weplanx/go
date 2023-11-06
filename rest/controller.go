package rest

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/weplanx/go/help"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"strings"
	"time"
)

type Controller struct {
	*Service
}

type CreateDto struct {
	Collection string `path:"collection" vd:"snake"`
	Data       M      `json:"data" vd:"gt=0"`
	Xdata      M      `json:"xdata，omitempty"`
	Txn        string `json:"txn，omitempty" vd:"omitempty,uuid"`
}

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
		c.Error(help.E("rest.transform", err.Error()))
		return
	}
	dto.Data["create_time"] = time.Now()
	dto.Data["update_time"] = time.Now()

	if dto.Txn != "" {
		if err := x.Service.Pending(ctx, dto.Txn, PendingDto{
			Action: ActionCreate,
			Name:   dto.Collection,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(204)
		return
	}

	r, err := x.Service.Create(ctx, dto.Collection, dto.Data)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(201, r)
}

type BulkCreateDto struct {
	Collection string `path:"collection" vd:"snake"`
	Data       []M    `json:"data" vd:"gt=0"`
	Xdata      M      `json:"xdata,omitempty"`
	Txn        string `json:"txn,omitempty" vd:"omitempty,uuid"`
}

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
			Action: ActionBulkCreate,
			Name:   dto.Collection,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(204)
		return
	}

	r, err := x.Service.BulkCreate(ctx, dto.Collection, docs)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(201, r)
}

type SizeDto struct {
	Collection string `path:"collection" vd:"snake"`
	Filter     M      `json:"filter" vd:"required"`
	Xfilter    M      `json:"xfilter,omitempty"`
}

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
	c.Status(204)
}

type FindDto struct {
	Collection string   `path:"collection" vd:"snake"`
	Pagesize   int64    `header:"x-pagesize" vd:"omitempty,min=0,max=1000"`
	Page       int64    `header:"x-page" vd:"omitempty,min=0"`
	Filter     M        `json:"filter" vd:"required"`
	Xfilter    M        `json:"xfilter,omitempty"`
	Sort       []string `query:"sort,omitempty" vd:"omitempty,dive,sort"`
	Keys       []string `query:"keys,omitempty"`
}

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
	c.JSON(200, data)
}

type FindOneDto struct {
	Collection string   `path:"collection" vd:"snake"`
	Filter     M        `json:"filter" vd:"gt=0"`
	Xfilter    M        `json:"xfilter,omitempty"`
	Keys       []string `query:"keys,omitempty"`
}

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

	c.JSON(200, data)
}

type FindByIdDto struct {
	Collection string   `path:"collection" vd:"snake"`
	Id         string   `path:"id" vd:"mongodb"`
	Keys       []string `query:"keys,omitempty"`
}

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

	c.JSON(200, data)
}

type UpdateDto struct {
	Collection   string        `path:"collection" vd:"snake"`
	Filter       M             `json:"filter" vd:"gt=0"`
	Xfilter      M             `json:"xfilter,omitempty"`
	Data         M             `json:"data" vd:"gt=0"`
	Xdata        M             `json:"xdata,omitempty"`
	ArrayFilters []interface{} `json:"arrayFilters,omitempty"`
	Txn          string        `json:"txn,omitempty" vd:"omitempty,uuid"`
}

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
	opt := options.Update()
	if dto.ArrayFilters != nil {
		opt = opt.SetArrayFilters(options.ArrayFilters{Filters: dto.ArrayFilters})
	}

	if dto.Txn != "" {
		if err := x.Service.Pending(ctx, dto.Txn, PendingDto{
			Action: ActionUpdate,
			Name:   dto.Collection,
			Filter: dto.Filter,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(204)
		return
	}

	r, err := x.Service.Update(ctx, dto.Collection, dto.Filter, dto.Data, opt)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, r)
}

type UpdateByIdDto struct {
	Collection   string        `path:"collection" vd:"snake"`
	Id           string        `path:"id" vd:"mongodb"`
	Data         M             `json:"data" vd:"gt=0"`
	Xdata        M             `json:"xdata,omitempty"`
	ArrayFilters []interface{} `json:"arrayFilters,omitempty"`
	Txn          string        `json:"txn,omitempty" vd:"omitempty,uuid"`
}

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
	opt := options.Update()
	if dto.ArrayFilters != nil {
		opt = opt.SetArrayFilters(options.ArrayFilters{Filters: dto.ArrayFilters})
	}

	if dto.Txn != "" {
		if err := x.Service.Pending(ctx, dto.Txn, PendingDto{
			Action: ActionUpdateById,
			Name:   dto.Collection,
			Id:     id,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(204)
		return
	}

	r, err := x.Service.UpdateById(ctx, dto.Collection, id, dto.Data, opt)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, r)
}

type ReplaceDto struct {
	Collection string `path:"collection" vd:"snake"`
	Id         string `path:"id" vd:"mongodb"`
	Data       M      `json:"data" vd:"gt=0"`
	Xdata      M      `json:"xdata,omitempty"`
	Txn        string `json:"txn,omitempty" vd:"omitempty,uuid"`
}

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
			Action: ActionReplace,
			Name:   dto.Collection,
			Id:     id,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(204)
		return
	}

	r, err := x.Service.Replace(ctx, dto.Collection, id, dto.Data)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, r)
}

type DeleteDto struct {
	Collection string `path:"collection" vd:"snake"`
	Id         string `path:"id" vd:"mongodb"`
	Txn        string `query:"txn,omitempty" vd:"omitempty,uuid"`
}

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
			Action: ActionDelete,
			Name:   dto.Collection,
			Id:     id,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(204)
		return
	}

	r, err := x.Service.Delete(ctx, dto.Collection, id, false)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, r)
}

type BulkDeleteDto struct {
	Collection string `path:"collection" vd:"snake"`
	Filter     M      `json:"filter" vd:"gt=0"`
	Xfilter    M      `json:"xfilter,omitempty"`
	Txn        string `json:"txn,omitempty" vd:"omitempty,uuid"`
}

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
			Action: ActionBulkDelete,
			Name:   dto.Collection,
			Filter: dto.Filter,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(204)
		return
	}

	r, err := x.Service.BulkDelete(ctx, dto.Collection, dto.Filter, false)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, r)
}

type SortDto struct {
	Collection string      `path:"collection" vd:"snake"`
	Data       SortDtoData `json:"data" vd:"structonly"`
	Txn        string      `json:"txn,omitempty" vd:"omitempty,uuid"`
}

type SortDtoData struct {
	Key    string               `json:"key"  vd:"required"`
	Values []primitive.ObjectID `json:"values"  vd:"gt=0,dive,mongodb"`
}

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
			Action: ActionSort,
			Name:   dto.Collection,
			Data:   dto.Data,
		}); err != nil {
			c.Error(err)
			return
		}

		c.Status(204)
		return
	}

	_, err := x.Service.Sort(ctx, dto.Collection, dto.Data.Key, dto.Data.Values)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(204)
}

func (x *Controller) Transaction(ctx context.Context, c *app.RequestContext) {
	txn := help.Uuid()
	x.Service.Transaction(ctx, txn)
	c.JSON(201, utils.H{
		"txn": txn,
	})
}

type CommitDto struct {
	Txn string `json:"txn" vd:"uuid"`
}

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

	c.JSON(200, r)
}
