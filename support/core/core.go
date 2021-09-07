package core

import (
	"bytes"
	"database/sql/driver"
	jsoniter "github.com/json-iterator/go"
	"go/format"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
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

func True() *bool {
	value := true
	return &value
}

func False() *bool {
	return new(bool)
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
						Require: True(),
						Unique:  True(),
						System:  True(),
					},
					{
						Key:     "name",
						Label:   "权限名称",
						Type:    "varchar",
						Require: True(),
						System:  True(),
					},
					{
						Key:    "description",
						Label:  "描述",
						Type:   "text",
						System: True(),
					},
				},
				Associates: []Associate{},
				System:     True(),
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

func GenerateModels(tx *gorm.DB) (err error) {
	var resources []Resource
	if err = tx.
		Where("schema <> ?", "{}").
		Find(&resources).Error; err != nil {
		return
	}
	var tmpl *template.Template
	if tmpl, err = template.New("model").Funcs(template.FuncMap{
		"title":     title,
		"addColumn": addColumn,
	}).Parse(modelTpl); err != nil {
		return
	}
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, resources); err != nil {
		return
	}
	if _, err = os.Stat("./model"); os.IsNotExist(err) {
		os.Mkdir("./model", os.ModeDir)
	}
	b, _ := format.Source(buf.Bytes())
	if err = ioutil.WriteFile("./model/model_gen.go", b, os.ModePerm); err != nil {
		return
	}
	return
}

func title(s string) string {
	return strings.Title(s)
}

func dataType(val string) string {
	switch val {
	case "int":
		return "int32"
	case "int8":
		return "int64"
	case "decimal":
		return "float64"
	case "float8":
		return "float64"
	case "varchar":
		return "string"
	case "text":
		return "string"
	case "bool":
		return "*bool"
	case "timestamptz":
		return "time.Time"
	case "jsonb":
		return "Object"
	case "uuid":
		return "uuid.UUID"
	}
	return val
}

func addColumn(val Column) string {
	var b strings.Builder
	b.WriteString(title(val.Key))
	b.WriteString(" ")
	b.WriteString(dataType(val.Type))
	b.WriteString(" `")
	b.WriteString(`gorm:"type:`)
	b.WriteString(val.Type)
	if val.Require == True() {
		b.WriteString(`;not null`)
	}
	if val.Unique == True() {
		b.WriteString(`;unique`)
	}
	if val.Default != "" {
		b.WriteString(`;default:`)
		b.WriteString(val.Default)
	}
	b.WriteString(`"`)
	if val.Hide == True() {
		b.WriteString(` json:"-"`)
	}
	b.WriteString("`\n")
	return b.String()
}
