package support

import (
	"database/sql/driver"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

type Resource struct {
	ID     int64  `json:"id"`
	Name   string `gorm:"type:varchar;not null" json:"name"`
	Key    string `gorm:"type:varchar;unique;not null" json:"key"`
	Parent string `gorm:"type:varchar;default:'root'" json:"parent"`
	Router *bool  `gorm:"default:false" json:"router"`
	Nav    *bool  `gorm:"default:false" json:"nav"`
	Icon   string `gorm:"type:varchar" json:"icon"`
	Schema Schema `gorm:"type:jsonb;default:'{}'" json:"schema"`
	Sort   int8   `gorm:"default:0" json:"sort"`
}

type Schema struct {
	Type    string   `json:"type"`
	Columns []Column `json:"columns"`
	System  bool     `json:"system,omitempty"`
}

func (x *Schema) Scan(input interface{}) error {
	return jsoniter.Unmarshal(input.([]byte), x)
}

func (x Schema) Value() (driver.Value, error) {
	return jsoniter.Marshal(x)
}

type Column struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Default  string   `json:"default,omitempty"`
	Unique   bool     `json:"unique,omitempty"`
	Require  bool     `json:"require,omitempty"`
	Hide     bool     `json:"hide,omitempty"`
	Relation Relation `json:"relation,omitempty"`
	System   bool     `json:"system,omitempty"`
}

type Relation struct {
	Mode       string `json:"mode,omitempty"`
	Target     string `json:"target,omitempty"`
	References string `json:"references,omitempty"`
}

func GenerateResources(tx *gorm.DB) (err error) {
	if tx.Migrator().HasTable(&Resource{}) {
		if err = tx.Migrator().DropTable(&Resource{}); err != nil {
			return
		}
	}
	if err = tx.AutoMigrate(&Resource{}); err != nil {
		return
	}
	data := []Resource{
		{
			Key:    "settings",
			Name:   "设置",
			Nav:    True(),
			Router: False(),
			Icon:   "setting",
		},
		{
			Key:    "role",
			Parent: "settings",
			Name:   "权限管理",
			Nav:    True(),
			Router: True(),
			Schema: Schema{
				Type: "collection",
				Columns: []Column{
					{
						Key:     "key",
						Label:   "权限代码",
						Type:    "varchar",
						Require: true,
						Unique:  true,
						System:  true,
					},
					{
						Key:     "name",
						Label:   "权限名称",
						Type:    "varchar",
						Require: true,
						System:  true,
					},
					{
						Key:    "description",
						Label:  "描述",
						Type:   "text",
						System: true,
					},
					{
						Key:   "permissions",
						Label: "策略",
						Type:  "rel",
						Relation: Relation{
							Mode:   "customize",
							Target: "resources",
						},
						System: true,
					},
				},
				System: true,
			},
		},
		{
			Key:    "admin",
			Parent: "settings",
			Name:   "成员管理",
			Nav:    True(),
			Router: True(),
			Schema: Schema{
				Type: "collection",
				Columns: []Column{
					{
						Key:     "uuid",
						Label:   "唯一标识",
						Type:    "uuid",
						Default: "uuid_generate_v4()",
						Require: true,
						Unique:  true,
						Hide:    true,
						System:  true,
					},
					{
						Key:     "username",
						Label:   "用户名",
						Type:    "varchar",
						Require: true,
						Unique:  true,
						System:  true,
					},
					{
						Key:     "password",
						Label:   "密码",
						Type:    "varchar",
						Require: true,
						System:  true,
					},
					{
						Key:     "roles",
						Label:   "权限",
						Type:    "rel",
						Require: true,
						Relation: Relation{
							Mode:       "many",
							Target:     "role",
							References: "key",
						},
						System: true,
					},
					{
						Key:   "permissions",
						Label: "附加策略",
						Type:  "rel",
						Relation: Relation{
							Mode:   "customize",
							Target: "resource",
						},
						System: true,
					},
					{
						Key:    "name",
						Label:  "姓名",
						Type:   "varchar",
						System: true,
					},
					{
						Key:    "email",
						Label:  "邮件",
						Type:   "varchar",
						System: true,
					},
					{
						Key:    "phone",
						Label:  "联系方式",
						Type:   "varchar",
						System: true,
					},
					{
						Key:     "avatar",
						Label:   "头像",
						Type:    "jsonb",
						Default: "'[]'",
						System:  true,
					},
				},
				System: true,
			},
		},
		{
			Key:    "resource",
			Parent: "settings",
			Name:   "资源管理",
			Nav:    True(),
			Router: True(),
			Schema: Schema{
				Type:    "customize",
				Columns: []Column{},
				System:  true,
			},
		},
	}
	if err = tx.Create(&data).Error; err != nil {
		return
	}
	return
}
