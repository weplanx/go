package csrf

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/gookit/goutil/strutil"
	"net/http"
)

type Csrf struct {
	Key           string
	CookieName    string
	SaltName      string
	HeaderName    string
	Domain        string
	IgnoreMethods map[string]bool
}

var (
	ErrMissingHeader = errors.New("CSRF missing csrf token in header")
	ErrInvalidToken  = errors.New("CSRF invalid token")
)

func New(options ...Option) *Csrf {
	x := &Csrf{
		CookieName: "XSRF-TOKEN",
		SaltName:   "XSRF-SALT",
		HeaderName: "X-XSRF-TOKEN",
		Domain:     "",
		IgnoreMethods: map[string]bool{
			"GET":     true,
			"HEAD":    true,
			"OPTIONS": true,
			"TRACE":   true,
		},
	}
	for _, v := range options {
		v(x)
	}
	return x
}

type Option func(x *Csrf)

func SetKey(v string) Option {
	return func(x *Csrf) {
		x.Key = v
	}
}

func SetCookieName(v string) Option {
	return func(x *Csrf) {
		x.CookieName = v
	}
}

func SetSaltName(v string) Option {
	return func(x *Csrf) {
		x.SaltName = v
	}
}

func SetHeaderName(v string) Option {
	return func(x *Csrf) {
		x.HeaderName = v
	}
}

func SetIgnoreMethods(methods []string) Option {
	return func(x *Csrf) {
		x.IgnoreMethods = map[string]bool{}
		for _, v := range methods {
			x.IgnoreMethods[v] = true
		}
	}
}

func SetDomain(v string) Option {
	return func(x *Csrf) {
		x.Domain = v
	}
}

func (x *Csrf) SetToken(c *app.RequestContext) {
	salt := strutil.MicroTimeHexID()
	c.SetCookie(x.SaltName, salt, 86400, "/", "", protocol.CookieSameSiteStrictMode, true, true)
	c.SetCookie(x.CookieName, x.Tokenize(salt), 86400, "/", x.Domain, protocol.CookieSameSiteStrictMode, true, false)
}

func (x *Csrf) Tokenize(salt string) string {
	h := hmac.New(sha256.New, []byte(x.Key))
	h.Write([]byte(salt))
	return hex.EncodeToString(h.Sum(nil))
}

func (x *Csrf) VerifyToken(skip bool) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if skip {
			c.Next(ctx)
			return
		}

		if x.IgnoreMethods[string(c.Method())] {
			c.Next(ctx)
			return
		}

		salt := string(c.Cookie(x.SaltName))

		extractor := c.GetHeader(x.HeaderName)
		if extractor == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, utils.H{
				"code":    0,
				"message": ErrMissingHeader.Error(),
			})
			return
		}

		if x.Tokenize(salt) != string(extractor) {
			c.AbortWithStatusJSON(http.StatusBadRequest, utils.H{
				"code":    0,
				"message": ErrInvalidToken.Error(),
			})
			return
		}

		c.Next(ctx)
	}
}
