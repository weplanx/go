package cookie

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Option struct {
	MaxAge   int
	Path     string
	Domain   string
	Secure   bool
	HttpOnly bool
	SameSite string
}

type Cookie struct {
	Option
	Ctx          *gin.Context
	HttpSameSite http.SameSite
}

func (x *Cookie) Get(name string) (string, error) {
	return x.Ctx.Cookie(name)
}

func (x *Cookie) Set(name string, value string) {
	x.Ctx.SetCookie(name, value, x.MaxAge, x.Path, x.Domain, x.Secure, x.HttpOnly)
	x.Ctx.SetSameSite(x.HttpSameSite)
}

func (x *Cookie) Del(name string) {
	x.Set(name, "")
}
