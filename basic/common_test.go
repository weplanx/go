package basic

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"testing"
)

var (
	client *mongo.Client
	db     *mongo.Database
)

func TestMain(m *testing.M) {
	var err error
	if client, err = mongo.Connect(
		context.TODO(),
		options.Client().ApplyURI(os.Getenv("DB_URI")),
	); err != nil {
		return
	}
	db = client.Database(os.Getenv("DB_NAME"))
	os.Exit(m.Run())
}
