package encryption

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCipher(t *testing.T) {
	_, err := NewCipher("123456")
	assert.Error(t, err)
	// 必须是32位的密钥
	cipher, err := NewCipher("6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK")
	assert.Nil(t, err)
	xcipher, err := NewCipher("74rILbVooYLirHrQJcslHEAvKZI7PKF9")
	assert.Nil(t, err)
	text := []byte("Gophers, gophers, gophers everywhere!")
	encryptedText, err := cipher.Encode(text)
	t.Log(encryptedText)
	assert.Nil(t, err)
	decryptedText, err := cipher.Decode(encryptedText)
	assert.Nil(t, err)
	assert.Equal(t, string(decryptedText), "Gophers, gophers, gophers everywhere!")
	_, err = xcipher.Decode(encryptedText)
	assert.Error(t, err)
}
