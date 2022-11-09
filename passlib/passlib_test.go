package passlib_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/utils/passlib"
	"testing"
)

func TestPassword(t *testing.T) {
	hash, err := passlib.Hash("pass@VAN1234")
	assert.NoError(t, err)
	match, err := passlib.Verify("pass@VAN1234", "asdaqweqwexcxzcqweqw")
	assert.NotNil(t, err)
	assert.False(t, match)
	match, err = passlib.Verify("pass@VAN1235", hash)
	assert.NoError(t, err)
	assert.False(t, match)
	match, err = passlib.Verify("pass@VAN1234", hash)
	assert.NoError(t, err)
	assert.True(t, match)
}
