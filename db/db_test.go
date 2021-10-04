package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
	_ "github.com/lib/pq"
	"log"
	"os"
	"testing"
)

var db *sqlx.DB

func TestMain(m *testing.M) {
	var err error
	if db, err = sqlx.Connect("postgres", os.Getenv("DATABASE_URL")); err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func toJSON(rows *sqlx.Rows, value *map[string]interface{}) (err error) {
	types, err := rows.ColumnTypes()
	if err != nil {
		return
	}
	for _, typ := range types {
		switch typ.DatabaseTypeName() {
		case "ARRAY":
		case "JSON":
		case "JSONB":
			var JSON map[string]interface{}
			if err = jsoniter.Unmarshal((*value)[typ.Name()].([]byte), &JSON); err != nil {
				return
			}
			(*value)[typ.Name()] = JSON
			break
		}
	}
	return
}

func TestQuery(t *testing.T) {
	rows, err := db.QueryxContext(context.TODO(), `select * from schema`)
	if err != nil {
		t.Error(err)
	}
	data := make([]map[string]interface{}, 0)
	for rows.Next() {
		value := make(map[string]interface{})
		rows.MapScan(value)
		toJSON(rows, &value)
		data = append(data, value)
	}
	t.Log(data)
}
