package passport_test

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/passport"
	"os"
	"testing"
	"time"
)

var x1 *passport.Passport
var x2 *passport.Passport

var key1 = "hZXD^@K9%wydDC3Z@cyDvE%5bz9SP7gy"

func TestMain(m *testing.M) {
	x1 = passport.New(
		passport.SetIssuer("dev"),
		passport.SetKey(key1),
	)
	x2 = passport.New(
		passport.SetIssuer("beta"),
		passport.SetKey("eK4qpn7yCBLo0u5mlAFFRCRsCmf2NQ76"),
	)
	os.Exit(m.Run())
}

var jti1 = "GIlmuxUX1n5N4wAVVF40i"
var userId1 = "FTFD1FnWKwueHAY8h-zXg"
var jti2 = "gxOWtI58ViI2pl3BHxSNs"
var userId2 = "HU3kev7LZEgoaghpIMrGn"
var token string
var otherToken string

func TestCreate(t *testing.T) {
	var err error
	token, err = x1.Create(userId1, jti1, time.Hour*2)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	otherToken, err = x2.Create(userId2, jti2, time.Hour*2)
	assert.NoError(t, err)
	assert.NotEmpty(t, otherToken)
}

func TestVerify(t *testing.T) {
	var err error
	var clamis1 passport.Claims
	clamis1, err = x1.Verify(token)
	assert.NoError(t, err)
	assert.Equal(t, clamis1.ID, jti1)
	assert.Equal(t, clamis1.UserId, userId1)
	assert.Equal(t, clamis1.Issuer, x1.Issuer)
	var clamis2 passport.Claims
	clamis2, err = x2.Verify(otherToken)
	assert.NoError(t, err)
	assert.Equal(t, clamis2.ID, jti2)
	assert.Equal(t, clamis2.UserId, userId2)
	assert.Equal(t, clamis2.Issuer, x2.Issuer)

	_, err = x1.Verify(otherToken)
	assert.Error(t, err)
	_, err = x2.Verify(token)
	assert.Error(t, err)
}

func TestSigningMethodHS384(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, passport.Claims{
		UserId: userId1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "dev",
			ID:        jti1,
		},
	})
	ts, err := token.SignedString([]byte(key1))
	assert.NoError(t, err)
	_, err = x1.Verify(ts)
	assert.NoError(t, err)
}

var ecPKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIAh5qA3rmqQQuu0vbKV/+zouz/y/Iy2pLpIcWUSyImSwoAoGCCqGSM49
AwEHoUQDQgAEYD54V/vp+54P9DXarYqx4MPcm+HKRIQzNasYSoRQHQ/6S6Ps8tpM
cT+KvIIC8W/e9k0W7Cm72M1P9jU7SLf/vg==
-----END EC PRIVATE KEY-----`

func TestOtherSigningMethod(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, passport.Claims{
		UserId: userId1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "dev",
			ID:        jti1,
		},
	})
	ecdsaKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(ecPKey))
	assert.NoError(t, err)
	ts, err := token.SignedString(ecdsaKey)
	assert.NoError(t, err)
	_, err = x1.Verify(ts)
	assert.Error(t, err)
	t.Log(err)
}
