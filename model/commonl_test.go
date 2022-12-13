package model_test

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"testing"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	var err error

	if db, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  os.Getenv("DATABASE_GORM"),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	}); err != nil {
		log.Fatalln(err)
	}

	os.Exit(m.Run())
}

