package authx

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"time"
)

var (
	Expired = errors.New("authentication has expired")
)

type Authx struct {
	Scenes map[string]*Auth
}

func New(scenes map[string]*Auth) *Authx {
	return &Authx{scenes}
}

type Auth struct {
	Key string   `yaml:"key"`
	Iss string   `yaml:"iss"`
	Aud []string `yaml:"aud"`
	Sub string   `yaml:"sub"`
	Nbf int64    `yaml:"nbf"`
	Exp int64    `yaml:"exp"`
}

func (x *Authx) Make(name string) *Auth {
	auth := x.Scenes[name]
	auth.Sub = name
	return auth
}

// Create 创建认证
func (x *Auth) Create(jti string, uid string, data interface{}) (tokenString string, err error) {
	claims := jwt.MapClaims{
		"iat":  time.Now().Unix(),
		"nbf":  time.Now().Add(time.Second * time.Duration(x.Nbf)).Unix(),
		"exp":  time.Now().Add(time.Second * time.Duration(x.Exp)).Unix(),
		"jti":  jti,
		"iss":  x.Iss,
		"aud":  x.Aud,
		"sub":  x.Sub,
		"uid":  uid,
		"data": data,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(x.Key))
}

// Verify 鉴权认证
func (x *Auth) Verify(tokenString string) (claims jwt.MapClaims, err error) {
	if tokenString == "" {
		return nil, Expired
	}
	var token *jwt.Token
	if token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(x.Key), nil
	}); err != nil {
		return
	}
	return token.Claims.(jwt.MapClaims), nil
}
