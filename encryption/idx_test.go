package encryption

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIdx(t *testing.T) {
	idx, err := NewIDx("6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK")
	assert.Nil(t, err)
	hash, err := idx.EncodeId([]int{651})
	assert.Nil(t, err)
	val, err := idx.DecodeId(hash)
	assert.Nil(t, err)
	assert.Equal(t, val, []int{651})
}
