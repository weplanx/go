package hash

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var hashText string

func TestMake(t *testing.T) {
	hash, err := Make(`pass@VAN1234`)
	assert.Nil(t, err)
	hashText = hash
}

func TestCheck(t *testing.T) {
	err := Verify(`pass@VAN1234`, hashText)
	assert.Nil(t, err)
}
