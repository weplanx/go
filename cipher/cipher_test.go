package cipher

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var x *Cipher

func TestMain(m *testing.M) {
	var err error
	x, err = New("6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK")
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	_, err := New("JeNl0biq")
	assert.Error(t, err)
}

func TestDexId(t *testing.T) {
	hash, err := x.EncodeId([]int{651})
	assert.Nil(t, err)
	val, err := x.DecodeId(hash)
	assert.Nil(t, err)
	assert.Equal(t, val, []int{651})
}

func TestDexData(t *testing.T) {
	data := []byte("Gophers, gophers, gophers everywhere!")
	ciphertext, err := x.Encode(data)
	assert.Nil(t, err)
	result, err := x.Decode(ciphertext)
	assert.Nil(t, err)
	assert.Equal(t, data, result)
	_, err = x.Decode(base64.StdEncoding.EncodeToString(data))
	assert.Error(t, err)
}
