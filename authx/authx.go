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

	// Multi-scene authentication
	Scenes map[string]*Auth
}

// New create authentication
// 	- scenes Multi-scene authentication, the key is equal to subject
func New(scenes map[string]*Auth) *Authx {
	return &Authx{scenes}
}

type Auth struct {

	// Key used for signing
	Key string `yaml:"key"`

	// Identifies principal that issued the JWT
	Iss string `yaml:"iss"`

	// Identifies the recipients that the JWT is intended for
	Aud []string `yaml:"aud"`

	// Identifies the subject of the JWT.
	Sub string `yaml:"sub"`

	// Identifies the time on which the JWT will start to be accepted for processing
	Nbf int64 `yaml:"nbf"`

	// Identifies the expiration time on and after which the JWT must not be accepted for processing
	Exp int64 `yaml:"exp"`
}

// Make obtain scene authorization
// 	- name Scene name
func (x *Authx) Make(name string) *Auth {
	auth := x.Scenes[name]
	auth.Sub = name
	return auth
}

// Create create authentication token
// 	- jti Case-sensitive unique identifier of the token even among different issuers
// 	- data Custom claims
func (x *Auth) Create(jti string, data map[string]interface{}) (tokenString string, err error) {
	claims := jwt.MapClaims{
		"iat":  time.Now().Unix(),
		"nbf":  time.Now().Add(time.Second * time.Duration(x.Nbf)).Unix(),
		"exp":  time.Now().Add(time.Second * time.Duration(x.Exp)).Unix(),
		"jti":  jti,
		"iss":  x.Iss,
		"aud":  x.Aud,
		"sub":  x.Sub,
		"data": data,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(x.Key))
}

// Verify Authentication
// 	- tokenString The token string
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
