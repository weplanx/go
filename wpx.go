package wpx

import (
	"github.com/gin-gonic/gin"
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
	Handler gin.HandlerFunc
	Effects []string
}

func SetMiddleware(middleware gin.HandlerFunc, effect ...string) OptionFunc {
	return func(option *Option) {
		option.Middlewares = append(option.Middlewares, MiddlewareOption{
			Handler: middleware,
			Effects: effect,
		})
	}
}

// Auto 生成路由
func Auto(r *gin.RouterGroup, i interface{}, options ...OptionFunc) *gin.RouterGroup {
	typ := reflect.TypeOf(i)
	val := reflect.ValueOf(i)
	opt := new(Option)
	for _, option := range options {
		option(opt)
	}
	s := r.Group(opt.Path)
	{
		scopes := make(map[string][]gin.HandlerFunc)
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
			scopes[name] = append(scopes[name], Returns(method.(func(c *gin.Context) interface{})))
			s.POST(xstrings.ToSnakeCase(name), scopes[name]...)
		}
	}
	return s
}

type E struct {
	Code    int64
	Message string
}

// Returns 返回统一结果
func Returns(fn func(c *gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch x := fn(c).(type) {
		case E:
			c.JSON(http.StatusOK, gin.H{
				"code":    x.Code,
				"message": x.Message,
			})
			break
		case error:
			c.JSON(http.StatusOK, gin.H{
				"code":    -1,
				"message": x.Error(),
			})
			break
		default:
			if x != nil {
				c.JSON(http.StatusOK, gin.H{
					"code": 0,
					"data": x,
				})
			} else {
				c.Status(http.StatusNotFound)
			}
		}
	}
}
