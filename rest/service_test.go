package rest_test

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/rest"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

func TestMorePipe(t *testing.T) {
	input := M{
		"data": M{
			"date": "abc",
			"dates": []interface{}{
				"a",
				"b",
			},
			"timestamps": []interface{}{
				"a",
				"b",
			},
			"metadata": nil,
			"items": []interface{}{
				M{"sn": "123"},
				M{"sn": "456"},
			},
		},
	}
	err := service.Pipe(input, []string{"data", "date"}, "date")
	assert.Error(t, err)
	err = service.Pipe(input, []string{"data", "dates"}, "dates")
	assert.Error(t, err)
	err = service.Pipe(input, []string{"data", "timestamps"}, "timestamps")
	assert.Error(t, err)
	err = service.Pipe(input, []string{"data", "metadata", "$", "v"}, "password")
	assert.NoError(t, err)
	err = service.Pipe(input, []string{"data", "items", "$", "sn"}, "date")
	assert.Error(t, err)
	err = service.Pipe(input, []string{"data", "unkown"}, "date")
	assert.NoError(t, err)
}

func TestDeleteDurable(t *testing.T) {
	ctx := context.TODO()
	err := service.Db.Collection("x_users").Drop(ctx)
	assert.NoError(t, err)
	r, err := service.Db.Collection("x_users").InsertOne(ctx, bson.M{
		"name":     "kain",
		"_durable": true,
	})
	assert.NoError(t, err)

	id := r.InsertedID.(primitive.ObjectID)
	_, err = service.Delete(ctx, "x_users", id, false)

	count, err := service.Db.Collection("x_users").CountDocuments(ctx, bson.M{"_id": id})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func MockDb(ctx context.Context) (err error) {
	usersOption := options.CreateCollection().
		SetValidator(bson.D{
			{"$jsonSchema", bson.D{
				{"title", "users"},
				{"required", bson.A{"_id", "name", "password", "department", "roles", "create_time", "update_time"}},
				{"properties", bson.D{
					{"_id", bson.M{"bsonType": "objectId"}},
					{"name", bson.M{"bsonType": "string"}},
					{"password", bson.M{"bsonType": "string"}},
					{"phone", bson.M{"bsonType": "string"}},
					{"department", bson.M{"bsonType": []string{"null", "objectId"}}},
					{"roles", bson.M{
						"bsonType": "array",
						"items":    bson.M{"bsonType": "objectId"},
					}},
					{"create_time", bson.M{"bsonType": "date"}},
					{"update_time", bson.M{"bsonType": "date"}},
				}},
				{"additionalProperties", false},
			}},
		})
	if err = db.CreateCollection(ctx, "users", usersOption); err != nil {
		return
	}
	ordersOption := options.CreateCollection().
		SetValidator(bson.D{
			{"$jsonSchema", bson.D{
				{"title", "orders"},
				{"required", bson.A{"_id", "no", "customer", "phone", "cost", "time", "create_time", "update_time"}},
				{"properties", bson.D{
					{"_id", bson.M{"bsonType": "objectId"}},
					{"no", bson.M{"bsonType": "string"}},
					{"customer", bson.M{"bsonType": "string"}},
					{"phone", bson.M{"bsonType": "string"}},
					{"cost", bson.M{"bsonType": "number"}},
					{"time", bson.M{"bsonType": "date"}},
					{"sort", bson.M{"bsonType": []string{"null", "number"}}},
					{"create_time", bson.M{"bsonType": "date"}},
					{"update_time", bson.M{"bsonType": "date"}},
				}},
				{"additionalProperties", false},
			}},
		})
	if err = db.CreateCollection(ctx, "orders", ordersOption); err != nil {
		return
	}
	projectsOption := options.CreateCollection().SetValidator(bson.D{
		{"$jsonSchema", bson.D{
			{"title", "projects"},
			{"required", bson.A{"_id", "name", "namespace", "secret", "create_time", "update_time"}},
			{"properties", bson.D{
				{"_id", bson.M{"bsonType": "objectId"}},
				{"name", bson.M{"bsonType": "string"}},
				{"namespace", bson.M{"bsonType": "string"}},
				{"secret", bson.M{"bsonType": "string"}},
				{"expire_time", bson.M{"bsonType": []string{"null", "date"}}},
				{"sort", bson.M{"bsonType": []string{"null", "number"}}},
				{"create_time", bson.M{"bsonType": "date"}},
				{"update_time", bson.M{"bsonType": "date"}},
			}},
			{"additionalProperties", false},
		}},
	})
	if err = db.CreateCollection(ctx, "projects", projectsOption); err != nil {
		return
	}
	return
}

func MockStream(ctx context.Context) (err error) {
	for k, v := range service.Values.RestControls {
		if v.Event {
			name := fmt.Sprintf(`EVENT_%s`, k)
			subject := fmt.Sprintf(`events.%s`, k)
			js.DeleteStream(name)
			if _, err := js.AddStream(&nats.StreamConfig{
				Name:      name,
				Subjects:  []string{subject},
				Retention: nats.WorkQueuePolicy,
			}, nats.Context(ctx)); err != nil {
				panic(err)
			}
		}
	}
	return
}

func MockSubscribe(t *testing.T, ch chan rest.PublishDto) {
	name := fmt.Sprintf(`EVENT_%s`, "projects")
	subject := fmt.Sprintf(`events.%s`, "projects")
	_, err := service.JetStream.QueueSubscribe(subject, name, func(msg *nats.Msg) {
		var data rest.PublishDto
		err := sonic.Unmarshal(msg.Data, &data)
		assert.NoError(t, err)

		ch <- data

		msg.Ack()
	}, nats.ManualAck())

	assert.NoError(t, err)
}

func RemoveStream(t *testing.T) {
	name := fmt.Sprintf(`EVENT_%s`, "projects")
	err := service.JetStream.DeleteStream(name)
	assert.NoError(t, err)
}

func RecoverStream(t *testing.T) {
	name := fmt.Sprintf(`EVENT_%s`, "projects")
	subject := fmt.Sprintf(`events.%s`, "projects")
	_, err := service.JetStream.AddStream(&nats.StreamConfig{
		Name:      name,
		Subjects:  []string{subject},
		Retention: nats.WorkQueuePolicy,
	})
	assert.NoError(t, err)
}
