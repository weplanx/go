package basic

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type Schema struct {
	Name    string  `bson:"name" json:"name"`
	Key     string  `bson:"key" json:"key"`
	Kind    string  `bson:"kind" json:"kind"`
	System  *bool   `bson:"system" json:"system"`
	Columns Columns `bson:"columns" json:"columns"`
}

type Columns map[string]Column

type Column struct {
	Label     string    `json:"label"`
	Type      string    `json:"type"`
	Default   string    `json:"default,omitempty"`
	Unique    bool      `json:"unique,omitempty"`
	Require   bool      `json:"require,omitempty"`
	Reference Reference `json:"reference,omitempty"`
	Private   bool      `json:"private,omitempty"`
	System    bool      `json:"system,omitempty"`
}

type Reference struct {
	Mode   string `json:"mode,omitempty"`
	Target string `json:"target,omitempty"`
	To     string `json:"to,omitempty"`
}

func GenerateSchema(db *mongo.Database) (err error) {
	if _, err = db.Collection("schema").InsertMany(context.TODO(), []interface{}{
		Schema{
			Name:   "页面集合",
			Key:    "page",
			Kind:   "manual",
			System: True(),
		},
		Schema{
			Name: "权限集合",
			Key:  "role",
			Kind: "collection",
			Columns: map[string]Column{
				"key": {
					Label:   "权限代码",
					Type:    "varchar",
					Require: true,
					Unique:  true,
					System:  true,
				},
				"name": {
					Label:   "权限名称",
					Type:    "varchar",
					Require: true,
					System:  true,
				},
				"description": {
					Label:  "描述",
					Type:   "text",
					System: true,
				},
				"routers": {
					Label:   "路由",
					Type:    "ref",
					Default: "'[]'",
					Reference: Reference{
						Mode:   "manual",
						Target: "resource",
					},
					System: true,
				},
				"permissions": {
					Label:   "策略",
					Type:    "ref",
					Default: "'[]'",
					Reference: Reference{
						Mode:   "manual",
						Target: "resource",
					},
					System: true,
				},
			},
			System: True(),
		},
		Schema{
			Name: "成员集合",
			Key:  "admin",
			Kind: "collection",
			Columns: map[string]Column{
				"uuid": {
					Label:   "唯一标识",
					Type:    "uuid",
					Default: "uuid_generate_v4()",
					Require: true,
					Unique:  true,
					Private: true,
					System:  true,
				},
				"username": {
					Label:   "用户名",
					Type:    "varchar",
					Require: true,
					Unique:  true,
					System:  true,
				},
				"password": {
					Label:   "密码",
					Type:    "varchar",
					Require: true,
					Private: true,
					System:  true,
				},
				"roles": {
					Label:   "权限",
					Type:    "ref",
					Require: true,
					Default: "'[]'",
					Reference: Reference{
						Mode:   "many",
						Target: "role",
						To:     "key",
					},
					System: true,
				},
				"name": {
					Label:  "姓名",
					Type:   "varchar",
					System: true,
				},
				"email": {
					Label:  "邮件",
					Type:   "varchar",
					System: true,
				},
				"phone": {
					Label:  "联系方式",
					Type:   "varchar",
					System: true,
				},
				"avatar": {
					Label:   "头像",
					Type:    "array",
					Default: "'[]'",
					System:  true,
				},
				"routers": {
					Label:   "路由",
					Type:    "ref",
					Default: "'[]'",
					Reference: Reference{
						Mode:   "manual",
						Target: "resource",
					},
					System: true,
				},
				"permissions": {
					Label:   "策略",
					Type:    "ref",
					Default: "'[]'",
					Reference: Reference{
						Mode:   "manual",
						Target: "resource",
					},
					System: true,
				},
			},
			System: True(),
		},
	}); err != nil {
		return
	}
	return
}
