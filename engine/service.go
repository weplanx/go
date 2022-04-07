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

func (x *Service) Params(ctx context.Context) *Params {
	return ctx.Value("params").(*Params)
}

func (x *Service) InsertOne(ctx context.Context, doc M) (_ interface{}, err error) {
	params := x.Params(ctx)
	if err = x.Format(doc, params.FormatDoc); err != nil {
		return
	}
	doc["create_time"], doc["update_time"] = time.Now(), time.Now()
	return x.Db.Collection(params.Model).InsertOne(ctx, doc)
}

func (x *Service) InsertMany(ctx context.Context, docs []M) (_ interface{}, err error) {
	params := x.Params(ctx)
	data := make([]interface{}, len(docs))
	for i, doc := range docs {
		if err = x.Format(doc, params.FormatDoc); err != nil {
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
	sorts []string,
	fields []string,
	opt *options.FindOptions,
) (data []M, err error) {
	params := x.Params(ctx)
	option := options.Find().
		SetProjection(x.Project(params.Model, fields)).
		SetSort(bson.M{"_id": -1}).
		SetLimit(100)
	// 排序
	if len(sorts) != 0 {
		sortOpt := make(bson.D, len(sorts))
		for i, order := range sorts {
			v := strings.Split(order, ".")
			sortOpt[i] = bson.E{Key: v[0]}
			sortOpt[i].Value, _ = strconv.Atoi(v[1])
		}
		option.SetSort(sortOpt)
		option.SetAllowDiskUse(true)
	}
	// 最大返回数量
	if params.Limit != 0 {
		option.SetLimit(params.Limit)
	}
	// 跳过数量
	if params.Skip != 0 {
		option.SetSkip(params.Skip)
	}
	if err = x.Format(filter, params.FormatFilter); err != nil {
		return
	}
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection(params.Model).Find(ctx, filter, option, opt); err != nil {
		return
	}
	data = make([]M, 0)
	if err = cursor.All(ctx, &data); err != nil {
		return
	}
	return
}

func (x *Service) FindByPage(
	ctx context.Context,
	filter M,
	sorts []string,
	fields []string,
) (data []M, err error) {
	params := x.Params(ctx)
	if params.Total, err = x.CountDocuments(ctx, filter); err != nil {
		return
	}
	option := options.Find().
		SetLimit(params.Size).
		SetSkip((params.Index - 1) * params.Size)
	if data, err = x.Find(ctx, filter, sorts, fields, option); err != nil {
		return
	}
	return
}

func (x *Service) FindOne(ctx context.Context, filter M, fields []string) (data M, err error) {
	params := ctx.Value("params").(*Params)
	if err = x.Format(filter, params.FormatFilter); err != nil {
		return
	}
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

func (x *Service) UpdateMany(ctx context.Context, filter M, update M) (_ interface{}, err error) {
	params := x.Params(ctx)
	if err = x.Format(filter, params.FormatFilter); err != nil {
		return
	}
	if update["$set"] != nil {
		if err = x.Format(update["$set"].(M), params.FormatDoc); err != nil {
			return
		}
		update["$set"].(M)["update_time"] = time.Now()
	}
	return x.Db.Collection(params.Model).UpdateMany(ctx, filter, update)
}

func (x *Service) UpdateOneById(ctx context.Context, update M) (_ interface{}, err error) {
	params := x.Params(ctx)
	if update["$set"] != nil {
		if err = x.Format(update["$set"].(M), params.FormatDoc); err != nil {
			return
		}
		update["$set"].(M)["update_time"] = time.Now()
	}
	oid, _ := primitive.ObjectIDFromHex(params.Id)
	return x.Db.Collection(params.Model).UpdateOne(ctx, M{"_id": oid}, update)
}

func (x *Service) ReplaceOneById(ctx context.Context, doc M) (_ interface{}, err error) {
	params := x.Params(ctx)
	oid, _ := primitive.ObjectIDFromHex(params.Id)
	if err = x.Format(doc, params.FormatDoc); err != nil {
		return
	}
	doc["create_time"], doc["update_time"] = time.Now(), time.Now()
	return x.Db.Collection(params.Model).ReplaceOne(ctx, M{"_id": oid}, doc)
}

func (x *Service) DeleteMany(ctx context.Context, filter M) (_ interface{}, err error) {
	params := x.Params(ctx)
	// 格式化定义
	if err = x.Format(filter, params.FormatFilter); err != nil {
		return
	}
	return x.Db.Collection(params.Model).DeleteMany(ctx, filter)
}

func (x *Service) DeleteOneById(ctx context.Context) (interface{}, error) {
	params := x.Params(ctx)
	oid, _ := primitive.ObjectIDFromHex(params.Id)
	return x.Db.Collection(params.Model).DeleteOne(ctx, M{"_id": oid})
}

func (x *Service) CountDocuments(ctx context.Context, filter M) (count int64, err error) {
	params := x.Params(ctx)
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

// ExistDocument 文档
func (x *Service) ExistDocument(ctx context.Context, filter M) (result bool, err error) {
	params := x.Params(ctx)
	var count int64
	if count, err = x.Db.Collection(params.Model).
		CountDocuments(ctx, filter); err != nil {
		return
	}
	result = count != 0
	return
}

// Format 针对 filter 或 document 字段格式化
func (x *Service) Format(data M, rules []string) (err error) {
	for _, rule := range rules {
		spec := strings.Split(rule, ":")
		keys, cursor := strings.Split(spec[0], "."), data
		n := len(keys) - 1
		for _, key := range keys[:n] {
			if v, ok := cursor[key].(M); ok {
				cursor = v
			}
		}
		key := keys[n]
		if cursor[key] == nil {
			continue
		}
		switch spec[1] {
		case "oid":
			// 转换为 ObjectId
			if cursor[key], err = primitive.ObjectIDFromHex(cursor[key].(string)); err != nil {
				return
			}
			break

		case "oids":
			// 转换为 ObjectId 数组
			oids := cursor[key].([]interface{})
			for i, id := range oids {
				if oids[i], err = primitive.ObjectIDFromHex(id.(string)); err != nil {
					return
				}
			}
			break

		case "password":
			// 密码类型，转换为 Argon2id
			if cursor[key], err = password.Create(cursor[key].(string)); err != nil {
				return
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
func (x *Service) Event(ctx context.Context, action string, data interface{}) (err error) {
	params := x.Params(ctx)
	if option, ok := x.Engine.Options[params.Model]; ok {
		if !option.Event {
			return
		}
		var payload []byte
		if payload, err = jsoniter.Marshal(M{
			"action": action,
			"data":   data,
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
