package core

import (
	"bytes"
	"database/sql/driver"
	jsoniter "github.com/json-iterator/go"
	"go/format"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
	"strconv"
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
	Lock       *bool       `json:"lock,omitempty"`
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
	Require *bool  `json:"require,omitempty"`
	Unique  *bool  `json:"unique,omitempty"`
	Length  uint   `json:"length,omitempty"`
	Hide    *bool  `json:"hide,omitempty"`
	Lock    *bool  `json:"lock,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type Associate struct {
	Mode           string   `json:"mode"`
	Namespace      string   `json:"namespace,omitempty"`
	Target         string   `json:"target"`
	ForeignKey     string   `json:"foreign_key"`
	References     string   `json:"references"`
	JoinForeignKey string   `json:"join_foreign_key"`
	JoinReferences string   `json:"join_references"`
	Columns        []Column `json:"columns"`
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
						Length:  20,
						Lock:    True(),
					},
					{
						Key:     "name",
						Label:   "权限名称",
						Type:    "varchar",
						Require: True(),
						Length:  20,
						Lock:    True(),
					},
					{
						Key:   "description",
						Label: "描述",
						Type:  "text",
						Lock:  True(),
					},
				},
				Associates: []Associate{
					{
						Mode:           "many2many",
						Namespace:      "core",
						Target:         "Resource",
						ForeignKey:     "Key",
						JoinForeignKey: "Key",
						Columns: []Column{
							{
								Key:     "permission",
								Label:   "读写权限",
								Type:    "char",
								Default: "r",
								Require: True(),
								Length:  1,
							},
						},
					},
				},
				Lock: True(),
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
		"title": title,
		"typ":   typ,
		"tag":   tag,
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

func typ(datatype string) string {
	switch datatype {
	case "bigint":
		return "int64"
	case "integer":
		return "int32"
	case "smallint":
		return "int16"
	case "numeric":
		return "float64"
	case "double":
		return "float64"
	case "real":
		return "float32"
	case "boolean":
		return "*bool"
	case "char":
		return "string"
	case "varchar":
		return "string"
	case "text":
		return "string"
	case "timestamptz":
		return "time.Time"
	case "uuid":
		return "string"
	case "jsonb":
		return "Object"
	case "json":
		return "Object"
	}
	return datatype
}

func tag(column Column) string {
	var b strings.Builder
	b.WriteString(`gorm:"type:`)
	b.WriteString(column.Type)
	if column.Length != 0 {
		b.WriteString(`(`)
		b.WriteString(strconv.Itoa(int(column.Length)))
		b.WriteString(`)`)
	}
	if column.Require == True() {
		b.WriteString(`;not null`)
	}
	if column.Unique == True() {
		b.WriteString(`;unique`)
	}
	if column.Default != "" {
		b.WriteString(`;default:`)
		b.WriteString(column.Default)
	}
	b.WriteString(`"`)
	if column.Hide == True() {
		b.WriteString(` json:"-"`)
	}
	return b.String()
}
