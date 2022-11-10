package helper

import (
	"context"
	"fmt"
	tbinding "github.com/bytedance/go-tagexpr/v2/binding"
	"github.com/bytedance/go-tagexpr/v2/validator"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/bytedance/sonic/decoder"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/weplanx/utils/dsl"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
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

// ErrHandler 错误处理
func ErrHandler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Next(ctx)
		err := c.Errors.Last()
		if err == nil {
			return
		}

		if err.IsType(errors.ErrorTypePublic) {
			statusCode := http.StatusBadRequest
			result := utils.H{"message": err.Error()}
			if meta, ok := err.Meta.(map[string]interface{}); ok {
				if meta["statusCode"] != nil {
					statusCode = meta["statusCode"].(int)
				}
				if meta["code"] != nil {
					result["code"] = meta["code"]
				}
			}
			c.JSON(statusCode, result)
			return
		}

		switch e := err.Err.(type) {
		case decoder.SyntaxError:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Description(),
			})
			break
		case *tbinding.Error:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Error(),
			})
			break
		case *validator.Error:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Error(),
			})
			break
		default:
			logger.Error(err)
			c.Status(http.StatusInternalServerError)
		}
	}
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
