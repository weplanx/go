package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/huandu/xstrings"
	"net/http"
	"reflect"
)

type Option struct {
	Path        string
	Middlewares []MiddlewareOption
}

type OptionFunc func(*Option)

func SetPath(path string) OptionFunc {
	return func(option *Option) {
		option.Path = path
	}
}

type MiddlewareOption struct {
	Handler fiber.Handler
	Effects []string
}

func SetMiddleware(middleware fiber.Handler, effect ...string) OptionFunc {
	return func(option *Option) {
		option.Middlewares = append(option.Middlewares, MiddlewareOption{
			Handler: middleware,
			Effects: effect,
		})
	}
}

// Auto 生成路由
func Auto(r fiber.Router, i interface{}, options ...OptionFunc) fiber.Router {
	typ := reflect.TypeOf(i)
	val := reflect.ValueOf(i)
	opt := new(Option)
	for _, option := range options {
		option(opt)
	}
	s := r.Group(opt.Path)
	{
		scopes := make(map[string][]fiber.Handler)
		for _, x := range opt.Middlewares {
			if len(x.Effects) == 0 {
				s.Use(x.Handler)
			} else {
				for _, v := range x.Effects {
					scopes[v] = append(scopes[v], x.Handler)
				}
			}
		}
		for x := 0; x < typ.NumMethod(); x++ {
			name := typ.Method(x).Name
			method := val.MethodByName(name).Interface()
			scopes[name] = append(scopes[name], Returns(method.(func(c *fiber.Ctx) interface{})))
			s.Post(xstrings.ToSnakeCase(name), scopes[name]...)
		}
	}
	return s
}

type E struct {
	Code    int64
	Message string
}

// Returns 返回统一结果
func Returns(fn func(c *fiber.Ctx) interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		switch x := fn(c).(type) {
		case E:
			return c.JSON(fiber.Map{
				"code":    x.Code,
				"message": x.Message,
			})
		case error:
			return c.JSON(fiber.Map{
				"code":    -1,
				"message": x.Error(),
			})
		default:
			if x != nil {
				return c.JSON(fiber.Map{
					"code": 0,
					"data": x,
				})
			} else {
				c.Status(http.StatusNotFound)
				return nil
			}
		}
	}
}
