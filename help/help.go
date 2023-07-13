package help

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/weplanx/go/rest"
	"github.com/weplanx/go/sessions"
	"github.com/weplanx/go/values"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegValidate() {
	binding.MustRegValidateFunc("mongoId", func(args ...interface{}) error {
		if len(args) != 1 {
			return fmt.Errorf("the args must be one")
		}
		if _, e := primitive.ObjectIDFromHex(args[0].(string)); e != nil {
			return e
		}
		return nil
	})
}

func ValuesRoutes(u *route.RouterGroup, x *values.Controller) {
	r := u.Group("values")
	{
		r.GET("", x.Get)
		r.PATCH("", x.Set)
		r.DELETE(":key", x.Remove)
	}
}

func SessionsRoutes(u *route.RouterGroup, x *sessions.Controller) {
	r := u.Group("sessions")
	{
		r.GET("", x.Lists)
		r.DELETE(":uid", x.Remove)
		r.POST("clear", x.Clear)
	}
}

func RestRoutes(u *route.RouterGroup, x *rest.Controller) {
	r := u.Group(":collection")
	{
		r.GET(":id", x.FindById)
		r.POST("create", x.Create)
		r.POST("bulk_create", x.BulkCreate)
		r.POST("size", x.Size)
		r.POST("find", x.Find)
		r.POST("find_one", x.FindOne)
		r.POST("update", x.Update)
		r.POST("bulk_delete", x.BulkDelete)
		r.POST("sort", x.Sort)
		r.PATCH(":id", x.UpdateById)
		r.PUT(":id", x.Replace)
		r.DELETE(":id", x.Delete)
	}
	u.POST("transaction", x.Transaction)
	u.POST("commit", x.Commit)
}
