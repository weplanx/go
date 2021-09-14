package support

import (
	"database/sql/driver"
	"github.com/alexedwards/argon2id"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

func True() *bool {
	value := true
	return &value
}

type Array []interface{}

func (x *Array) Scan(input interface{}) error {
	return jsoniter.Unmarshal(input.([]byte), x)
}

func (x Array) Value() (driver.Value, error) {
	return jsoniter.Marshal(x)
}

func InitSeeder(tx *gorm.DB) (err error) {
	var routers Array
	if err = tx.Model(&Resource{}).Pluck("id", &routers).Error; err != nil {
		return
	}
	roles := []map[string]interface{}{
		{
			"key":         "*",
			"name":        "超级管理员",
			"description": "超级管理员拥有完整权限不能编辑，若不使用可以禁用该权限",
			"routers":     Array{},
			"permissions": Array{},
		},
		{
			"key":         "admin",
			"name":        "管理员",
			"description": "分配管理用户",
			"routers":     routers,
			"permissions": Array{
				"resource:*",
				"role:*",
				"admin:*",
			},
		},
	}
	if err = tx.Table("role").Create(&roles).Error; err != nil {
		return
	}
	var password string
	if password, err = argon2id.CreateHash(
		"pass@VAN1234",
		argon2id.DefaultParams,
	); err != nil {
		return
	}
	admins := []map[string]interface{}{
		{
			"username": "admin",
			"password": password,
			"roles":    Array{"*"},
		},
	}
	if err = tx.Table("admin").Create(&admins).Error; err != nil {
		return
	}
	return
}
