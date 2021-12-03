package encryption

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCipher(t *testing.T) {
	cipher, err := NewCipher("6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK")
	assert.Nil(t, err)
	text := []byte("Gophers, gophers, gophers everywhere!")
	encryptedText, err := cipher.Encode(text)
	assert.Nil(t, err)
	t.Log(encryptedText)
	decryptedText, err := cipher.Decode(encryptedText)
	assert.Nil(t, err)
	assert.Equal(t, string(decryptedText), "Gophers, gophers, gophers everywhere!")
}
