package model

import (
	"database/sql/driver"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

type Datastore struct {
	ID     uint   `json:"id"`
	Key    string `gorm:"type:varchar(50);not null;unique" json:"key"`
	Type   string `gorm:"type:varchar(20);default:collection;not null" json:"type"`
	Schema Schema `gorm:"type:jsonb;default:'{}'" json:"schema"`
}

type Schema []Column

type Column struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Type    string `json:"type"`
	Default string `json:"default,omitempty"`
	Require bool   `json:"require,omitempty"`
	Unique  bool   `json:"unique,omitempty"`
	Length  uint   `json:"length,omitempty"`
	Comment string `json:"comment,omitempty"`
	Hide    bool   `json:"hide,omitempty"`
}

func (x *Schema) Scan(input interface{}) error {
	data, ok := input.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSON value:", input))
	}
	return jsoniter.Unmarshal(data, x)
}

func (x Schema) Value() (driver.Value, error) {
	if len(x) == 0 {
		return nil, nil
	}
	data, err := jsoniter.Marshal(x)
	return string(data), err
}
