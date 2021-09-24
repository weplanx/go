package helper

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"github.com/speps/go-hashids/v2"
	"golang.org/x/crypto/chacha20poly1305"
)

type CipherHelper struct {
	aead   cipher.AEAD
	hashId *hashids.HashID
}

func NewCipherHelper(key string) (x *CipherHelper, err error) {
	x = new(CipherHelper)
	if x.aead, err = chacha20poly1305.NewX([]byte(key)); err != nil {
		return
	}
	hd := hashids.NewData()
	hd.Salt = key
	if x.hashId, err = hashids.NewWithData(hd); err != nil {
		return
	}
	return
}

// EncodeId ID encryption
func (x *CipherHelper) EncodeId(value []int) (string, error) {
	return x.hashId.Encode(value)
}

// DecodeId ID decryption
func (x *CipherHelper) DecodeId(value string) ([]int, error) {
	return x.hashId.DecodeWithError(value)
}

// Encode data encryption
func (x *CipherHelper) Encode(data []byte) (string, error) {
	nonce := make([]byte, x.aead.NonceSize(), x.aead.NonceSize()+len(data)+x.aead.Overhead())
	rand.Read(nonce)
	encrypted := x.aead.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decode data decryption
func (x *CipherHelper) Decode(text string) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}
	nonce, ciphertext := encrypted[:x.aead.NonceSize()], encrypted[x.aead.NonceSize():]
	return x.aead.Open(nil, nonce, ciphertext, nil)
}
