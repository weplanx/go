package basic

import (
	"context"
	"github.com/alexedwards/argon2id"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GenerateRoleAndAdmin(ctx context.Context, db *mongo.Database) (err error) {
	if _, err = db.Collection("role").InsertOne(ctx, bson.M{
		"key":         "*",
		"name":        "超级管理员",
		"status":      true,
		"description": "",
		"pages":       bson.A{},
	}); err != nil {
		return
	}
	var password string
	if password, err = argon2id.CreateHash("pass@VAN1234", argon2id.DefaultParams); err != nil {
		return
	}
	if _, err = db.Collection("admin").InsertOne(ctx, bson.M{
		"username": "admin",
		"password": password,
		"status":   true,
		"roles":    bson.A{"*"},
		"name":     "超级管理员",
		"email":    "",
		"phone":    "",
		"avatar":   "",
	}); err != nil {
		return
	}
	return
}
