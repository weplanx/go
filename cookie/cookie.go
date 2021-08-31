package cookie

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Option struct {

	// The Expires attribute defines a specific date and time for when the browser should delete the cookie.
	MaxAge int `yaml:"max_age" env:"COOKIE_MAX_AGE"`

	// The Path attributes define the scope of the cookie.
	Path string `yaml:"path" env:"COOKIE_PATH"`

	// The Domain attributes define the scope of the cookie.
	Domain string `yaml:"domain" env:"COOKIE_DOMAIN"`

	// A secure cookie can only be transmitted over an encrypted connection (i.e. HTTPS).
	Secure bool `yaml:"secure" env:"COOKIE_SECURE"`

	// An http-only cookie cannot be accessed by client-side APIs, such as JavaScript.
	HttpOnly bool `yaml:"http_only" env:"COOKIE_HTTP_ONLY"`
}

type Cookie struct {
	Option

	// The attribute SameSite can have a value of `strict`, `lax` or `none`.
	//
	// With attribute SameSite=Strict, the browsers would only send cookies to a target domain that is the same as the origin domain.
	// This would effectively mitigate cross-site request forgery (CSRF) attacks.
	//
	// With SameSite=Lax, browsers would send cookies with requests to a target domain even it is different from the origin domain,
	// but only for safe requests such as GET (POST is unsafe) and not third-party cookies (inside iframe).
	//
	// Attribute SameSite=None would allow third-party (cross-site) cookies,
	// however, most browsers require secure attribute on SameSite=None cookies.
	SameSite http.SameSite
}

// New create a unified cookie configuration
func New(option Option, samesite http.SameSite) *Cookie {
	return &Cookie{
		option,
		samesite,
	}
}

// Get get a cookie
func (x *Cookie) Get(c *gin.Context, name string) (string, error) {
	return c.Cookie(name)
}

// Set create or update a cookie
func (x *Cookie) Set(c *gin.Context, name string, value string) {
	c.SetCookie(name, value, x.MaxAge, x.Path, x.Domain, x.Secure, x.HttpOnly)
	c.SetSameSite(x.SameSite)
}

// Del clear a cookie
func (x *Cookie) Del(c *gin.Context, name string) {
	x.Set(c, name, "")
}
