package go_bit

import (
	"github.com/kainonly/go-bit/crud"
	"gorm.io/gorm"
)

type Bit struct {
	Db *gorm.DB
}

// Initialize 初始化辅助器
func Initialize(db *gorm.DB) *Bit {
	return &Bit{Db: db}
}

// Crud 创建控制器通用资源操作
//	参数:
//	 model: 模型名称
//	 options: 配置
func (x *Bit) Crud(model interface{}, options ...crud.Option) *crud.Crud {
	c := &crud.Crud{
		Db:    x.Db,
		Model: model,
	}
	for _, apply := range options {
		apply(c)
	}
	return c
}
