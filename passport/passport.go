package passport

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Passport struct {
	Issuer string
	Key    string
}

func New(options ...Option) *Passport {
	x := new(Passport)
	for _, v := range options {
		v(x)
	}
	return x
}

type Option func(x *Passport)

func SetIssuer(v string) Option {
	return func(x *Passport) {
		x.Issuer = v
	}
}

func SetKey(v string) Option {
	return func(x *Passport) {
		x.Key = v
	}
}

type Claims struct {
	ActiveId string
	Data     map[string]interface{}

	jwt.RegisteredClaims
}

func NewClaims(activeId string, expire time.Duration) *Claims {
	return &Claims{
		ActiveId: activeId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
}

func (x *Claims) SetJTI(v string) *Claims {
	x.RegisteredClaims.ID = v
	return x
}

func (x *Claims) SetIssuer(v string) *Claims {
	x.Issuer = v
	return x
}

func (x *Claims) SetData(v map[string]interface{}) *Claims {
	x.Data = v
	return x
}

func (x *Passport) Create(claims *Claims) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(x.Key))
}

func (x *Passport) Verify(tokenString string) (claims Claims, err error) {
	if _, err = jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(x.Key), nil
	}); err != nil {
		return
	}
	return
}
