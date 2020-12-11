package model

type Example struct {
	ID         uint64
	KeyId      string `gorm:"size:200;unique;not null"`
	Name       string `gorm:"size:50;not null"`
	Status     bool   `gorm:"not null;default:true"`
	CreateTime uint64 `gorm:"not null;autoCreateTime"`
	UpdateTime uint64 `gorm:"not null;autoUpdateTime"`
}
