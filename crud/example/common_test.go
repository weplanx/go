package example

import (
	"github.com/gin-gonic/gin"
	"github.com/kainonly/go-bit/crud"
	"github.com/kainonly/go-bit/mvc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"testing"
)

type Example struct {
	ID         uint64 `json:"id"`
	Email      string `gorm:"type:varchar(20);not null;unique" json:"path"`
	Name       string `gorm:"type:varchar(20);not null" json:"name"`
	Age        int    `gorm:"not null" json:"age"`
	Gender     string `gorm:"type:varchar(10);not null" json:"gender"`
	Department string `gorm:"type:varchar(20);not null" json:"department"`
}

var db *gorm.DB
var err error
var r *gin.Engine

func TestMain(m *testing.M) {
	if db, err = gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}); err != nil {
		log.Fatalln(err)
	}
	if err = db.Migrator().DropTable(&Example{}); err != nil {
		log.Fatalln(err)
	}
	if err = db.AutoMigrate(&Example{}); err != nil {
		log.Fatalln(err)
	}
	data := []Example{
		{Email: "Vandal@VX.com", Name: "Vandal", Age: 25, Gender: "Male", Department: "IT"},
		{Email: "Questa@VX.com", Name: "Questa", Age: 21, Gender: "Female", Department: "IT"},
		{Email: "Simone@VX.com", Name: "Simone", Age: 23, Gender: "Male", Department: "IT"},
		{Email: "Stuart@VX.com", Name: "Stuart", Age: 27, Gender: "Female", Department: "Sale"},
		{Email: "Vivianne@VX.com", Name: "Vivianne", Age: 36, Gender: "Male", Department: "Sale"},
		{Email: "Max@VX.com", Name: "Max", Age: 28, Gender: "Female", Department: "Designer"},
		{Email: "Eagle-Eyed@VX.com", Name: "Eagle-Eyed", Age: 31, Gender: "Male", Department: "Support"},
		{Email: "Marcia@VX.com", Name: "Marcia", Age: 37, Gender: "Female", Department: "Support"},
		{Email: "Joanna@VX.com", Name: "Joanna", Age: 40, Gender: "Male", Department: "Manager"},
		{Email: "Judy@VX.com", Name: "Judy", Age: 50, Gender: "Female", Department: "Manager"},
		{Email: "Robert@VX.com", Name: "Robert", Age: 22, Gender: "Male", Department: "IT"},
		{Email: "Kayla@VX.com", Name: "Kayla", Age: 55, Gender: "Female", Department: "Leader"},
		{Email: "Odette@VX.com", Name: "Odette", Age: 33, Gender: "Male", Department: "Sale"},
		{Email: "Nancy@VX.com", Name: "Nancy", Age: 31, Gender: "Female", Department: "Sale"},
		{Email: "Roxanne@VX.com", Name: "Roxanne", Age: 32, Gender: "Male", Department: "Sale"},
		{Email: "Ancestress@VX.com", Name: "Ancestress", Age: 27, Gender: "Female", Department: "Designer"},
		{Email: "Holly@VX.com", Name: "Holly", Age: 26, Gender: "Male", Department: "Designer"},
		{Email: "Gifford@VX.com", Name: "Gifford", Age: 38, Gender: "Female", Department: "Sale"},
		{Email: "Edgar@VX.com", Name: "Edgar", Age: 41, Gender: "Male", Department: "Sale"},
		{Email: "Forrest@VX.com", Name: "Forrest", Age: 45, Gender: "Female", Department: "Sale"},
	}
	if err = db.Create(&data).Error; err != nil {
		log.Fatalln(err)
	}
	gin.SetMode(gin.TestMode)
	c1 := new(UserController)
	c1.Crud = crud.New(db, &Example{})
	r = gin.Default()
	s1 := r.Group("user")
	{
		s1.POST("r/find/one", mvc.Bind(c1.FindOne))
		s1.POST("r/find/many", mvc.Bind(c1.FindMany))
		s1.POST("r/find/page", mvc.Bind(c1.FindPage))
		s1.POST("w/create", mvc.Bind(c1.Create))
		s1.POST("w/update", mvc.Bind(c1.Update))
		s1.POST("w/delete", mvc.Bind(c1.Delete))
	}
	c2 := new(UserMixController)
	c2.Crud = crud.New(db, &Example{})
	s2 := r.Group("user-mix")
	{
		s2.POST("r/find/one", mvc.Bind(c2.FindOne))
		s2.POST("r/find/many", mvc.Bind(c2.FindMany))
		s2.POST("r/find/page", mvc.Bind(c2.FindPage))
		s2.POST("w/create", mvc.Bind(c2.Create))
		s2.POST("w/update", mvc.Bind(c2.Update))
		s2.POST("w/delete", mvc.Bind(c2.Delete))

	}
	os.Exit(m.Run())
}
