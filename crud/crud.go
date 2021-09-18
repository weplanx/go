package crud

import (
	"github.com/gin-gonic/gin"
)

type Crud struct {
}

// API create resource operation
func (x *Crud) API(name string) *API {
	return &API{}
}

type API struct {
	Operations
}

type Operations interface {
	First(c *gin.Context) interface{}
	Find(c *gin.Context) interface{}
	Page(c *gin.Context) interface{}
	Create(c *gin.Context) interface{}
	Update(c *gin.Context) interface{}
	Delete(c *gin.Context) interface{}
}
