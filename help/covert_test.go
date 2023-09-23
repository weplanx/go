package help_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/help"
	"testing"
)

func TestReverse(t *testing.T) {
	v := []string{"a", "b", "c"}
	help.Reverse(v)
	assert.Equal(t, []string{"c", "b", "a"}, v)
	t.Log(v)
}

func TestShuffle(t *testing.T) {
	v := []int{1, 2, 3, 4, 5, 6, 7}
	help.Shuffle(v)
	t.Log(v)
}

func TestReverseString(t *testing.T) {
	v := help.ReverseString("abcdefg")
	assert.Equal(t, "gfedcba", v)
	t.Log(v)
}

func TestShuffleString(t *testing.T) {
	v := help.ShuffleString("abcdefg")
	t.Log(v)
}
