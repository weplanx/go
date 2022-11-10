package passport

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type Passport struct {
	Namespace string
	Key       string
}

func New(namespace string, key string) *Passport {
	return &Passport{
		Namespace: namespace,
		Key:       key,
	}
}

type Claims struct {
	UserId string
	jwt.RegisteredClaims
}

// Create 生成令牌
func (x *Passport) Create(userId string, jti string) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    x.Namespace,
			ID:        jti,
		},
	})
	return token.SignedString([]byte(x.Key))
}

// Verify 验证令牌
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
