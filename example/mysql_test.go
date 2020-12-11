package example

import (
	"example/model"
	jsoniter "github.com/json-iterator/go"
	curd "github.com/kainonly/gin-curd"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"testing"
)

var db *gorm.DB
var c *curd.Curd

func TestMain(m *testing.M) {
	var err error
	dsn := os.Getenv("mysql_dsn")
	if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
		log.Fatalln(err)
	}
	c = curd.Initialize(db)
	os.Exit(m.Run())
}

func TestInitialize(t *testing.T) {
	var err error
	if err = db.Migrator().DropTable(&model.Example{}); err != nil {
		t.Error(err)
	}
	if err = db.AutoMigrate(&model.Example{}); err != nil {
		t.Error(err)
	}
	data := []model.Example{
		{KeyId: "main", Name: "Common Module", Status: false},
		{KeyId: "resource", Name: "Resource Module", Status: true},
		{KeyId: "acl", Name: "Acl Module", Status: true},
		{KeyId: "policy", Name: "Policy Module", Status: false},
		{KeyId: "admin", Name: "Admin Module", Status: true},
		{KeyId: "role", Name: "Role Module", Status: true},
	}
	if err = db.Create(&data).Error; err != nil {
		t.Error(err)
	}
}

type originListsBody struct {
	curd.OriginLists
}

func TestOriginLists(t *testing.T) {
	var err error
	var body originListsBody
	if err = jsoniter.Unmarshal(
		[]byte(`{"where":[["key_id","=","main"]]}`),
		&body,
	); err != nil {
		t.Error(err)
	}
	result := c.Operates(
		curd.Plan(&model.Example{}, body),
	).Originlists()
	t.Log(result)
}
