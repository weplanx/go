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
	Type       string      `json:"type"`
	Columns    []Column    `json:"columns"`
	Associates []Associate `json:"associates,omitempty"`
	System     bool        `json:"system,omitempty"`
}

func (x *Schema) Scan(input interface{}) error {
	return jsoniter.Unmarshal(input.([]byte), x)
}

func (x Schema) Value() (driver.Value, error) {
	return jsoniter.Marshal(x)
}

type Column struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Type    string `json:"type"`
	Default string `json:"default,omitempty"`
	Unique  bool   `json:"unique,omitempty"`
	Require bool   `json:"require,omitempty"`
	Hide    bool   `json:"hide,omitempty"`
	System  bool   `json:"system,omitempty"`
}

type Associate struct {
	Mode       string `json:"mode"`
	Target     string `json:"target"`
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
				},
				Associates: []Associate{},
				System:     true,
			},
		},
		{
			Key:    "admin",
			Parent: "settings",
			Name:   "成员管理",
			Nav:    True(),
			Router: True(),
		},
		{
			Key:    "resource",
			Parent: "settings",
			Name:   "资源管理",
			Nav:    True(),
			Router: True(),
		},
	}
	if err = tx.Create(&data).Error; err != nil {
		return
	}
	return
}
