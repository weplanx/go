package pages

import (
	"context"
	"github.com/weplanx/server/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	Db *mongo.Database
}

type Nav struct {
	ID     primitive.ObjectID `bson:"_id" json:"_id"`
	Parent interface{}        `json:"parent"`
	Name   string             `json:"name"`
	Icon   string             `json:"icon"`
	Kind   string             `json:"kind"`
	Sort   int64              `json:"sort"`
}

// GetNavs 筛选导航数据
func (x *Service) GetNavs(ctx context.Context) (data []Nav, err error) {
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection("pages").
		Find(ctx, bson.M{"status": true}); err != nil {
		return
	}
	if err = cursor.All(ctx, &data); err != nil {
		return
	}
	return
}

// FindOneById 通过 ID 查找
func (x *Service) FindOneById(ctx context.Context, id primitive.ObjectID) (data model.Page, err error) {
	if err = x.Db.Collection("pages").
		FindOne(ctx, bson.M{"_id": id}).
		Decode(&data); err != nil {
		return
	}
	return
}
