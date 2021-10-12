package basic

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Schema struct {
	Key         string  `bson:"key" json:"key"`
	Label       string  `bson:"label" json:"label"`
	Kind        string  `bson:"kind" json:"kind"`
	Description string  `bson:"description,omitempty" json:"description"`
	System      *bool   `bson:"system,omitempty" json:"system"`
	Fields      []Field `bson:"fields,omitempty" json:"fields"`
}

type Field struct {
	Key         string      `bson:"key" json:"key"`
	Label       string      `bson:"label" json:"label"`
	Type        string      `bson:"type" json:"type"`
	Description string      `bson:"description,omitempty" json:"description"`
	Default     string      `bson:"default,omitempty" json:"default,omitempty"`
	Unique      *bool       `bson:"unique,omitempty" json:"unique,omitempty"`
	Required    *bool       `bson:"required,omitempty" json:"required,omitempty"`
	Private     *bool       `bson:"private,omitempty" json:"private,omitempty"`
	System      *bool       `bson:"system,omitempty" json:"system,omitempty"`
	Option      FieldOption `bson:"option,omitempty" json:"option,omitempty"`
}

type FieldOption struct {
	// 数字类型
	Max interface{} `bson:"max,omitempty" json:"max,omitempty"`
	Min interface{} `bson:"min,omitempty" json:"min,omitempty"`
	// 引用类型
	Mode   string `bson:"mode,omitempty" json:"mode,omitempty"`
	Target string `bson:"target,omitempty" json:"target,omitempty"`
	To     string `bson:"to,omitempty" json:"to,omitempty"`
}

func GenerateSchema(ctx context.Context, db *mongo.Database) (err error) {
	collection := db.Collection("schema")
	if _, err = collection.InsertMany(ctx, []interface{}{
		Schema{
			Key:    "page",
			Label:  "动态页面",
			Kind:   "manual",
			System: True(),
		},
		Schema{
			Key:   "role",
			Label: "权限组",
			Kind:  "collection",
			Fields: []Field{
				{
					Key:      "key",
					Label:    "权限代码",
					Type:     "text",
					Required: True(),
					Unique:   True(),
					System:   True(),
				},
				{
					Key:      "name",
					Label:    "权限名称",
					Type:     "text",
					Required: True(),
					System:   True(),
				},
				{
					Key:    "description",
					Label:  "描述",
					Type:   "text",
					System: True(),
				},
				{
					Key:     "pages",
					Label:   "页面",
					Type:    "reference",
					Default: "'[]'",
					System:  True(),
					Option: FieldOption{
						Mode:   "manual",
						Target: "page",
					},
				},
			},
			System: True(),
		},
		Schema{
			Label: "成员",
			Key:   "admin",
			Kind:  "collection",
			Fields: []Field{
				{
					Key:      "username",
					Label:    "用户名",
					Type:     "text",
					Required: True(),
					Unique:   True(),
					System:   True(),
				},
				{
					Key:      "password",
					Label:    "密码",
					Type:     "password",
					Required: True(),
					Private:  True(),
					System:   True(),
				},
				{
					Key:      "roles",
					Label:    "权限",
					Type:     "reference",
					Required: True(),
					Default:  "'[]'",
					System:   True(),
					Option: FieldOption{
						Mode:   "many",
						Target: "role",
						To:     "key",
					},
				},
				{
					Key:    "name",
					Label:  "姓名",
					Type:   "text",
					System: True(),
				},
				{
					Key:    "email",
					Label:  "邮件",
					Type:   "email",
					System: True(),
				},
				{
					Key:    "phone",
					Label:  "联系方式",
					Type:   "text",
					System: True(),
				},
				{
					Key:     "avatar",
					Label:   "头像",
					Type:    "media",
					Default: "'[]'",
					System:  True(),
				},
			},
			System: True(),
		},
	}); err != nil {
		return
	}
	if _, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{
			"key": 1,
		},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return
	}
	return
}
