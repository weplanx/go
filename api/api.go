package api

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type API struct {
	MongoClient *mongo.Client
	Db          *mongo.Database
}

func New(client *mongo.Client, db *mongo.Database) *API {
	x := new(API)
	x.MongoClient = client
	x.Db = db
	return x
}

func (x *API) getCollectionName(ctx context.Context) string {
	return ctx.Value("collection").(string)
}

func (x *API) Create(ctx context.Context, body interface{}) (*mongo.InsertOneResult, error) {
	name := x.getCollectionName(ctx)
	switch data := body.(type) {
	case bson.M:
		data["create_time"] = time.Now()
		data["update_time"] = time.Now()
		return x.Db.Collection(name).InsertOne(ctx, data)
	case bson.D:
		data = append(data,
			bson.D{
				{"create_time", time.Now()},
				{"update_time", time.Now()},
			}...,
		)
	}
	return x.Db.Collection(name).InsertOne(ctx, body)
}

func (x *API) FindOne(ctx context.Context, body *FindOneDto, data interface{}) error {
	name := x.getCollectionName(ctx)
	var filter bson.M
	if body.Id.IsZero() == false {
		filter = bson.M{"_id": body.Id}
	} else {
		filter = body.Where
	}
	return x.Db.Collection(name).FindOne(ctx, filter).Decode(data)
}

func (x *API) Find(ctx context.Context, body *FindDto, data interface{}) (err error) {
	name := x.getCollectionName(ctx)
	var filter bson.M
	if len(body.Id) != 0 {
		filter = bson.M{"_id": bson.M{"$in": body.Id}}
	} else {
		filter = body.Where
	}
	opts := options.Find()
	if len(body.Sort) != 0 {
		var sorts bson.D
		for _, v := range body.Sort {
			sorts = append(sorts, bson.E{Key: v[0].(string), Value: v[1]})
		}
		opts.SetSort(sorts)
		opts.SetAllowDiskUse(true)
	} else {
		opts.SetSort(bson.M{"_id": -1})
	}
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection(name).
		Find(ctx, filter, opts); err != nil {
		return
	}
	if err = cursor.All(ctx, data); err != nil {
		return
	}
	return
}

func (x *API) FindByPage(ctx context.Context, body *FindByPageDto) (result FindByPageResult, err error) {
	name := x.getCollectionName(ctx)
	if len(body.Where) != 0 {
		if result.Total, err = x.Db.Collection(name).CountDocuments(ctx, body.Where); err != nil {
			return
		}
	} else {
		if result.Total, err = x.Db.Collection(name).EstimatedDocumentCount(ctx); err != nil {
			return
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
	} else {
		opts.SetSort(bson.M{"_id": -1})
	}
	opts.SetLimit(page.Size)
	opts.SetSkip((page.Index - 1) * page.Size)
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection(name).Find(ctx, body.Where, opts); err != nil {
		return
	}
	if err = cursor.All(ctx, &result.Value); err != nil {
		return
	}
	return
}

func (x *API) Update(ctx context.Context, body *UpdateDto) (result *mongo.UpdateResult, err error) {
	name := x.getCollectionName(ctx)
	filter := body.Where
	if body.Id.IsZero() == false {
		filter = bson.M{"_id": body.Id}
	}
	body.Update["$set"].(map[string]interface{})["update_time"] = time.Now()
	for _, key := range body.Refs {
		for i, id := range body.Update["$set"].(map[string]interface{})[key].([]interface{}) {
			if body.Update["$set"].(map[string]interface{})[key].([]interface{})[i], err = primitive.
				ObjectIDFromHex(id.(string)); err != nil {
				return
			}
		}
	}
	return x.Db.Collection(name).UpdateOne(ctx, filter, body.Update)
}

func (x *API) Delete(ctx context.Context, body *DeleteDto) (*mongo.DeleteResult, error) {
	name := x.getCollectionName(ctx)
	var filter bson.M
	if len(body.Id) != 0 {
		filter = bson.M{"_id": bson.M{"$in": body.Id}}
	} else {
		filter = body.Where
	}
	return x.Db.Collection(name).DeleteMany(ctx, filter)
}
