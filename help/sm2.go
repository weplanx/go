package help

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/smx509"
)

var SM2UID = []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38}

func PubKeySM2FromBase64(v string) (pubKey *ecdsa.PublicKey, err error) {
	var der []byte
	if der, err = base64.StdEncoding.DecodeString(v); err != nil {
		return
	}
	var key any
	if key, err = smx509.ParsePKIXPublicKey(der); err != nil {
		return
	}
	return key.(*ecdsa.PublicKey), nil
}

func PrivKeySM2FromBase64(v string) (priKey *sm2.PrivateKey, err error) {
	var der []byte
	if der, err = base64.StdEncoding.DecodeString(v); err != nil {
		return
	}
	var key any
	if key, err = smx509.ParsePKCS8PrivateKey(der); err != nil {
		return
	}
	return key.(*sm2.PrivateKey), nil
}

func Sm2Sign(key *sm2.PrivateKey, text string) (_ string, err error) {
	var signature []byte
	if signature, err = key.Sign(rand.Reader, []byte(text), sm2.DefaultSM2SignerOpts); err != nil {
		return
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func Sm2Verify(pubKey *ecdsa.PublicKey, text string, sign string) (_ bool, err error) {
	var b []byte
	if b, err = base64.StdEncoding.DecodeString(sign); err != nil {
		return
	}
	return sm2.VerifyASN1WithSM2(pubKey, SM2UID, []byte(text), b), nil
}
