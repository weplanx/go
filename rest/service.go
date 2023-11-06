package rest

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/weplanx/go/cipher"
	"github.com/weplanx/go/passlib"
	"github.com/weplanx/go/values"
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
	Mgo       *mongo.Client
	Db        *mongo.Database
	RDb       *redis.Client
	JetStream nats.JetStreamContext
	KeyValue  nats.KeyValue
	Values    *values.DynamicValues
	Cipher    *cipher.Cipher
}

func (x *Service) IsForbid(name string) bool {
	if x.Values.RestControls[name] == nil {
		return true
	}
	return !x.Values.RestControls[name].Status
}

const (
	ActionCreate     = 1
	ActionBulkCreate = 2
	ActionUpdate     = 3
	ActionUpdateById = 4
	ActionReplace    = 5
	ActionDelete     = 6
	ActionBulkDelete = 7
	ActionSort       = 8
)

func (x *Service) Create(ctx context.Context, name string, doc interface{}) (result interface{}, err error) {
	if result, err = x.Db.Collection(name).InsertOne(ctx, doc); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Action: ActionCreate,
		Data:   doc,
		Result: result,
	}); err != nil {
		return
	}
	return
}

func (x *Service) BulkCreate(ctx context.Context, name string, docs []interface{}) (result interface{}, err error) {
	if result, err = x.Db.Collection(name).InsertMany(ctx, docs); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Action: ActionBulkCreate,
		Data:   docs,
		Result: result,
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
	for cursor.Next(ctx) {
		var v M
		if err = cursor.Decode(&v); err != nil {
			return
		}
		x.Sensitive(name, v)
		data = append(data, v)
	}
	return
}

func (x *Service) FindOne(ctx context.Context, name string, filter M, option *options.FindOneOptions) (data M, err error) {
	if err = x.Db.Collection(name).FindOne(ctx, filter, option).Decode(&data); err != nil {
		return
	}
	x.Sensitive(name, data)
	return
}

func (x *Service) Update(ctx context.Context, name string, filter M, update interface{}, option *options.UpdateOptions) (result interface{}, err error) {
	if result, err = x.Db.Collection(name).UpdateMany(ctx, filter, update, option); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Action: ActionUpdate,
		Filter: filter,
		Data:   update,
		Result: result,
	}); err != nil {
		return
	}
	return
}

func (x *Service) UpdateById(ctx context.Context, name string, id primitive.ObjectID, update interface{}, option *options.UpdateOptions) (result interface{}, err error) {
	filter := M{"_id": id}
	if result, err = x.Db.Collection(name).UpdateOne(ctx, filter, update, option); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Action: ActionUpdateById,
		Id:     id.Hex(),
		Data:   update,
		Result: result,
	}); err != nil {
		return
	}
	return
}

func (x *Service) Replace(ctx context.Context, name string, id primitive.ObjectID, doc interface{}) (result interface{}, err error) {
	filter := M{"_id": id}
	if result, err = x.Db.Collection(name).ReplaceOne(ctx, filter, doc); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Action: ActionReplace,
		Id:     id.Hex(),
		Data:   doc,
		Result: result,
	}); err != nil {
		return
	}
	return
}

func (x *Service) Delete(ctx context.Context, name string, id primitive.ObjectID, transaction bool) (result interface{}, err error) {
	filter := M{
		"_id":      id,
		"_durable": bson.M{"$exists": false},
	}
	var doc M
	if !transaction {
		if err = x.Db.Collection(name).
			FindOne(ctx, filter).
			Decode(&doc); err != nil {
			return
		}
	}
	if result, err = x.Db.Collection(name).
		DeleteOne(ctx, filter); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Action: ActionDelete,
		Id:     id.Hex(),
		Data:   doc,
		Result: result,
	}); err != nil {
		return
	}
	return
}

func (x *Service) BulkDelete(ctx context.Context, name string, filter M, transaction bool) (result interface{}, err error) {
	filter["_durable"] = bson.M{"$exists": false}
	var docs []M
	if !transaction {
		var cursor *mongo.Cursor
		if cursor, err = x.Db.Collection(name).Find(ctx, filter); err != nil {
			return
		}
		if err = cursor.All(ctx, &docs); err != nil {
			return
		}
	}
	if result, err = x.Db.Collection(name).
		DeleteMany(ctx, filter); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Action: ActionBulkDelete,
		Filter: filter,
		Data:   docs,
		Result: result,
	}); err != nil {
		return
	}
	return
}

func (x *Service) Sort(ctx context.Context, name string, key string, ids []primitive.ObjectID) (result interface{}, err error) {
	var wms []mongo.WriteModel
	for i, id := range ids {
		update := M{"$set": M{key: i, "update_time": time.Now()}}
		wms = append(wms, mongo.NewUpdateOneModel().
			SetFilter(M{"_id": id}).
			SetUpdate(update),
		)
	}
	if result, err = x.Db.Collection(name).BulkWrite(ctx, wms); err != nil {
		return
	}
	if err = x.Publish(ctx, name, PublishDto{
		Action: ActionSort,
		Data: M{
			"key":    key,
			"values": ids,
		},
		Result: result,
	}); err != nil {
		return
	}
	return
}

func (x *Service) Transaction(ctx context.Context, txn string) {
	key := fmt.Sprintf(`transaction:%s`, txn)
	x.RDb.LPush(ctx, key, time.Now().Format(time.RFC3339)).Val()
	x.RDb.Expire(ctx, key, time.Hour*5).Val()
}

type PendingDto struct {
	Action int                `bson:"action"`
	Name   string             `bson:"name"`
	Id     primitive.ObjectID `bson:"id,omitempty"`
	Filter M                  `bson:"filter,omitempty"`
	Data   interface{}        `bson:"data,omitempty"`
}

func (x *Service) TxnNotExists(ctx context.Context, key string) (err error) {
	var exists int64
	if exists, err = x.RDb.Exists(ctx, key).Result(); err != nil {
		return
	}
	if exists != 1 {
		return ErrTxnNotExist
	}
	return
}

func (x *Service) Pending(ctx context.Context, txn string, dto PendingDto) (err error) {
	key := fmt.Sprintf(`transaction:%s`, txn)
	if err = x.TxnNotExists(ctx, key); err != nil {
		return
	}
	var b []byte
	if b, err = bson.Marshal(dto); err != nil {
		return
	}
	if err = x.RDb.LPush(ctx, key, b).Err(); err != nil {
		return
	}
	return
}

func (x *Service) Commit(ctx context.Context, txn string) (_ interface{}, err error) {
	key := fmt.Sprintf(`transaction:%s`, txn)
	if err = x.TxnNotExists(ctx, key); err != nil {
		return
	}
	var begin time.Time
	if begin, err = x.RDb.RPop(ctx, key).Time(); err != nil {
		return
	}
	if time.Since(begin) > x.Values.RestTxnTimeout {
		err = ErrTxnTimeOut
		return
	}

	var n int64
	if n, err = x.RDb.LLen(ctx, key).Result(); err != nil {
		return
	}

	opts := options.Session().SetDefaultReadConcern(readconcern.Majority())
	var session mongo.Session
	if session, err = x.Mgo.StartSession(opts); err != nil {
		return
	}
	defer session.EndSession(ctx)

	txnOpts := options.Transaction().SetReadPreference(readpref.PrimaryPreferred())
	return session.WithTransaction(ctx, func(txnCtx mongo.SessionContext) (_ interface{}, err error) {
		var results []interface{}
		for n > 0 {
			var b []byte
			if b, err = x.RDb.RPop(ctx, key).Bytes(); err != nil {
				return
			}
			var dto PendingDto
			if err = bson.Unmarshal(b, &dto); err != nil {
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
	case ActionCreate:
		return x.Create(ctx, dto.Name, dto.Data)
	case ActionBulkCreate:
		return x.BulkCreate(ctx, dto.Name, dto.Data.(primitive.A))
	case ActionUpdate:
		return x.Update(ctx, dto.Name, dto.Filter, dto.Data, nil)
	case ActionUpdateById:
		return x.UpdateById(ctx, dto.Name, dto.Id, dto.Data, nil)
	case ActionReplace:
		return x.Replace(ctx, dto.Name, dto.Id, dto.Data)
	case ActionDelete:
		return x.Delete(ctx, dto.Name, dto.Id, true)
	case ActionBulkDelete:
		return x.BulkDelete(ctx, dto.Name, dto.Filter, true)
	case ActionSort:
		data := dto.Data.(primitive.D)
		var key string
		var ids []primitive.ObjectID
		for _, v := range data {
			switch v.Key {
			case "key":
				key = v.Value.(string)
				break
			case "values":
				for _, id := range v.Value.(primitive.A) {
					ids = append(ids, id.(primitive.ObjectID))
				}
				break
			}
		}
		return x.Sort(ctx, dto.Name, key, ids)
	}
	return
}

func (x *Service) Transform(data M, rules M) (err error) {
	for key, value := range rules {
		paths := strings.Split(key, "->")
		if err = x.Pipe(data, paths, value); err != nil {
			return
		}
	}
	return
}

func (x *Service) Pipe(input M, paths []string, kind interface{}) (err error) {
	var cursor interface{} = input
	n := len(paths) - 1
	for i, path := range paths[:n] {
		if path == "$" {
			for _, item := range cursor.([]interface{}) {
				if err = x.Pipe(item.(M), paths[i+1:], kind); err != nil {
					return
				}
			}
			return
		}
		if cursor.(M)[path] == nil {
			return
		}
		cursor = cursor.(M)[path]
	}
	key := paths[n]
	if cursor == nil || cursor.(M)[key] == nil {
		return
	}
	unknow := cursor.(M)[key]
	var data interface{}
	switch kind {
	case "oid":
		if data, err = primitive.ObjectIDFromHex(unknow.(string)); err != nil {
			return
		}
		break
	case "oids":
		oids := unknow.([]interface{})
		for i, id := range oids {
			if oids[i], err = primitive.ObjectIDFromHex(id.(string)); err != nil {
				return
			}
		}
		data = oids
		break
	case "date":
		if data, err = time.Parse(time.RFC1123, unknow.(string)); err != nil {
			return
		}
		break
	case "dates":
		dates := unknow.([]interface{})
		for i, date := range dates {
			if dates[i], err = time.Parse(time.RFC1123, date.(string)); err != nil {
				return
			}
		}
		data = dates
		break
	case "timestamp":
		if data, err = time.Parse(time.RFC3339, unknow.(string)); err != nil {
			return
		}
		break
	case "timestamps":
		timestamps := unknow.([]interface{})
		for i, timestamp := range timestamps {
			if timestamps[i], err = time.Parse(time.RFC3339, timestamp.(string)); err != nil {
				return
			}
		}
		data = timestamps
		break
	case "password":
		data, _ = passlib.Hash(unknow.(string))
		break
	case "cipher":
		if data, err = x.Cipher.Encode([]byte(unknow.(string))); err != nil {
			return
		}
		break
	}
	cursor.(M)[key] = data
	return
}

func (x *Service) Projection(name string, keys []string) (result bson.M) {
	result = make(bson.M)
	if x.Values.RestControls != nil && x.Values.RestControls[name] != nil {
		for _, key := range x.Values.RestControls[name].Keys {
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

func (x *Service) Sensitive(name string, v M) {
	if x.Values.RestControls != nil && x.Values.RestControls[name] != nil {
		for _, key := range x.Values.RestControls[name].Sensitives {
			if v[key] == nil {
				v[key] = "-"
			} else {
				v[key] = "*"
			}
		}
	}
}

type PublishDto struct {
	Action int         `json:"action"`
	Id     string      `json:"id,omitempty"`
	Filter M           `json:"filter,omitempty"`
	Data   interface{} `json:"data,omitempty"`
	Result interface{} `json:"result"`
}

func (x *Service) Publish(ctx context.Context, name string, dto PublishDto) (err error) {
	if v, ok := x.Values.RestControls[name]; ok {
		if !v.Event {
			return
		}

		b, _ := sonic.Marshal(dto)
		subject := fmt.Sprintf(`events.%s`, name)
		if _, err = x.JetStream.Publish(subject, b, nats.Context(ctx)); err != nil {
			return
		}
	}
	return
}
