package api

import "github.com/gin-gonic/gin"

type Hook struct {
	body interface{}
}

func StartHook(c *gin.Context) *Hook {
	h := new(Hook)
	c.Set("hook", h)
	return h
}

func (x *Hook) SetBody(value interface{}) {
	x.body = value
}
