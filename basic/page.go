package basic

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Page struct {
	Parent   string     `bson:"parent" json:"parent"`
	Fragment string     `bson:"fragment" json:"fragment"`
	Name     string     `bson:"name" json:"name"`
	Nav      *bool      `bson:"nav" json:"nav"`
	Icon     string     `bson:"icon" json:"icon"`
	Sort     uint8      `bson:"sort" json:"sort"`
	Router   string     `bson:"router" json:"router"`
	Option   PageOption `bson:"option,omitempty" json:"option,omitempty"`
}

type PageOption struct {
	Schema     string       `bson:"schema,omitempty" json:"schema,omitempty"`
	Fetch      *bool        `bson:"fetch,omitempty" json:"fetch,omitempty"`
	Fields     []ViewFields `bson:"fields,omitempty" json:"fields,omitempty"`
	Validation string       `bson:"validation,omitempty" json:"validation,omitempty"`
}

type ViewFields struct {
	Key     string `bson:"key" json:"key"`
	Label   string `bson:"label" json:"label"`
	Display *bool  `bson:"display" json:"display"`
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
		Icon:     "dashboard",
		Router:   "manual",
	}); err != nil {
		return
	}
	center, err := collection.InsertOne(ctx, Page{
		Parent:   "root",
		Fragment: "center",
		Name:     "个人中心",
		Nav:      False(),
	})
	if err != nil {
		return
	}
	if _, err = collection.InsertMany(ctx, []interface{}{
		Page{
			Parent:   center.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "profile",
			Name:     "我的信息",
			Nav:      False(),
			Router:   "manual",
		},
		Page{
			Parent:   center.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "notification",
			Name:     "消息通知",
			Nav:      False(),
			Router:   "manual",
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
	roleViewFields := []ViewFields{
		{
			Key:     "name",
			Label:   "权限代码",
			Display: True(),
		},
		{
			Key:     "key",
			Label:   "权限代码",
			Display: True(),
		},
		{
			Key:     "description",
			Label:   "描述",
			Display: True(),
		},
		{
			Key:     "pages",
			Label:   "页面",
			Display: True(),
		},
	}
	adminViewFields := []ViewFields{
		{
			Key:     "username",
			Label:   "用户名",
			Display: True(),
		},
		{
			Key:     "password",
			Label:   "密码",
			Display: True(),
		},
		{
			Key:     "roles",
			Label:   "权限",
			Display: True(),
		},
		{
			Key:     "name",
			Label:   "姓名",
			Display: True(),
		},
		{
			Key:     "email",
			Label:   "邮件",
			Display: True(),
		},
		{
			Key:     "phone",
			Label:   "联系方式",
			Display: True(),
		},
		{
			Key:     "avatar",
			Label:   "头像",
			Display: True(),
		},
	}
	if _, err = collection.InsertMany(ctx, []interface{}{
		Page{
			Parent:   settings.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "schema",
			Name:     "模式管理",
			Nav:      True(),
			Router:   "manual",
		},
		Page{
			Parent:   settings.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "page",
			Name:     "页面管理",
			Nav:      True(),
			Router:   "manual",
		},
		Page{
			Parent:   settings.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "role",
			Name:     "权限管理",
			Nav:      True(),
			Router:   "table",
			Option: PageOption{
				Schema: "role",
				Fields: roleViewFields,
			},
		},
		Page{
			Parent:   settings.InsertedID.(primitive.ObjectID).Hex(),
			Fragment: "admin",
			Name:     "成员管理",
			Nav:      True(),
			Router:   "table",
			Option: PageOption{
				Schema: "admin",
				Fields: adminViewFields,
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
			Nav:      False(),
			Router:   "form",
			Option: PageOption{
				Schema: "role",
				Fetch:  False(),
				Fields: roleViewFields,
			},
		},
		Page{
			Parent:   role["_id"].(primitive.ObjectID).Hex(),
			Fragment: "update",
			Name:     "更新资源",
			Nav:      False(),
			Router:   "form",
			Option: PageOption{
				Schema: "role",
				Fetch:  True(),
				Fields: roleViewFields,
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
			Nav:      False(),
			Router:   "form",
			Option: PageOption{
				Schema: "admin",
				Fields: adminViewFields,
			},
		},
		Page{
			Parent:   admin["_id"].(primitive.ObjectID).Hex(),
			Fragment: "update",
			Name:     "更新资源",
			Nav:      False(),
			Router:   "form",
			Option: PageOption{
				Schema: "admin",
				Fetch:  True(),
				Fields: adminViewFields,
			},
		},
	}); err != nil {
		return
	}
	return
}
