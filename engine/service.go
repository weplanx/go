package engine

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"github.com/weplanx/go/password"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	Engine *Engine
	Db     *mongo.Database
}

func (x *Service) InsertOne(ctx context.Context, doc M, format M) (_ interface{}, err error) {
	params := ctx.Value("params").(*Params)
	if err = x.Format(doc, format); err != nil {
		return
	}
	doc["create_time"], doc["update_time"] = time.Now(), time.Now()
	return x.Db.Collection(params.Model).InsertOne(ctx, doc)
}

func (x *Service) InsertMany(ctx context.Context, docs []M, format M) (_ interface{}, err error) {
	params := ctx.Value("params").(*Params)
	data := make([]interface{}, len(docs))
	for i, doc := range docs {
		if err = x.Format(doc, format); err != nil {
			return
		}
		doc["create_time"], doc["update_time"] = time.Now(), time.Now()
		data[i] = doc
	}
	return x.Db.Collection(params.Model).InsertMany(ctx, data)
}

func (x *Service) Find(
	ctx context.Context,
	filter M,
	orders []string,
	fields []string,
	limit int64,
	skip int64,
	opts ...*options.FindOptions,
) (data []M, err error) {
	params := ctx.Value("params").(*Params)

	option := options.Find().
		SetProjection(x.Project(params.Model, fields)).
		SetSort(bson.M{"_id": -1}).
		SetLimit(100)

	// 自定义排序
	if len(orders) != 0 {
		sort := make(bson.D, len(orders))
		for i, order := range orders {
			v := strings.Split(order, ".")
			sort[i] = bson.E{Key: v[0]}
			if sort[i].Value, err = strconv.Atoi(v[1]); err != nil {
				return
			}
		}
		option.SetSort(sort)
		option.SetAllowDiskUse(true)
	}

	if skip != 0 {
		option.SetSkip(skip)
	}
	if limit != 0 {
		option.SetLimit(limit)
	}

	var cursor *mongo.Cursor
	opts, data = append(opts, option), make([]M, 0)
	if cursor, err = x.Db.Collection(params.Model).
		Find(ctx, filter, opts...); err != nil {
		return
	}
	if err = cursor.All(ctx, &data); err != nil {
		return
	}
	return
}

func (x *Service) FindById(ctx context.Context, ids []string, orders []string, fields []string) ([]M, error) {
	oids := make([]primitive.ObjectID, len(ids))
	for i, v := range ids {
		oids[i], _ = primitive.ObjectIDFromHex(v)
	}
	return x.Find(ctx, bson.M{"_id": bson.M{"$in": oids}}, orders, fields, 0, 0)
}

func (x *Service) FindByPage(
	ctx context.Context,
	filter M,
	orders []string,
	fields []string,
) (data []M, err error) {
	p := ctx.Value("page").(*Pagination)
	if p.Total, err = x.Count(ctx, filter); err != nil {
		return
	}
	option := options.Find().
		SetLimit(p.Size).
		SetSkip((p.Index - 1) * p.Size)
	if data, err = x.Find(ctx, filter, orders, fields, 0, 0, option); err != nil {
		return
	}
	return
}

func (x *Service) FindOne(ctx context.Context, filter M, fields []string) (data M, err error) {
	params := ctx.Value("params").(*Params)
	option := options.FindOne().
		SetProjection(x.Project(params.Model, fields))

	if err = x.Db.Collection(params.Model).
		FindOne(ctx, filter, option).
		Decode(&data); err != nil {
		return
	}
	return
}

func (x *Service) FindOneById(ctx context.Context, fields []string) (M, error) {
	params := ctx.Value("params").(*Params)
	oid, _ := primitive.ObjectIDFromHex(params.Id)
	return x.FindOne(ctx, M{"_id": oid}, fields)
}

func (x *Service) UpdateMany(ctx context.Context, filter M, update M, format M) (_ interface{}, err error) {
	params := ctx.Value("params").(*Params)
	if update["$set"] != nil {
		if err = x.Format(update["$set"].(M), format); err != nil {
			return
		}
		update["$set"].(M)["update_time"] = time.Now()
	}
	return x.Db.Collection(params.Model).UpdateMany(ctx, filter, update)
}

func (x *Service) UpdateManyById(ctx context.Context, ids []string, update M, format M) (interface{}, error) {
	oids := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		oids[i], _ = primitive.ObjectIDFromHex(id)
	}
	return x.UpdateMany(ctx, M{"_id": M{"$in": oids}}, update, format)
}

func (x *Service) UpdateOne(ctx context.Context, update M, format M) (_ interface{}, err error) {
	params := ctx.Value("params").(*Params)
	if update["$set"] != nil {
		if err = x.Format(update["$set"].(M), format); err != nil {
			return
		}
		update["$set"].(M)["update_time"] = time.Now()
	}
	oid, _ := primitive.ObjectIDFromHex(params.Id)
	return x.Db.Collection(params.Model).UpdateOne(ctx, M{"_id": oid}, update)
}

func (x *Service) ReplaceOne(ctx context.Context, doc M, format M) (_ interface{}, err error) {
	params := ctx.Value("params").(*Params)
	oid, _ := primitive.ObjectIDFromHex(params.Id)
	if err = x.Format(doc, format); err != nil {
		return
	}
	doc["create_time"], doc["update_time"] = time.Now(), time.Now()
	return x.Db.Collection(params.Model).ReplaceOne(ctx, M{"_id": oid}, doc)
}

func (x *Service) DeleteMany(ctx context.Context, filter M) (interface{}, error) {
	params := ctx.Value("params").(*Params)
	return x.Db.Collection(params.Model).DeleteMany(ctx, filter)
}

func (x *Service) DeleteManyById(ctx context.Context, oids []primitive.ObjectID) (interface{}, error) {
	return x.DeleteMany(ctx, M{"_id": M{"$in": oids}})
}

func (x *Service) DeleteOne(ctx context.Context) (interface{}, error) {
	params := ctx.Value("params").(*Params)
	oid, _ := primitive.ObjectIDFromHex(params.Id)
	return x.Db.Collection(params.Model).DeleteOne(ctx, M{"_id": oid})
}

func (x *Service) Count(ctx context.Context, filter M) (count int64, err error) {
	params := ctx.Value("params").(*Params)
	if len(filter) != 0 {
		if count, err = x.Db.Collection(params.Model).
			CountDocuments(ctx, filter); err != nil {
			return
		}
	} else {
		if count, err = x.Db.Collection(params.Model).
			EstimatedDocumentCount(ctx); err != nil {
			return
		}
	}
	return
}

func (x *Service) Exists(ctx context.Context, filter M) (result bool, err error) {
	params := ctx.Value("params").(*Params)
	var count int64
	if count, err = x.Db.Collection(params.Model).
		CountDocuments(ctx, filter); err != nil {
		return
	}
	result = count != 0
	return
}

// Format 格式化
func (x *Service) Format(doc M, rules M) (err error) {
	for field, rule := range rules {
		if _, ok := doc[field]; !ok {
			continue
		}
		switch rule {
		case "object_id":
			// 转换为 ObjectId，为空保留
			if id, ok := doc[field].(string); ok {
				if doc[field], err = primitive.ObjectIDFromHex(id); err != nil {
					return
				}
			}
			break

		case "password":
			// 密码类型，转换为 Argon2id
			if doc[field], err = password.Create(doc[field].(string)); err != nil {
				return
			}
			break

		case "ref":
			// 应用类型，转换为 ObjectId 数组
			value := doc[field].([]interface{})
			for i, v := range value {
				if value[i], err = primitive.ObjectIDFromHex(v.(string)); err != nil {
					return
				}
			}
			break
		}
	}
	return
}

// Project 映射
func (x *Service) Project(model string, fields []string) bson.M {
	project, tmp := make(bson.M), make(bson.M)
	if static, ok := x.Engine.Options[model]; ok {
		for _, v := range static.Field {
			tmp[v] = 1
		}
	}
	if len(fields) != 0 {
		for _, v := range fields {
			if len(tmp) != 0 {
				if _, ok := tmp[v]; ok {
					project[v] = 1
				}
			} else {
				project[v] = 1
			}
		}
	}
	if len(project) == 0 && len(tmp) != 0 {
		project = tmp
	}
	return project
}

// Event 发送事件
func (x *Service) Event(ctx context.Context, action string, content interface{}) (err error) {
	params := ctx.Value("params").(*Params)
	if option, ok := x.Engine.Options[params.Model]; ok {
		if !option.Event {
			return
		}
		var payload []byte
		if payload, err = jsoniter.Marshal(M{
			"action":  action,
			"content": content,
		}); err != nil {
			return
		}
		subject := fmt.Sprintf(`%s.events.%s`, x.Engine.App, params.Model)
		if _, err = x.Engine.Js.Publish(subject, payload, nats.Context(ctx)); err != nil {
			return
		}
	}
	return
}
