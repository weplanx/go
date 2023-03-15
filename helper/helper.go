package helper

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/weplanx/utils/resources"
	"github.com/weplanx/utils/sessions"
	"github.com/weplanx/utils/values"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

func BindValues(u *route.RouterGroup, x *values.Controller) {
	r := u.Group("values")
	{
		r.GET("", x.Get)
		r.PATCH("", x.Set)
		r.DELETE(":key", x.Remove)
	}
}

func BindSessions(u *route.RouterGroup, x *sessions.Controller) {
	r := u.Group("sessions")
	{
		r.GET("", x.Lists)
		r.DELETE(":uid", x.Remove)
		r.DELETE("", x.Clear)
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
	u.POST("transaction", x.Transaction)
	u.POST("commit", x.Commit)
}
