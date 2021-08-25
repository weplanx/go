package cipher

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"github.com/speps/go-hashids/v2"
	"golang.org/x/crypto/chacha20poly1305"
)

type Cipher struct {
	aead   cipher.AEAD
	hashId *hashids.HashID
}

// New initialize data encryption
func New(key string) (x *Cipher, err error) {
	x = new(Cipher)
	if x.aead, err = chacha20poly1305.NewX([]byte(key)); err != nil {
		return
	}
	hd := hashids.NewData()
	hd.Salt = key
	x.hashId, _ = hashids.NewWithData(hd)
	return
}

// EncodeId ID encryption
func (x *Cipher) EncodeId(value []int) (string, error) {
	return x.hashId.Encode(value)
}

// DecodeId ID decryption
func (x *Cipher) DecodeId(value string) ([]int, error) {
	return x.hashId.DecodeWithError(value)
}

// Encode data encryption
func (x *Cipher) Encode(data []byte) (string, error) {
	nonce := make([]byte, x.aead.NonceSize(), x.aead.NonceSize()+len(data)+x.aead.Overhead())
	rand.Read(nonce)
	encrypted := x.aead.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decode data decryption
func (x *Cipher) Decode(text string) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}
	nonce, ciphertext := encrypted[:x.aead.NonceSize()], encrypted[x.aead.NonceSize():]
	return x.aead.Open(nil, nonce, ciphertext, nil)
}
