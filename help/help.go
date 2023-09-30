package help

import (
	"context"
	errorsx "errors"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/go-playground/validator/v10"
	"github.com/hertz-contrib/binding/go_playground"
	"github.com/hertz-contrib/requestid"
	"os"
	"reflect"
	"regexp"
)

func Ptr[T any](i T) *T {
	return &i
}

func IsEmpty(i any) bool {
	if i == nil || i == "" || i == false {
		return true
	}

	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.Len() == 0 || v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr, reflect.Func, reflect.Chan:
		return v.IsNil()
	}

	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func HertzOptions(input ...config.Option) *config.Options {
	vd := go_playground.NewValidator()
	vd.SetValidateTag("vd")
	vdx := vd.Engine().(*validator.Validate)
	vdx.RegisterValidation("snake", func(fl validator.FieldLevel) bool {
		matched, err := regexp.MatchString("^[a-z_]+$", fl.Field().Interface().(string))
		if err != nil {
			return false
		}
		return matched
	})
	vdx.RegisterValidation("sort", func(fl validator.FieldLevel) bool {
		matched, err := regexp.MatchString("^[a-z_]+:(-1|1)$", fl.Field().Interface().(string))
		if err != nil {
			return false
		}
		return matched
	})
	opts := []config.Option{
		server.WithCustomValidator(vd),
	}
	for _, x := range input {
		opts = append(opts, x)
	}
	if os.Getenv("MODE") != "release" {
		opts = append(opts, server.WithExitWaitTime(0))
	}
	return config.NewOptions(opts)
}

type EMeta struct {
	Code string
}

func E(code string, msg string) *errors.Error {
	return errors.NewPublic(msg).SetMeta(&EMeta{Code: code})
}

func EHandler() app.HandlerFunc {
	release := os.Getenv("MODE") == "release"
	return func(ctx context.Context, c *app.RequestContext) {
		c.Next(ctx)
		e := c.Errors.Last()
		if e == nil {
			return
		}

		if e.IsType(errors.ErrorTypePublic) {
			r := utils.H{
				"code": "system.*",
				"msg":  e.Error(),
			}
			if meta, ok := e.Meta.(*EMeta); ok {
				r["code"] = meta.Code
			}
			c.JSON(400, r)
			return
		}

		var ves validator.ValidationErrors
		if errorsx.As(e.Err, &ves) {
			message := make([]interface{}, len(ves))
			for i, v := range ves {
				message[i] = utils.H{
					"namespace": v.Namespace(),
					"field":     v.Field(),
					"tag":       v.Tag(),
				}
			}
			c.JSON(400, utils.H{
				"code":    "system.validation",
				"message": message,
			})
			return
		}

		if !release {
			c.JSON(500, e.JSON())
			return
		}

		logger.Error(requestid.Get(c), e)
		c.Status(500)
	}
}
