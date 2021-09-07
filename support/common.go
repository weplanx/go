package support

import (
	"bytes"
	"gorm.io/gorm"
	"strings"
	"text/template"
)

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

func GenerateModels(tx *gorm.DB) (buf bytes.Buffer, err error) {
	var resources []Resource
	if err = tx.
		Where("schema <> ?", "{}").
		Find(&resources).Error; err != nil {
		return
	}
	var tmpl *template.Template
	if tmpl, err = template.New("model").Funcs(template.FuncMap{
		"title":  title,
		"column": column,
	}).Parse(modelTpl); err != nil {
		return
	}
	if err = tmpl.Execute(&buf, resources); err != nil {
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

func column(val Column) string {
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
