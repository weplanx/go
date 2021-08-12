package cookie

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Option struct {
	MaxAge   int    `mapstructure:"max_age"`
	Path     string `mapstructure:"path"`
	Domain   string `mapstructure:"domain"`
	Secure   bool   `mapstructure:"secure"`
	HttpOnly bool   `mapstructure:"http_only"`
	SameSite string `mapstructure:"same_site"`
}

type Cookie struct {
	Option
	HttpSameSite http.SameSite
}

func (x *Cookie) Get(c *gin.Context, name string) (string, error) {
	return c.Cookie(name)
}

func (x *Cookie) Set(c *gin.Context, name string, value string) {
	c.SetCookie(name, value, x.MaxAge, x.Path, x.Domain, x.Secure, x.HttpOnly)
	c.SetSameSite(x.HttpSameSite)
}

func (x *Cookie) Del(c *gin.Context, name string) {
	x.Set(c, name, "")
}
