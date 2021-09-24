package helper

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"
)

var cipherHelper *CipherHelper

func TestNewCipherHelper(t *testing.T) {
	var err error
	if cipherHelper, err = NewCipherHelper("6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK"); err != nil {
		t.Error(err)
	}
}

func TestCipherHelper_Id(t *testing.T) {
	hash, err := cipherHelper.EncodeId([]int{651})
	assert.Nil(t, err)
	val, err := cipherHelper.DecodeId(hash)
	assert.Nil(t, err)
	assert.Equal(t, val, []int{651})
}

func TestCipherHelper_Data(t *testing.T) {
	data := []byte("Gophers, gophers, gophers everywhere!")
	ciphertext, err := cipherHelper.Encode(data)
	assert.Nil(t, err)
	result, err := cipherHelper.Decode(ciphertext)
	assert.Nil(t, err)
	assert.Equal(t, data, result)
	_, err = cipherHelper.Decode(base64.StdEncoding.EncodeToString(data))
	assert.Error(t, err)
}
