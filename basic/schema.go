package basic

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Schema struct {
	Name       string  `bson:"name" json:"name"`
	Collection string  `bson:"collection" json:"collection"`
	Kind       string  `bson:"kind" json:"kind"`
	System     *bool   `bson:"system,omitempty" json:"system"`
	Fields     []Field `bson:"fields,omitempty" json:"fields"`
}

type Field struct {
	Name      string    `bson:"name,omitempty" json:"name"`
	Type      string    `bson:"type,omitempty" json:"type"`
	Label     string    `bson:"label,omitempty" json:"label"`
	Default   string    `bson:"default,omitempty" json:"default,omitempty"`
	Unique    bool      `bson:"unique,omitempty" json:"unique,omitempty"`
	Require   bool      `bson:"require,omitempty" json:"require,omitempty"`
	Reference Reference `bson:"reference,omitempty" json:"reference,omitempty"`
	Private   bool      `bson:"private,omitempty" json:"private,omitempty"`
	System    bool      `bson:"system,omitempty" json:"system,omitempty"`
}

type Reference struct {
	Mode   string `bson:"mode,omitempty" json:"mode,omitempty"`
	Target string `bson:"target,omitempty" json:"target,omitempty"`
	To     string `bson:"to,omitempty" json:"to,omitempty"`
}

func GenerateSchema(ctx context.Context, db *mongo.Database) (err error) {
	collection := db.Collection("schema")
	if _, err = collection.InsertMany(ctx, []interface{}{
		Schema{
			Name:       "动态页面",
			Collection: "page",
			Kind:       "manual",
			System:     True(),
		},
		Schema{
			Name:       "权限组",
			Collection: "role",
			Kind:       "collection",
			Fields: []Field{
				{
					Name:    "key",
					Type:    "String",
					Label:   "权限代码",
					Require: true,
					Unique:  true,
					System:  true,
				},
				{
					Name:    "name",
					Type:    "String",
					Label:   "权限名称",
					Require: true,
					System:  true,
				},
				{
					Name:   "description",
					Type:   "String",
					Label:  "描述",
					System: true,
				},
				{
					Name:    "pages",
					Type:    "Array",
					Label:   "页面",
					Default: "'[]'",
					Reference: Reference{
						Mode:   "manual",
						Target: "page",
					},
					System: true,
				},
			},
			System: True(),
		},
		Schema{
			Name:       "成员",
			Collection: "admin",
			Kind:       "collection",
			Fields: []Field{
				{
					Name:    "username",
					Type:    "String",
					Label:   "用户名",
					Require: true,
					Unique:  true,
					System:  true,
				},
				{
					Name:    "password",
					Type:    "String",
					Label:   "密码",
					Require: true,
					Private: true,
					System:  true,
				},
				{
					Name:    "roles",
					Type:    "Array",
					Label:   "权限",
					Require: true,
					Default: "'[]'",
					Reference: Reference{
						Mode:   "many",
						Target: "role",
						To:     "key",
					},
					System: true,
				},
				{
					Name:   "name",
					Type:   "String",
					Label:  "姓名",
					System: true,
				},
				{
					Name:   "email",
					Type:   "String",
					Label:  "邮件",
					System: true,
				},
				{
					Name:   "phone",
					Type:   "String",
					Label:  "联系方式",
					System: true,
				},
				{
					Name:    "avatar",
					Type:    "Array",
					Label:   "头像",
					Default: "'[]'",
					System:  true,
				},
			},
			System: True(),
		},
	}); err != nil {
		return
	}
	if _, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{
			"collection": 1,
		},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return
	}
	return
}
