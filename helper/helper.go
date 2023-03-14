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

func BindKV(u *route.RouterGroup, kv *kv.Controller) {
	r := u.Group("values")
	{
		r.GET("", kv.Get)
		r.PATCH("", kv.Set)
		r.DELETE(":key", kv.Remove)
	}
}

func BindSessions(u *route.RouterGroup, sessions *sessions.Controller) {
	r := u.Group("session")
	{
		r.GET("", sessions.Lists)
		r.DELETE(":uid", sessions.Remove)
		r.DELETE("", sessions.Clear)
	}
}

func BindResources(u *route.RouterGroup, x *resources.Controller) {
	r := u.Group(":collection")
	{
		r.POST("", x.Create)
		r.POST("bulk_create", x.BulkCreate)
		r.GET("_size", x.Size)
		r.GET("", x.Find)
		r.GET("_one", x.FindOne)
		r.GET(":id", x.FindById)
		r.PATCH("", x.Update)
		r.PATCH(":id", x.UpdateById)
		r.PUT(":id", x.Replace)
		r.DELETE(":id", x.Delete)
		r.POST("bulk_delete", x.BulkDelete)
		r.POST("sort", x.Sort)
	}
	r.POST("transaction", x.Transaction)
	r.POST("commit", x.Commit)
}
