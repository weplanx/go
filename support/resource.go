package support

import (
	"database/sql/driver"
	jsoniter "github.com/json-iterator/go"
)

type Resource struct {
	ID     uint64 `json:"id"`
	Name   string `gorm:"type:varchar(20);not null" json:"name"`
	Key    string `gorm:"type:varchar(20);unique;not null" json:"key"`
	Parent string `gorm:"type:varchar(20);default:'root'" json:"parent"`
	Router *bool  `gorm:"default:false;comment:是否为路由页面" json:"router"`
	Nav    *bool  `gorm:"default:false;comment:是否为导航" json:"nav"`
	Icon   string `gorm:"type:varchar(20);comment:导航字体图标" json:"icon"`
	Schema Schema `gorm:"type:jsonb;default:'{}';comment:模型声明" json:"schema"`
	Sort   uint8  `gorm:"default:0;comment:导航排序" json:"sort"`
}

type Schema struct {
	Type       string      `json:"type"`
	Columns    []Column    `json:"columns"`
	Associates []Associate `json:"associates,omitempty"`
	System     *bool       `json:"system,omitempty"`
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
	Unique  *bool  `json:"unique,omitempty"`
	Require *bool  `json:"require,omitempty"`
	Hide    *bool  `json:"hide,omitempty"`
	System  *bool  `json:"system,omitempty"`
}

type Associate struct {
	Mode       string `json:"mode"`
	Target     string `json:"target"`
	References string `json:"references,omitempty"`
}
