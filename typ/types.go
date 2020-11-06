package typ

import "gorm.io/gorm"

type Common struct {
	Db *gorm.DB
}

type Conditions [][]interface{}
type Query func(tx *gorm.DB) *gorm.DB
type JSON map[string]interface{}
