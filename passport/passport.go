package passport

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

var (
	ErrAuthExpired  = errors.New("Authentication token has expired")
	ErrAuthConflict = errors.New("Authentication already taken by new client")
)

type Passport struct {

	// Key used for signing
	Key string `yaml:"key"`

	// Token option
	Option `yaml:"option"`
}

type Option struct {
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

// New authentication
func New(key string, option Option) *Passport {
	return &Passport{
		Key:    key,
		Option: option,
	}
}

// Create authentication token
func (x *Passport) Create(jti string, context map[string]interface{}) (tokenString string, err error) {
	claims := jwt.MapClaims{
		"iat":     time.Now().Unix(),
		"nbf":     time.Now().Add(time.Second * time.Duration(x.Nbf)).Unix(),
		"exp":     time.Now().Add(time.Second * time.Duration(x.Exp)).Unix(),
		"jti":     jti,
		"iss":     x.Iss,
		"aud":     x.Aud,
		"sub":     x.Sub,
		"context": context,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(x.Key))
}

// Verify authentication token
func (x *Passport) Verify(tokenString string) (claims jwt.MapClaims, err error) {
	var token *jwt.Token
	if token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("验证失败，签名方式不一致")
		}
		return []byte(x.Key), nil
	}); err != nil {
		return
	}
	return token.Claims.(jwt.MapClaims), nil
}

type Claims struct {
	jwt.MapClaims
	Context map[string]interface{}
}

func (x *Passport) GetClaims(c *gin.Context, key string) (*Claims, error) {
	value, exists := c.Get(key)
	if !exists {
		c.Set("status_code", 401)
		c.Set("code", "AUTH_EXPIRED")
		return nil, ErrAuthExpired
	}
	claims := value.(jwt.MapClaims)
	return &Claims{
		MapClaims: claims,
		Context:   claims["context"].(map[string]interface{}),
	}, nil
}
