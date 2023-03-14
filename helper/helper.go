package helper

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/weplanx/utils/kv"
	"github.com/weplanx/utils/resources"
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

func BindKV(r *route.RouterGroup, kv *kv.Controller) {
	r.GET("", kv.Get)
	r.PATCH("", kv.Set)
	r.DELETE(":key", kv.Remove)
}

func BindSessions(r *route.RouterGroup, sessions *sessions.Controller) {
	r.GET("", sessions.Lists)
	r.DELETE(":uid", sessions.Remove)
	r.DELETE("", sessions.Clear)
}

func BindDSL(r *route.RouterGroup, x *resources.Controller) {
	r.POST("", x.Create)
	r.POST("bulk-create", x.BulkCreate)
	r.GET("_size", x.Size)
	r.GET("", x.Find)
	r.GET("_one", x.FindOne)
	r.GET(":id", x.FindById)
	r.PATCH("", x.Update)
	r.PATCH(":id", x.UpdateById)
	r.PUT(":id", x.Replace)
	r.DELETE(":id", x.Delete)
	r.POST("bulk-delete", x.BulkDelete)
	r.POST("sort", x.Sort)
}
