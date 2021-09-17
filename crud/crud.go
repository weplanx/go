package crud

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type Crud struct {
	Db *mongo.Database
}

func New(db *mongo.Database) *Crud {
	return &Crud{Db: db}
}

// API create resource operation
func (x *Crud) API(name string) *API {
	return &API{
		Collection: x.Db.Collection(name),
	}
}

type API struct {
	Collection *mongo.Collection
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
