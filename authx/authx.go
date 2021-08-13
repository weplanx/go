package authx

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/kainonly/go-bit/cookie"
	"github.com/kainonly/go-bit/str"
	"time"
)

var (
	Expired             = errors.New("authentication has expired")
	RefreshTokenInvalid = errors.New("refresh token is invalid")
)

type Authx struct {
	Scenes map[string]*Auth
}

func New(options map[string]Option) *Authx {
	scenes := make(map[string]*Auth)
	for k, v := range options {
		scenes[k] = &Auth{
			Option: v,
		}
	}
	return &Authx{scenes}
}

type Option struct {
	Key string   `yaml:"key"`
	Iss string   `yaml:"iss"`
	Aud []string `yaml:"aud"`
	Nbf int64    `yaml:"nbf"`
	Exp int64    `yaml:"exp"`
}

type Auth struct {
	Option
	Sub       string
	cookie    *cookie.Cookie
	refreshFn RefreshFn
}

func (x *Authx) Make(name string, cookie *cookie.Cookie, refreshFn RefreshFn) *Auth {
	auth := x.Scenes[name]
	auth.Sub = name
	auth.cookie = cookie
	auth.refreshFn = refreshFn
	return auth
}

// Create 创建认证
func (x *Auth) Create(c *gin.Context, uid string, data interface{}) (tokenString string, err error) {
	claims := jwt.MapClaims{
		"iat":  time.Now().Unix(),
		"nbf":  time.Now().Add(time.Second * time.Duration(x.Nbf)).Unix(),
		"exp":  time.Now().Add(time.Second * time.Duration(x.Exp)).Unix(),
		"jti":  str.Uuid().String(),
		"iss":  x.Iss,
		"aud":  x.Aud,
		"sub":  x.Sub,
		"uid":  uid,
		"data": data,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if tokenString, err = token.SignedString([]byte(x.Key)); err != nil {
		return
	}
	if x.cookie != nil {
		x.cookie.Set(c, x.Sub+":access_token", tokenString)
	}
	if x.refreshFn != nil {
		x.refreshFn.Factory(claims)
	}
	c.Set("claims", claims)
	return
}

// Verify 鉴权认证
func (x *Auth) Verify(c *gin.Context, args ...interface{}) (err error) {
	var tokenString string
	if x.cookie != nil {
		if tokenString, err = x.cookie.Get(c, x.Sub+":access_token"); err != nil {
			return Expired
		}
	} else {
		if len(args) != 0 {
			tokenString = args[0].(string)
		} else {
			return Expired
		}
	}
	if tokenString == "" {
		return Expired
	}
	var token *jwt.Token
	if token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(x.Key), nil
	}); err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors == jwt.ValidationErrorExpired && x.refreshFn != nil && token != nil {
				claims := token.Claims.(jwt.MapClaims)
				if result := x.refreshFn.Verify(claims); !result {
					return RefreshTokenInvalid
				}
				updateClaims := jwt.MapClaims{
					"iat":  time.Now().Unix(),
					"nbf":  time.Now().Add(time.Second * time.Duration(x.Nbf)).Unix(),
					"exp":  time.Now().Add(time.Second * time.Duration(x.Exp)).Unix(),
					"jti":  str.Uuid().String(),
					"iss":  claims["iss"],
					"aud":  claims["aud"],
					"sub":  claims["sub"],
					"uid":  claims["uid"],
					"data": claims["data"],
				}
				token = jwt.NewWithClaims(jwt.SigningMethodHS256, updateClaims)
				if tokenString, err = token.SignedString([]byte(x.Key)); err != nil {
					return
				}
				if x.cookie != nil {
					x.cookie.Set(c, x.Sub+":access_token", tokenString)
				}
				c.Set("token", token)
			}
		}
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	if x.refreshFn != nil {
		x.refreshFn.Renewal(claims)
	}
	c.Set("claims", claims)
	return
}

// Destory 销毁认证
func (x *Auth) Destory(c *gin.Context, args ...interface{}) (err error) {
	if err = x.Verify(c, args); err != nil {
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		return fmt.Errorf("environment verification is abnormal")
	}
	if x.cookie != nil {
		x.cookie.Del(c, x.Sub+":access_token")
	}
	if x.refreshFn != nil {
		if err = x.refreshFn.Destory(claims.(jwt.MapClaims)); err != nil {
			return
		}
	}
	return
}

// Middleware 鉴权认证中间件
func Middleware(auth Auth, args ...interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := auth.Verify(c, args); err != nil {
			c.AbortWithStatusJSON(200, gin.H{
				"error": 0,
				"msg":   err.Error(),
			})
			return
		}
		c.Next()
	}
}
