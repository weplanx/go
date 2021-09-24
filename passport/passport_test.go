package passport

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var x *Authx

func TestMain(m *testing.M) {
	x = New(map[string]*Auth{
		"system": {
			Key: "6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK",
			Iss: "go",
			Aud: []string{"admin"},
			Nbf: 1,
			Exp: 720,
		},
	})
	os.Exit(m.Run())
}

var auth *Auth

func TestAuthx_Make(t *testing.T) {
	auth = x.Make("system")
	assert.Equal(t, auth.Key, "6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK")
	assert.Equal(t, auth.Iss, "go")
	assert.Equal(t, auth.Aud, []string{"admin"})
	assert.Equal(t, auth.Sub, "system")
	assert.Equal(t, auth.Nbf, int64(1))
	assert.Equal(t, auth.Exp, int64(720))
}

var jti = uuid.New().String()
var tokenString string

func TestAuth_Create(t *testing.T) {
	var err error
	tokenString, err = auth.Create(jti, map[string]interface{}{
		"uid": "xs1fp",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, tokenString)
}

func TestAuth_Verify(t *testing.T) {
	claims, err := auth.Verify(tokenString)
	assert.Error(t, err)
	time.Sleep(time.Second)
	claims, err = auth.Verify(tokenString)
	assert.Nil(t, err)
	assert.Equal(t, claims["jti"], jti)
	assert.Equal(t, claims["iss"], "go")
	assert.Equal(t, claims["aud"], []interface{}{"admin"})
	assert.Equal(t, claims["sub"], "system")
	assert.Equal(t, claims["data"], map[string]interface{}{
		"uid": "xs1fp",
	})

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.StandardClaims{})

	pkey, err := jwt.ParseECPrivateKeyFromPEM([]byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIAh5qA3rmqQQuu0vbKV/+zouz/y/Iy2pLpIcWUSyImSwoAoGCCqGSM49
AwEHoUQDQgAEYD54V/vp+54P9DXarYqx4MPcm+HKRIQzNasYSoRQHQ/6S6Ps8tpM
cT+KvIIC8W/e9k0W7Cm72M1P9jU7SLf/vg==
-----END EC PRIVATE KEY-----`))

	assert.Nil(t, err)
	other, err := token.SignedString(pkey)
	assert.Nil(t, err)
	_, err = auth.Verify(other)
	assert.Error(t, err)
}
