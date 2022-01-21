package engine

import (
	"context"
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
		option.SetSort(bson.M{"create_time": -1})
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
	if err = x.setFormat(format, update["$set"]); err != nil {
		return
	}
	if err = x.setRef(ref, update["$set"]); err != nil {
		return
	}
	update["$set"].(map[string]interface{})["update_time"] = time.Now()
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
	if err = x.setFormat(format, update["$set"]); err != nil {
		return
	}
	if err = x.setRef(ref, update["$set"]); err != nil {
		return
	}
	update["$set"].(map[string]interface{})["update_time"] = time.Now()
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
