package engine

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"github.com/weplanx/go/password"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Engine struct {
	App     string
	Js      nats.JetStreamContext
	Options map[string]Option
}

type Option struct {
	Event      bool   `yaml:"event"`
	Projection bson.M `yaml:"projection"`
}

type OptionFunc func(engine *Engine)

func SetApp(name string) OptionFunc {
	return func(engine *Engine) {
		engine.App = name
	}
}

func UseStaticOptions(options map[string]Option) OptionFunc {
	return func(engine *Engine) {
		engine.Options = options
	}
}

func UseEvents(js nats.JetStreamContext) OptionFunc {
	return func(engine *Engine) {
		for k, v := range engine.Options {
			if v.Event {
				name := fmt.Sprintf(`%s:events:%s`, engine.App, k)
				subject := fmt.Sprintf(`%s.events.%s`, engine.App, k)
				if _, err := js.AddStream(&nats.StreamConfig{
					Name:      name,
					Subjects:  []string{subject},
					Retention: nats.WorkQueuePolicy,
				}); err != nil {
					log.Fatalln(err)
				}
			}
		}
		engine.Js = js
	}
}

func New(options ...OptionFunc) *Engine {
	x := &Engine{App: ""}
	for _, v := range options {
		v(x)
	}
	return x
}

type Pagination struct {
	Index int64 `header:"x-page" binding:"omitempty,gt=0,number"`
	Size  int64 `header:"x-page-size" binding:"omitempty,number"`
}

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)

type Controller struct {
	Engine  *Engine
	Service *Service
}

type CommonParams struct {
	Model string `uri:"model" binding:"omitempty,key"`
	Id    string `uri:"id" binding:"omitempty,objectId"`
}

func (x *Controller) Params(c *gin.Context) (params *CommonParams, err error) {
	if err = c.ShouldBindUri(&params); err != nil {
		return
	}
	if value, exists := c.Get("model"); exists {
		params.Model = value.(string)
	}
	return
}

type CreateBody struct {
	Doc    map[string]interface{}   `json:"doc" binding:"required_without=Docs"`
	Docs   []map[string]interface{} `json:"docs" binding:"required_without=Doc,excluded_with=Doc,dive,gt=0"`
	Format map[string]interface{}   `json:"format" binding:"omitempty,dive,gt=0"`
	Ref    []string                 `json:"ref" binding:"omitempty,dive,gt=0"`
}

// Create 创建文档
func (x *Controller) Create(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	var body CreateBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	var result interface{}
	if len(body.Docs) != 0 {
		if result, err = x.Service.InsertMany(ctx,
			params.Model, body.Docs, body.Format, body.Ref,
		); err != nil {
			return err
		}
	} else {
		if result, err = x.Service.InsertOne(ctx,
			params.Model, body.Doc, body.Format, body.Ref,
		); err != nil {
			return err
		}
	}
	c.Set("status_code", http.StatusCreated)
	if err = x.Service.Event(ctx, params.Model, result); err != nil {
		return err
	}
	return result
}

type FindQuery struct {
	Id     []string               `form:"id" binding:"omitempty,excluded_with=Where Single,dive,objectId"`
	Where  map[string]interface{} `form:"where" binding:"omitempty,excluded_with=Id"`
	Single bool                   `form:"single"`
	Sort   []string               `form:"sort" binding:"omitempty,dive,gt=0,sort"`
}

// Find 通过获取多个文档
func (x *Controller) Find(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	var page Pagination
	if err = c.ShouldBindHeader(&page); err != nil {
		return err
	}
	var query FindQuery
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	ctx := c.Request.Context()
	if query.Single == true {
		result, err := x.Service.FindOne(ctx, params.Model, query.Where)
		if err != nil {
			return err
		}
		return result
	}
	if len(query.Id) != 0 {
		result, err := x.Service.FindById(ctx, params.Model, query.Id, query.Sort)
		if err != nil {
			return err
		}
		return result
	}
	if page.Index != 0 && page.Size != 0 {
		result, err := x.Service.FindByPage(ctx, params.Model, page, query.Where, query.Sort)
		if err != nil {
			return err
		}
		c.Header("x-page-total", strconv.FormatInt(result.Total, 10))
		return result.Data
	}
	result, err := x.Service.Find(ctx, params.Model, query.Where, query.Sort)
	if err != nil {
		return err
	}
	return result
}

// FindOneById 通过 ID 获取单个文档
func (x *Controller) FindOneById(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	result, err := x.Service.FindOneById(c.Request.Context(), params.Model, params.Id)
	if err != nil {
		return err
	}
	return result
}

type UpdateQuery struct {
	Id     []string               `form:"id" binding:"required_without=Where,excluded_with=Multiple,dive,objectId"`
	Where  map[string]interface{} `form:"where" binding:"required_without=Id,excluded_with=Id"`
	Single bool                   `form:"single"`
}

type UpdateBody struct {
	Update map[string]interface{} `json:"update" binding:"required"`
	Format map[string]interface{} `json:"format" binding:"omitempty,dive,gt=0"`
	Ref    []string               `json:"ref" binding:"omitempty,dive,gt=0"`
}

// Update 更新文档
func (x *Controller) Update(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	var query UpdateQuery
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	var body UpdateBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	if len(query.Id) != 0 {
		result, err := x.Service.
			UpdateManyById(ctx, params.Model, query.Id, body.Update, body.Format, body.Ref)
		if err != nil {
			return err
		}
		return result
	}
	if !query.Single {
		result, err := x.Service.
			UpdateMany(ctx, params.Model, query.Where, body.Update, body.Format, body.Ref)
		if err != nil {
			return err
		}
		return result
	}
	result, err := x.Service.
		UpdateOne(ctx, params.Model, query.Where, body.Update, body.Format, body.Ref)
	if err != nil {
		return err
	}
	if err = x.Service.Event(ctx, params.Model, result); err != nil {
		return err
	}
	return result
}

func (x *Controller) UpdateOneById(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	var body UpdateBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	result, err := x.Service.
		UpdateOneById(ctx, params.Model, params.Id, body.Update, body.Format, body.Ref)
	if err != nil {
		return err
	}
	if err = x.Service.Event(ctx, params.Model, result); err != nil {
		return err
	}
	return result
}

type ReplaceOneBody struct {
	Doc    map[string]interface{} `json:"doc" binding:"required"`
	Format map[string]interface{} `json:"format" binding:"omitempty,dive,gt=0"`
	Ref    []string               `json:"ref" binding:"omitempty,dive,gt=0"`
}

func (x *Controller) ReplaceOneById(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	var body ReplaceOneBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	result, err := x.Service.ReplaceOneById(ctx, params.Model, params.Id, body.Doc, body.Format, body.Ref)
	if err != nil {
		return err
	}
	if err = x.Service.Event(ctx, params.Model, result); err != nil {
		return err
	}
	return result
}

func (x *Controller) DeleteOneById(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	ctx := c.Request.Context()
	result, err := x.Service.DeleteOneById(ctx, params.Model, params.Id)
	if err != nil {
		return err
	}
	if err = x.Service.Event(ctx, params.Model, result); err != nil {
		return err
	}
	return result
}

type Service struct {
	Engine *Engine
	Db     *mongo.Database
}

// Event 发送事件
func (x *Service) Event(ctx context.Context, model string, data interface{}) (err error) {
	if option, ok := x.Engine.Options[model]; ok {
		if !option.Event {
			return
		}
		var payload []byte
		if payload, err = jsoniter.Marshal(data); err != nil {
			return
		}
		subject := fmt.Sprintf(`%s.events.%s`, x.Engine.App, model)
		if _, err = x.Engine.Js.Publish(subject, payload, nats.Context(ctx)); err != nil {
			return
		}
	}
	return
}

func (x *Service) setFormat(formats map[string]interface{}, v interface{}) (err error) {
	doc, _ := v.(map[string]interface{})
	for key, format := range formats {
		if _, ok := doc[key]; !ok {
			continue
		}
		switch format {
		case "object_id":
			if doc[key], err = primitive.
				ObjectIDFromHex(doc[key].(string)); err != nil {
				return
			}
			break
		case "password":
			if doc[key], err = password.Create(doc[key].(string)); err != nil {
				return
			}
			break
		}
	}
	return
}

func (x *Service) setRef(refs []string, v interface{}) (err error) {
	doc, _ := v.(map[string]interface{})
	for _, ref := range refs {
		if _, ok := doc[ref]; !ok {
			continue
		}
		for i, id := range doc[ref].([]interface{}) {
			if doc[ref].([]interface{})[i], err = primitive.
				ObjectIDFromHex(id.(string)); err != nil {
				return
			}
		}
	}
	return
}

func (x *Service) InsertOne(
	ctx context.Context,
	model string,
	doc map[string]interface{},
	format map[string]interface{},
	ref []string,
) (result *mongo.InsertOneResult, err error) {
	if err = x.setFormat(format, doc); err != nil {
		return
	}
	if err = x.setRef(ref, doc); err != nil {
		return
	}
	doc["create_time"] = time.Now()
	doc["update_time"] = time.Now()
	return x.Db.Collection(model).InsertOne(ctx, doc)
}

func (x *Service) InsertMany(
	ctx context.Context,
	model string,
	docs []map[string]interface{},
	format map[string]interface{},
	ref []string,
) (result *mongo.InsertManyResult, err error) {
	data := make([]interface{}, len(docs))
	for i, doc := range docs {
		if err = x.setFormat(format, doc); err != nil {
			return
		}
		if err = x.setRef(ref, doc); err != nil {
			return
		}
		doc["create_time"] = time.Now()
		doc["update_time"] = time.Now()
		data[i] = doc
	}
	return x.Db.Collection(model).InsertMany(ctx, data)
}

func (x *Service) Find(
	ctx context.Context,
	model string,
	filter bson.M,
	sort []string,
	opts ...*options.FindOptions,
) (data []map[string]interface{}, err error) {
	option := options.Find()
	if len(sort) != 0 {
		sorts := make(bson.D, len(sort))
		for i, x := range sort {
			v := strings.Split(x, ".")
			var direction int
			if direction, err = strconv.Atoi(v[1]); err != nil {
				return
			}
			sorts[i] = bson.E{Key: v[0], Value: direction}
		}
		option.SetSort(sorts)
		option.SetAllowDiskUse(true)
	} else {
		option.SetSort(bson.M{"_id": -1})
	}
	if staticOpt, ok := x.Engine.Options[model]; ok {
		if len(staticOpt.Projection) != 0 {
			option.SetProjection(staticOpt.Projection)
		}
	}
	opts = append(opts, option)
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection(model).Find(ctx, filter, opts...); err != nil {
		return
	}
	data = make([]map[string]interface{}, 0)
	if err = cursor.All(ctx, &data); err != nil {
		return
	}
	return
}

func (x *Service) FindById(
	ctx context.Context,
	model string,
	ids []string,
	sort []string,
) (data []map[string]interface{}, err error) {
	oids := make([]primitive.ObjectID, len(ids))
	for i, v := range ids {
		oids[i], _ = primitive.ObjectIDFromHex(v)
	}
	return x.Find(ctx, model, bson.M{"_id": bson.M{"$in": oids}}, sort)
}

type FindResult struct {
	Total int64                    `json:"total"`
	Data  []map[string]interface{} `json:"data"`
}

func (x *Service) FindByPage(
	ctx context.Context,
	model string,
	page Pagination,
	filter map[string]interface{},
	sort []string,
) (result FindResult, err error) {
	if len(filter) != 0 {
		if result.Total, err = x.Db.Collection(model).CountDocuments(ctx, filter); err != nil {
			return
		}
	} else {
		if result.Total, err = x.Db.Collection(model).EstimatedDocumentCount(ctx); err != nil {
			return
		}
	}
	option := options.Find()
	option.SetLimit(page.Size)
	option.SetSkip((page.Index - 1) * page.Size)
	if result.Data, err = x.Find(ctx, model, filter, sort, option); err != nil {
		return
	}
	return
}

func (x *Service) FindOne(
	ctx context.Context,
	model string,
	filter bson.M,
) (data map[string]interface{}, err error) {
	option := options.FindOne()
	if staticOpt, ok := x.Engine.Options[model]; ok {
		if len(staticOpt.Projection) != 0 {
			option.SetProjection(staticOpt.Projection)
		}
	}
	if err = x.Db.Collection(model).FindOne(ctx, filter, option).Decode(&data); err != nil {
		return
	}
	return
}

func (x *Service) FindOneById(
	ctx context.Context,
	model string,
	id string,
) (data map[string]interface{}, err error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	return x.FindOne(ctx, model, bson.M{"_id": oid})
}

func (x *Service) UpdateMany(ctx context.Context,
	model string,
	filter bson.M,
	update map[string]interface{},
	format map[string]interface{},
	ref []string,
) (result *mongo.UpdateResult, err error) {
	if update["$set"] != nil {
		if err = x.setFormat(format, update["$set"]); err != nil {
			return
		}
		if err = x.setRef(ref, update["$set"]); err != nil {
			return
		}
		update["$set"].(map[string]interface{})["update_time"] = time.Now()
	}
	return x.Db.Collection(model).UpdateMany(ctx, filter, update)
}

func (x *Service) UpdateManyById(
	ctx context.Context,
	model string,
	ids []string,
	update map[string]interface{},
	format map[string]interface{},
	ref []string,
) (result *mongo.UpdateResult, err error) {
	oids := make([]primitive.ObjectID, len(ids))
	for i, v := range ids {
		oids[i], _ = primitive.ObjectIDFromHex(v)
	}
	return x.UpdateMany(ctx, model, bson.M{"_id": bson.M{"$in": oids}}, update, format, ref)
}

func (x *Service) UpdateOne(
	ctx context.Context,
	model string,
	filter bson.M,
	update map[string]interface{},
	format map[string]interface{},
	ref []string,
) (result *mongo.UpdateResult, err error) {
	if update["$set"] != nil {
		if err = x.setFormat(format, update["$set"]); err != nil {
			return
		}
		if err = x.setRef(ref, update["$set"]); err != nil {
			return
		}
		update["$set"].(map[string]interface{})["update_time"] = time.Now()
	}
	return x.Db.Collection(model).UpdateOne(ctx, filter, update)
}

func (x *Service) UpdateOneById(
	ctx context.Context,
	model string,
	id string,
	update map[string]interface{},
	format map[string]interface{},
	ref []string,
) (result *mongo.UpdateResult, err error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	return x.UpdateOne(ctx, model, bson.M{"_id": oid}, update, format, ref)
}

func (x *Service) ReplaceOneById(
	ctx context.Context,
	model string,
	id string,
	doc map[string]interface{},
	format map[string]interface{},
	ref []string,
) (result *mongo.UpdateResult, err error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": oid}
	if err = x.setFormat(format, doc); err != nil {
		return
	}
	if err = x.setRef(ref, doc); err != nil {
		return
	}
	doc["create_time"] = time.Now()
	doc["update_time"] = time.Now()
	return x.Db.Collection(model).ReplaceOne(ctx, filter, doc)
}

func (x *Service) DeleteOneById(
	ctx context.Context,
	model string,
	id string,
) (result *mongo.DeleteResult, err error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	return x.Db.Collection(model).DeleteOne(ctx, bson.M{"_id": oid})
}
