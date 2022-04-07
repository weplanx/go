package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weplanx/go/engine"
	"net/http"
)

type Option struct {
	Model string
}

type OptionFunc func(*Option)

func SetModel(v string) OptionFunc {
	return func(option *Option) {
		option.Model = v
	}
}

func Use(fn func(c *gin.Context) interface{}, options ...OptionFunc) gin.HandlerFunc {
	opt := new(Option)
	for _, v := range options {
		v(opt)
	}
	return func(c *gin.Context) {
		if opt.Model != "" {
			c.Set("model", opt.Model)
		}
		switch x := fn(c).(type) {
		case error:
			statusCode, exists := c.Get("status_code")
			if !exists {
				statusCode = http.StatusBadRequest
			}
			code, exists := c.Get("code")
			if !exists {
				code = "INVALID"
			}
			c.JSON(statusCode.(int), gin.H{
				"code":    code,
				"message": x.Error(),
			})
			break
		default:
			if x != nil {
				statusCode, exists := c.Get("status_code")
				if !exists {
					statusCode = http.StatusOK
				}
				c.JSON(statusCode.(int), x)
			} else {
				c.Status(http.StatusNoContent)
			}
		}
	}
}

func Engine(r *gin.RouterGroup, engine *engine.Controller) {
	r.POST("/:model", Use(engine.Actions))
	r.HEAD("/:model/_count", Use(engine.Count))
	r.HEAD("/:model/_exists", Use(engine.Exists))
	r.GET("/:model", Use(engine.Get))
	r.GET("/:model/:id", Use(engine.GetById))
	r.PATCH("/:model", Use(engine.Patch))
	r.PATCH("/:model/:id", Use(engine.PatchById))
	r.PUT("/:model/:id", Use(engine.Put))
	r.DELETE("/:model/:id", Use(engine.Delete))
}
