package model

import (
	"database/sql/driver"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

type Example struct {
	ID         uint64
	KeyId      string `gorm:"size:200;unique;not null"`
	Name       Object `gorm:"type:json;size:50;not null"`
	Status     bool   `gorm:"not null;default:true"`
	CreateTime uint64 `gorm:"not null;autoCreateTime"`
	UpdateTime uint64 `gorm:"not null;autoUpdateTime"`
}

type Object map[string]interface{}

func (c *Object) Scan(input interface{}) error {
	bs, ok := input.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSON value:", input))
	}
	return jsoniter.Unmarshal(bs, c)
}

func (c Object) Value() (driver.Value, error) {
	if len(c) == 0 {
		return nil, nil
	}
	bs, err := jsoniter.Marshal(c)
	return string(bs), err
}
