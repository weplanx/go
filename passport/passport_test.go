package passport_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go-wpx/passport"
	"os"
	"testing"
)

var x1 *passport.Passport
var x2 *passport.Passport

func TestMain(m *testing.M) {
	x1 = passport.New(
		passport.SetNamespace("dev"),
		passport.SetKey("hZXD^@K9%wydDC3Z@cyDvE%5bz9SP7gy"),
	)
	x2 = passport.New(
		passport.SetNamespace("beta"),
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
	token, err = x1.Create(userId1, jti1)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	otherToken, err = x2.Create(userId2, jti2)
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
	assert.Equal(t, clamis1.Issuer, x1.Namespace)
	var clamis2 passport.Claims
	clamis2, err = x2.Verify(otherToken)
	assert.NoError(t, err)
	assert.Equal(t, clamis2.ID, jti2)
	assert.Equal(t, clamis2.UserId, userId2)
	assert.Equal(t, clamis2.Issuer, x2.Namespace)
}

func TestVerifyBad(t *testing.T) {
	_, err := x1.Verify(otherToken)
	assert.Error(t, err)
	_, err = x2.Verify(token)
	assert.Error(t, err)
}
