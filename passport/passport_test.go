package passport_test

import (
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/utils/passport"
	"os"
	"testing"
)

var x *passport.Passport

func TestMain(m *testing.M) {
	x = passport.New("dev", "hZXD^@K9%wydDC3Z@cyDvE%5bz9SP7gy")
	os.Exit(m.Run())
}

func TestPassport(t *testing.T) {
	jti, _ := gonanoid.Nanoid()
	tokenString, err := x.Create("xs1fp", jti)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
	var clamis passport.Claims
	clamis, err = x.Verify(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, clamis.ID, jti)
	assert.Equal(t, clamis.UserId, "xs1fp")
	assert.Equal(t, clamis.Issuer, x.Namespace)
}
