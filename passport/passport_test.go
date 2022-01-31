package passport

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var passport *Passport

func TestMain(m *testing.M) {
	passport = New("6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK", Option{
		Iss: "weplanx",
		Sub: "system",
		Aud: []string{"api"},
		Nbf: 1,
		Exp: 720,
	})
	os.Exit(m.Run())
}

func TestPassport(t *testing.T) {
	jti := uuid.New().String()
	tokenString, err := passport.Create(jti, map[string]interface{}{
		"userId": "xs1fp",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, tokenString)
	assert.Nil(t, err)
	t.Log(tokenString)
	var clamis jwt.MapClaims
	_, err = passport.Verify(tokenString)
	assert.Error(t, err)
	time.Sleep(time.Second)
	clamis, err = passport.Verify(tokenString)
	assert.Nil(t, err)
	assert.Equal(t, "weplanx", clamis["iss"])
	assert.Equal(t, "system", clamis["sub"])
	assert.Equal(t, []interface{}{"api"}, clamis["aud"])
	assert.Equal(t, map[string]interface{}{
		"userId": "xs1fp",
	}, clamis["context"])
}
