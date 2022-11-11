package helper

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/weplanx/utils/dsl"
	"github.com/weplanx/utils/sessions"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RegValidate 扩展验证
func RegValidate() {
	binding.MustRegValidateFunc("mongoId", func(args ...interface{}) error {
		if len(args) != 1 {
			return fmt.Errorf("the args must be one")
		}
		if !primitive.IsValidObjectID(args[0].(string)) {
			return primitive.ErrInvalidHex
		}
		return nil
	})
}

func BindSessions(r *route.RouterGroup, sessions *sessions.Controller) {
	r.GET("", sessions.Lists)
	r.DELETE(":uid", sessions.Remove)
	r.DELETE("", sessions.Clear)
}

func BindDSL(r *route.RouterGroup, dsl *dsl.Controller) {
	r.POST("", dsl.Create)
	r.POST("bulk-create", dsl.BulkCreate)
	r.GET("_size", dsl.Size)
	r.GET("", dsl.Find)
	r.GET("_one", dsl.FindOne)
	r.GET(":id", dsl.FindById)
	r.PATCH("", dsl.Update)
	r.PATCH(":id", dsl.UpdateById)
	r.PUT(":id", dsl.Replace)
	r.DELETE(":id", dsl.Delete)
	r.POST("bulk-delete", dsl.BulkDelete)
	r.POST("sort", dsl.Sort)
}
