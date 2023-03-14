package resources

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/weplanx/utils/passlib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strings"
	"time"
)

type Service struct {
	*DSL
}

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

func (x *Service) Create(ctx context.Context, name string, doc M) (r interface{}, err error) {
	if r, err = x.Db.Collection(name).InsertOne(ctx, doc); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Event:  "create",
		Data:   doc,
		Result: r,
	}); err != nil {
		return
	}
	return
}

func (x *Service) BulkCreate(ctx context.Context, name string, docs []interface{}) (r interface{}, err error) {
	if r, err = x.Db.Collection(name).InsertMany(ctx, docs); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Event:  "bulk-create",
		Data:   docs,
		Result: r,
	}); err != nil {
		return
	}
	return
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

func (x *Service) Update(ctx context.Context, name string, filter M, update M) (r interface{}, err error) {
	if r, err = x.Db.Collection(name).UpdateMany(ctx, filter, update); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Event:  "update",
		Filter: filter,
		Data:   update,
		Result: r,
	}); err != nil {
		return
	}
	return
}

func (x *Service) UpdateById(ctx context.Context, name string, id primitive.ObjectID, update M) (r interface{}, err error) {
	filter := M{"_id": id}
	if r, err = x.Db.Collection(name).UpdateOne(ctx, filter, update); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Event:  "update",
		Id:     id.Hex(),
		Data:   update,
		Result: r,
	}); err != nil {
		return
	}
	return
}

func (x *Service) Replace(ctx context.Context, name string, id primitive.ObjectID, doc M) (r interface{}, err error) {
	filter := M{"_id": id}
	if r, err = x.Db.Collection(name).ReplaceOne(ctx, filter, doc); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Event:  "replace",
		Id:     id.Hex(),
		Data:   doc,
		Result: r,
	}); err != nil {
		return
	}
	return
}

func (x *Service) Delete(ctx context.Context, name string, id primitive.ObjectID) (r interface{}, err error) {
	filter := M{
		"_id":                  id,
		"metadata.undeletable": bson.M{"$exists": false},
	}
	if r, err = x.Db.Collection(name).DeleteOne(ctx, filter); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Event:  "delete",
		Id:     id.Hex(),
		Result: r,
	}); err != nil {
		return
	}
	return
}

func (x *Service) BulkDelete(ctx context.Context, name string, filter M) (r interface{}, err error) {
	filter["metadata.undeletable"] = bson.M{"$exists": false}
	if r, err = x.Db.Collection(name).DeleteMany(ctx, filter); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Event:  "bulk-delete",
		Data:   filter,
		Result: r,
	}); err != nil {
		return
	}
	return
}

func (x *Service) Sort(ctx context.Context, name string, ids []primitive.ObjectID) (r interface{}, err error) {
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
	if r, err = x.Db.Collection(name).BulkWrite(ctx, wms); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Event:  "sort",
		Data:   ids,
		Result: r,
	}); err != nil {
		return
	}
	return
}

func (x *Service) Transaction(ctx context.Context, txn string) (err error) {
	key := fmt.Sprintf(`%s:transaction:%s`, x.Namespace, txn)
	if err = x.Redis.LPush(ctx, key, time.Now().Format(time.RFC3339)).Err(); err != nil {
		return
	}
	if err = x.Redis.Expire(ctx, key, time.Hour*5).Err(); err != nil {
		return
	}
	return
}

type PendingDto struct {
	Action string
	Name   string
	Data   map[string]interface{}
}

func (x *Service) Pending(ctx context.Context, txn string, dto PendingDto) (err error) {
	key := fmt.Sprintf(`%s:transaction:%s`, x.Namespace, txn)
	var b []byte
	if b, err = msgpack.Marshal(dto); err != nil {
		return
	}
	if err = x.Redis.LPush(ctx, key, b).Err(); err != nil {
		return
	}
	return
}

var ErrTxnTimeOut = errors.NewPublic("the transaction has timed out")

func (x *Service) Commit(ctx context.Context, txn string) (_ interface{}, err error) {
	key := fmt.Sprintf(`%s:transaction:%s`, x.Namespace, txn)
	var begin time.Time
	if begin, err = x.Redis.RPop(ctx, key).Time(); err != nil {
		return
	}
	if time.Since(begin) > time.Second*30 {
		err = ErrTxnTimeOut
		return
	}

	var n int64
	if n, err = x.Redis.LLen(ctx, key).Result(); err != nil {
		return
	}

	opts := options.Session().SetDefaultReadConcern(readconcern.Majority())
	var session mongo.Session
	if session, err = x.MongoClient.StartSession(opts); err != nil {
		return
	}
	defer session.EndSession(ctx)

	txnOpts := options.Transaction().SetReadPreference(readpref.PrimaryPreferred())
	return session.WithTransaction(ctx, func(txnCtx mongo.SessionContext) (_ interface{}, err error) {
		var results []interface{}
		for n > 0 {
			var b []byte
			if b, err = x.Redis.RPop(ctx, key).Bytes(); err != nil {
				return
			}
			var dto PendingDto
			if err = msgpack.Unmarshal(b, &dto); err != nil {
				return
			}
			var r interface{}
			if r, err = x.Invoke(txnCtx, dto); err != nil {
				return
			}
			results = append(results, r)
			n--
		}
		return results, nil
	}, txnOpts)
}

func (x *Service) Invoke(ctx context.Context, dto PendingDto) (_ interface{}, _ error) {
	switch dto.Action {
	case "Create":
		return x.Create(ctx, dto.Name, dto.Data)
	}
	return
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
	Event  string      `json:"event"`
	Id     string      `json:"id,omitempty"`
	Filter M           `json:"filter,omitempty"`
	Data   interface{} `json:"data,omitempty"`
	Result interface{} `json:"result"`
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
