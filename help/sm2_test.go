package help_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/smx509"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/help"
	"testing"
)

var priKeyStr = `MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg1QG/R5oI4OO3mSh7Nss5RP7d8rV571CCyW+7cI1+w5qgCgYIKoEcz1UBgi2hRANCAAShrB20h+g1nL++oRUMpCsqdAb+ALVoUSpnR4jencQj3arGNQJA9rSdmvh6k64eI6gLZNxxk2YXXm5A70a/s1iz`

func TestSm2(t *testing.T) {
	priKey, err := help.PrivKeySM2FromBase64(priKeyStr)
	assert.NoError(t, err)

	sig, err := help.Sm2Sign(priKey, `Hello world!`)
	assert.NoError(t, err)
	t.Log(sig)

	pub := priKey.PublicKey
	ecdsaPub := &ecdsa.PublicKey{
		Curve: pub.Curve,
		X:     pub.X,
		Y:     pub.Y,
	}

	r, err := help.Sm2Verify(ecdsaPub, `Hello world`, sig)
	assert.NoError(t, err)
	assert.False(t, r)

	r, err = help.Sm2Verify(ecdsaPub, `Hello world!`, sig)
	assert.NoError(t, err)
	assert.True(t, r)
}

func TestSm2PublicKey(t *testing.T) {
	priKey, err := sm2.GenerateKey(rand.Reader)
	assert.NoError(t, err)
	b, err := smx509.MarshalPKIXPublicKey(&priKey.PublicKey)
	assert.NoError(t, err)

	pubKeyStr := base64.StdEncoding.EncodeToString(b)
	t.Log(pubKeyStr)
	pubKey, err := help.PubKeySM2FromBase64(pubKeyStr)
	assert.NoError(t, err)

	t.Log(pubKey)
	t.Log(sm2.IsSM2PublicKey(pubKey))
}
