package encryption

import (
	"github.com/speps/go-hashids/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHid(t *testing.T) {
	idx, err := NewIDx("6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK", hashids.DefaultAlphabet)
	assert.Nil(t, err)
	hash, err := idx.EncodeId([]int{651})
	assert.Nil(t, err)
	val, err := idx.DecodeId(hash)
	assert.Nil(t, err)
	assert.Equal(t, []int{651}, val)
	_, err = NewIDx("6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK", "abcd")
	assert.Error(t, err)
}
