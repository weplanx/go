package help

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	errx "errors"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app"
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
	if i == nil || i == "" {
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
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

func Sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

func HmacSha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

func Validator() *go_playground.Validator {
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
	return vd
}

type R struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func Ok() R {
	return R{
		Code:    0,
		Message: "ok",
	}
}

func Fail(code int64, msg string) R {
	return R{
		Code:    code,
		Message: msg,
	}
}

type ErrorMeta struct {
	Code int64
}

func E(code int64, msg string) *errors.Error {
	return errors.NewPublic(msg).SetMeta(&ErrorMeta{Code: code})
}

func ErrorHandler() app.HandlerFunc {
	release := os.Getenv("MODE") == "release"
	return func(ctx context.Context, c *app.RequestContext) {
		c.Next(ctx)
		e := c.Errors.Last()
		if e == nil {
			return
		}

		if e.IsType(errors.ErrorTypePublic) {
			r := R{Code: 0, Message: e.Error()}
			if meta, ok := e.Meta.(*ErrorMeta); ok {
				r.Code = meta.Code
			}
			c.JSON(400, r)
			return
		}

		var ves validator.ValidationErrors
		if errx.As(e.Err, &ves) {
			message := make([]interface{}, len(ves))
			for i, v := range ves {
				message[i] = utils.H{
					"namespace": v.Namespace(),
					"field":     v.Field(),
					"tag":       v.Tag(),
				}
			}
			c.JSON(400, utils.H{
				"code":    0,
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
