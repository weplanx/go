package cipher

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var c *Cipher
var err error

func TestMain(m *testing.M) {
	c, err = New("6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK")
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestDexId(t *testing.T) {
	hash, err := c.EncodeId([]int{651})
	assert.Nil(t, err)
	val, err := c.DecodeId(hash)
	assert.Nil(t, err)
	assert.Equal(t, val, []int{651})
}

func TestDexData(t *testing.T) {
	data := []byte("Gophers, gophers, gophers everywhere!")
	ciphertext, err := c.Encode(data)
	assert.Nil(t, err)
	result, err := c.Decode(ciphertext)
	assert.Nil(t, err)
	assert.Equal(t, data, result)
}
