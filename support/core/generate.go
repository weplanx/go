package core

import (
	"bytes"
	"go/format"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"
)

func GenerateModel(tx *gorm.DB) (err error) {
	var datastores []Datastore
	if err = tx.Find(&datastores).Error; err != nil {
		return
	}
	var tmpl *template.Template

	log.Println(os.Getwd())
	if tmpl, err = template.New("model.tpl").Funcs(template.FuncMap{
		"title": title,
		"typ":   typ,
		"tag":   tag,
	}).ParseFiles("./template/model.tpl"); err != nil {
		return
	}
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, datastores); err != nil {
		return
	}
	if _, err = os.Stat("./model"); os.IsNotExist(err) {
		os.Mkdir("./model", os.ModeDir)
	}
	b, _ := format.Source(buf.Bytes())
	if err = ioutil.WriteFile("./model/model_gen.go", b, os.ModeAppend); err != nil {
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
	if column.Require {
		b.WriteString(`;not null`)
	}
	if column.Unique {
		b.WriteString(`;unique`)
	}
	if column.Default != "" {
		b.WriteString(`;default:`)
		b.WriteString(column.Default)
	}
	b.WriteString(`"`)
	if column.Hide {
		b.WriteString(` json:"-"`)
	}
	return b.String()
}
