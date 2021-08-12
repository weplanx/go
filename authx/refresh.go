package authx

import "github.com/golang-jwt/jwt"

type RefreshFn interface {
	Factory(claims jwt.MapClaims, args ...interface{})
	Verify(claims jwt.MapClaims, args ...interface{}) bool
	Renewal(claims jwt.MapClaims, args ...interface{})
	Destory(claims jwt.MapClaims, args ...interface{}) (err error)
}
