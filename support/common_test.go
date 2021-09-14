package support

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"testing"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	var err error
	if db, err = gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}); err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestInitSeeder(t *testing.T) {
	if err := InitSeeder(db); err != nil {
		t.Error(err)
	}
}
