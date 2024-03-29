package cipher

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/chacha20poly1305"
)

type Cipher struct {
	AEAD cipher.AEAD
}

func New(key string) (x *Cipher, err error) {
	x = new(Cipher)
	if x.AEAD, err = chacha20poly1305.NewX([]byte(key)); err != nil {
		return
	}
	return
}

func (x *Cipher) Encode(data []byte) (ciphertext string, err error) {
	nonce := make([]byte, x.AEAD.NonceSize(), x.AEAD.NonceSize()+len(data)+x.AEAD.Overhead())
	if _, err = rand.Read(nonce); err != nil {
		return
	}
	encrypted := x.AEAD.Seal(nonce, nonce, data, nil)
	ciphertext = base64.StdEncoding.EncodeToString(encrypted)
	return
}

func (x *Cipher) Decode(ciphertext string) (data []byte, err error) {
	var encrypted []byte
	if encrypted, err = base64.StdEncoding.DecodeString(ciphertext); err != nil {
		return
	}
	nonce, text := encrypted[:x.AEAD.NonceSize()], encrypted[x.AEAD.NonceSize():]
	return x.AEAD.Open(nil, nonce, text, nil)
}
