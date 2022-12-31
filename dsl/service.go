package dsl

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/weplanx/utils/passlib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

type Service struct {
	*DSL
}

// Load Values
func (x *Service) Load(ctx context.Context) (err error) {
	for k, v := range x.DynamicValues.DSL {
		if v.Event {
			name := fmt.Sprintf(`%s:events:%s`, x.Namespace, k)
			subject := fmt.Sprintf(`%s.events.%s`, x.Namespace, k)
			if _, err = x.Js.AddStream(&nats.StreamConfig{
				Name:      name,
				Subjects:  []string{subject},
				Retention: nats.WorkQueuePolicy,
			}, nats.Context(ctx)); err != nil {
				return
			}
		}
	}
	return
}

func (x *Service) Create(ctx context.Context, name string, doc M) (_ interface{}, err error) {
	return x.Db.Collection(name).InsertOne(ctx, doc)
}

func (x *Service) BulkCreate(ctx context.Context, name string, docs []interface{}) (_ interface{}, err error) {
	return x.Db.Collection(name).InsertMany(ctx, docs)
}

func (x *Service) Size(ctx context.Context, name string, filter M) (_ int64, err error) {
	if len(filter) == 0 {
		return x.Db.Collection(name).EstimatedDocumentCount(ctx)
	}
	return x.Db.Collection(name).CountDocuments(ctx, filter)
}

func (x *Service) Find(ctx context.Context, name string, filter M, option *options.FindOptions) (data []M, err error) {
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection(name).Find(ctx, filter, option); err != nil {
		return
	}
	data = make([]M, 0)
	if err = cursor.All(ctx, &data); err != nil {
		return
	}
	return
}

func (x *Service) FindOne(ctx context.Context, name string, filter M, option *options.FindOneOptions) (data M, err error) {
	if err = x.Db.Collection(name).FindOne(ctx, filter, option).Decode(&data); err != nil {
		return
	}
	return
}

func (x *Service) Update(ctx context.Context, name string, filter M, update M) (_ interface{}, err error) {
	return x.Db.Collection(name).UpdateMany(ctx, filter, update)
}

func (x *Service) UpdateById(ctx context.Context, name string, id primitive.ObjectID, update M) (_ interface{}, err error) {
	return x.Db.Collection(name).UpdateOne(ctx, M{"_id": id}, update)
}

func (x *Service) Replace(ctx context.Context, name string, id primitive.ObjectID, doc M) (_ interface{}, err error) {
	return x.Db.Collection(name).ReplaceOne(ctx, M{"_id": id}, doc)
}

func (x *Service) Delete(ctx context.Context, name string, id primitive.ObjectID) (_ interface{}, err error) {
	return x.Db.Collection(name).DeleteOne(ctx, M{"_id": id, "labels.fixed": bson.M{"$exists": false}})
}

func (x *Service) BulkDelete(ctx context.Context, name string, filter M) (_ interface{}, err error) {
	filter["labels.fixed"] = bson.M{"$exists": false}
	return x.Db.Collection(name).DeleteMany(ctx, filter)
}

func (x *Service) Sort(ctx context.Context, name string, ids []primitive.ObjectID) (_ interface{}, err error) {
	var wms []mongo.WriteModel
	for i, id := range ids {
		update := M{
			"$set": M{
				"sort":        i,
				"update_time": time.Now(),
			},
		}

		wms = append(wms, mongo.NewUpdateOneModel().
			SetFilter(M{"_id": id}).
			SetUpdate(update),
		)
	}
	return x.Db.Collection(name).BulkWrite(ctx, wms)
}

func (x *Service) Transform(data M, format M) (err error) {
	for path, spec := range format {
		keys, cursor := strings.Split(path, "."), data
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
		switch spec {
		case "oid":
			if cursor[key], err = primitive.ObjectIDFromHex(cursor[key].(string)); err != nil {
				return
			}
			break

		case "oids":
			oids := cursor[key].([]interface{})
			for i, id := range oids {
				if oids[i], err = primitive.ObjectIDFromHex(id.(string)); err != nil {
					return
				}
			}
			break

		case "date":
			if cursor[key], err = time.Parse(time.RFC1123, cursor[key].(string)); err != nil {
				return
			}
			break

		case "timestamp":
			if cursor[key], err = time.Parse(time.RFC3339, cursor[key].(string)); err != nil {
				return
			}
			break

		case "password":
			if cursor[key], _ = passlib.Hash(cursor[key].(string)); err != nil {
				return
			}
			break
		}
	}
	return
}

func (x *Service) Projection(name string, keys []string) (result bson.M) {
	result = make(bson.M)
	if x.DynamicValues.DSL != nil && x.DynamicValues.DSL[name] != nil {
		for _, key := range x.DynamicValues.DSL[name].Keys {
			result[key] = 1
		}
	}
	if len(keys) != 0 {
		projection := make(bson.M)
		for _, key := range keys {
			if _, ok := result[key]; len(result) != 0 && !ok {
				continue
			}
			projection[key] = 1
		}
		result = projection
	}
	return
}

type PublishDto struct {
	Event        string      `json:"event"`
	Id           string      `json:"id,omitempty"`
	Filter       M           `json:"filter,omitempty"`
	FilterFormat M           `json:"filter_format,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	DataFormat   M           `json:"data_format,omitempty"`
	Result       interface{} `json:"result"`
}

func (x *Service) Publish(ctx context.Context, name string, dto PublishDto) (err error) {
	if v, ok := x.DynamicValues.DSL[name]; ok {
		if !v.Event {
			return
		}

		b, _ := sonic.Marshal(dto)
		subject := fmt.Sprintf(`%s.events.%s`, x.Namespace, name)
		if _, err = x.Js.Publish(subject, b, nats.Context(ctx)); err != nil {
			return
		}
	}
	return
}
