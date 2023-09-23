package help_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/help"
	"testing"
)

func TestRandom(t *testing.T) {
	v1 := help.Random(16)
	assert.Len(t, v1, 16)
	t.Log(v1)
	v2 := help.Random(32)
	assert.Len(t, v2, 32)
	t.Log(v2)
	v3 := help.Random(8)
	assert.Len(t, v3, 8)
	t.Log(v3)
	v4 := help.Random(32, "0123456789abcdef")
	assert.Len(t, v4, 32)
	t.Log(v4)
	v5 := help.Random(64, "0123456789abcdef")
	assert.Len(t, v5, 64)
	t.Log(v5)
}

func TestRandomNumber(t *testing.T) {
	v := help.RandomNumber(6)
	assert.Len(t, v, 6)
	t.Log(v)
}

func TestRandomAlphabet(t *testing.T) {
	v := help.RandomAlphabet(16)
	assert.Len(t, v, 16)
	t.Log(v)
}

func TestRandomUppercase(t *testing.T) {
	v := help.RandomUppercase(8)
	assert.Len(t, v, 8)
	t.Log(v)
}

func TestRandomLowercase(t *testing.T) {
	v := help.RandomLowercase(8)
	assert.Len(t, v, 8)
	t.Log(v)
}
