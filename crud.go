package bit

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Crud struct {
	db      *gorm.DB
	model   interface{}
	orderBy []string
}

func (x *Crud) Get(c *gin.Context) interface{} {
	return gin.H{}
}
