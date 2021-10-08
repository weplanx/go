package basic

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Page struct {
	Parent   string       `bson:"parent" json:"parent"`
	Fragment string       `bson:"fragment" json:"fragment"`
	Name     string       `bson:"name" json:"name"`
	Router   RouterOption `bson:"router" json:"router"`
	Nav      *bool        `bson:"nav" json:"nav"`
	Icon     string       `bson:"icon" json:"icon"`
	Sort     uint8        `bson:"sort" json:"sort"`
}

type RouterOption struct {
	Collection string       `json:"collection,omitempty"`
	Template   string       `json:"template,omitempty"`
	Fetch      bool         `json:"fetch,omitempty"`
	Fields     []ViewFields `json:"columns,omitempty"`
}

type ViewFields struct {
	Field string `json:"field"`
}

func GeneratePage(ctx context.Context, db *mongo.Database) (err error) {
	collection := db.Collection("page")
	if _, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{"parent", 1},
			{"fragment", 1},
		},
		Options: options.Index().SetName("parent_fragment_idx").SetUnique(true),
	}); err != nil {
		return
	}
	if _, err = collection.InsertOne(ctx, Page{
		Parent:   "root",
		Fragment: "dashboard",
		Name:     "仪表盘",
		Nav:      True(),
		Router: RouterOption{
			Template: "manual",
		},
		Icon: "dashboard",
	}); err != nil {
		return
	}
	center, err := collection.InsertOne(ctx, Page{
		Parent:   "root",
		Fragment: "center",
		Name:     "个人中心",
	})
	if err != nil {
		return
	}
	if _, err = collection.InsertMany(ctx, []interface{}{
		Page{
			Parent:   center.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "profile",
			Name:     "我的信息",
			Router: RouterOption{
				Template: "manual",
			},
		},
		Page{
			Parent:   center.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "notification",
			Name:     "消息通知",
			Router: RouterOption{
				Template: "manual",
			},
		},
	}); err != nil {
		return
	}
	settings, err := collection.InsertOne(ctx, Page{
		Parent:   "root",
		Fragment: "settings",
		Name:     "设置",
		Nav:      True(),
		Icon:     "setting",
	})
	if err != nil {
		return
	}
	if _, err = collection.InsertMany(ctx, []interface{}{
		Page{
			Parent:   settings.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "schema",
			Name:     "模型管理",
			Nav:      True(),
			Router: RouterOption{
				Template: "manual",
			},
		},
		Page{
			Parent:   settings.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "page",
			Name:     "页面管理",
			Nav:      True(),
			Router: RouterOption{
				Template: "manual",
			},
		},
		Page{
			Parent:   settings.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "role",
			Name:     "权限管理",
			Nav:      True(),
			Router: RouterOption{
				Collection: "role",
				Template:   "list",
			},
		},
		Page{
			Parent:   settings.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "admin",
			Name:     "成员管理",
			Nav:      True(),
			Router: RouterOption{
				Collection: "admin",
				Template:   "list",
			},
		},
	}); err != nil {
		return
	}
	var role map[string]interface{}
	if err = collection.FindOne(ctx, bson.M{
		"parent":   settings.InsertedID.(primitive.ObjectID).Hex(),
		"fragment": "role",
	}).Decode(&role); err != nil {
		return
	}
	if _, err = collection.InsertMany(ctx, []interface{}{
		Page{
			Parent:   role["_id"].(primitive.ObjectID).Hex(),
			Fragment: "create",
			Name:     "创建资源",
			Router: RouterOption{
				Collection: "role",
				Template:   "form",
			},
		},
		Page{
			Parent:   role["_id"].(primitive.ObjectID).Hex(),
			Fragment: "update",
			Name:     "更新资源",
			Router: RouterOption{
				Collection: "role",
				Template:   "form",
				Fetch:      true,
			},
		},
	}); err != nil {
		return
	}
	var admin map[string]interface{}
	if err = collection.FindOne(ctx, bson.M{
		"parent":   settings.InsertedID.(primitive.ObjectID).Hex(),
		"fragment": "admin",
	}).Decode(&admin); err != nil {
		return
	}
	if _, err = collection.InsertMany(ctx, []interface{}{
		Page{
			Parent:   admin["_id"].(primitive.ObjectID).Hex(),
			Fragment: "create",
			Name:     "创建资源",
			Router: RouterOption{
				Collection: "admin",
				Template:   "form",
			},
		},
		Page{
			Parent:   admin["_id"].(primitive.ObjectID).Hex(),
			Fragment: "update",
			Name:     "更新资源",
			Router: RouterOption{
				Collection: "admin",
				Template:   "form",
				Fetch:      true,
			},
		},
	}); err != nil {
		return
	}
	return
}
