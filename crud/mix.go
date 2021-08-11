package crud

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const mixStart = "mix.start"
const mixComplete = "mix.complete"

type mix struct {
	Body   interface{}
	data   interface{}
	query  func(tx *gorm.DB) *gorm.DB
	txNext func(tx *gorm.DB, args ...interface{}) error
}

type Operator func(*mix)

// SetBody 自定义请求体
//	场景描述:
//	 OriginLists,Lists,Get,Delete: 需要组合结构体 OriginListsBody, ListsBody, GetBody, DeleteBody
func SetBody(body interface{}) Operator {
	return func(c *mix) {
		if c.Body == nil {
			c.Body = body
		}
	}
}

// SetData 自定义数据
//	场景描述:
//	 OriginLists,Lists,Get: 指定 gorm 模型, 用于查询与最终的数据返回
//	 Add: 自定义创建数据
//	 Edit: 自定义更新数据
func SetData(data interface{}) Operator {
	return func(c *mix) {
		if c.data == nil {
			c.data = data
		}
	}
}

// Query 自定义查询
//	参数:
//	 fn: func(tx *gorm.DB) *gorm.DB
func Query(fn func(tx *gorm.DB) *gorm.DB) Operator {
	return func(c *mix) {
		c.query = fn
	}
}

// TxNext 设置事务包含的数据操作
//	参数:
//	 fn: func(tx *gorm.DB, args ...interface{}) error
func TxNext(fn func(tx *gorm.DB, args ...interface{}) error) Operator {
	return func(c *mix) {
		c.txNext = fn
	}
}

// Mix 定义混合操作
//	参数:
//	 c: *gin.Context
//	 operator: 操作
func Mix(c *gin.Context, operator ...Operator) {
	v := new(mix)
	for _, operator := range operator {
		operator(v)
	}
	c.Set(mixStart, v)
}
