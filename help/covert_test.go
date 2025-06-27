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

type Sign struct {
	input    map[string]any
	expected string
}

func TestMapToSignText(t *testing.T) {
	mocks := []Sign{
		{
			input:    map[string]any{},
			expected: "",
		},
		{
			input:    map[string]any{"key1": "value1"},
			expected: "key1=value1",
		},
		{
			input: map[string]any{
				"b": "2",
				"a": 1,
				"c": "3",
			},
			expected: "a=1&b=2&c=3",
		},
		{
			input: map[string]any{
				"key2": true,
				"key1": "value1",
				"key3": "value3",
			},
			expected: "key1=value1&key2=true&key3=value3",
		},
		{
			input: map[string]any{
				"key3": "",
				"key4": "",
				"key1": "123",
				"key2": "value2",
			},
			expected: "key1=123&key2=value2",
		},
	}
	for _, m := range mocks {
		result := help.MapToSignText(m.input)
		assert.Equal(t, m.expected, result)
	}
}
