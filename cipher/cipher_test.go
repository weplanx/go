package cipher_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go-wpx/cipher"
	"testing"
)

var x1 *cipher.Cipher
var x2 *cipher.Cipher

func TestUseCipher(t *testing.T) {
	var err error
	_, err = cipher.New("123456")
	assert.Error(t, err)
	x1, err = cipher.New("6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK")
	assert.NoError(t, err)
	x2, err = cipher.New("74rILbVooYLirHrQJcslHEAvKZI7PKF9")
	assert.NoError(t, err)
}

var text = "Gophers, gophers, gophers everywhere!"
var encryptedText string

func TestCipher_Encode(t *testing.T) {
	var err error
	encryptedText, err = x1.Encode([]byte(text))
	t.Log(encryptedText)
	assert.Nil(t, err)
}

func TestCipher_Decode(t *testing.T) {
	decryptedText, err := x1.Decode(encryptedText)
	assert.NoError(t, err)
	assert.Equal(t, text, string(decryptedText))
	_, err = x2.Decode(encryptedText)
	assert.Error(t, err)
	_, err = x1.Decode("asdasdasd")
	assert.Error(t, err)
}
