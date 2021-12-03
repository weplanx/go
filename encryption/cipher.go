package encryption

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/chacha20poly1305"
)

type Cipher struct {
	AEAD cipher.AEAD
}

func NewCipher(key string) (x *Cipher, err error) {
	x = new(Cipher)
	if x.AEAD, err = chacha20poly1305.NewX([]byte(key)); err != nil {
		return
	}
	return
}

// Encode data encryption
func (x *Cipher) Encode(data []byte) (string, error) {
	nonce := make([]byte, x.AEAD.NonceSize(), x.AEAD.NonceSize()+len(data)+x.AEAD.Overhead())
	rand.Read(nonce)
	encrypted := x.AEAD.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decode data decryption
func (x *Cipher) Decode(text string) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}
	nonce, ciphertext := encrypted[:x.AEAD.NonceSize()], encrypted[x.AEAD.NonceSize():]
	return x.AEAD.Open(nil, nonce, ciphertext, nil)
}
