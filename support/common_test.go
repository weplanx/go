package support

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"testing"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	var err error
	if db, err = gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}); err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestGenerateResources(t *testing.T) {
	if err := GenerateResources(db); err != nil {
		t.Error(err)
	}
}

func TestGenerateModels(t *testing.T) {
	buf, err := GenerateModels(db)
	if err != nil {
		t.Error(err)
	}
	t.Log(buf.String())
}
