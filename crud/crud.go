package crud

import (
	"gorm.io/gorm"
)

type Crud struct {
	Db *gorm.DB
}

func New(db *gorm.DB) *Crud {
	return &Crud{db}
}

// Make 创建控制器通用资源操作
//	参数:
//	 model: 模型名称
//	 options: 配置
func (x *Crud) Make(model interface{}, options ...Option) *Resource {
	c := &Resource{
		Db:    x.Db,
		Model: model,
	}
	for _, apply := range options {
		apply(c)
	}
	return c
}

// Conditions 条件数组
type Conditions [][3]interface{}

func (c Conditions) GetConditions() Conditions {
	return c
}

// Orders 排序对象
type Orders map[string]string

func (c Orders) GetOrders() Orders {
	return c
}
