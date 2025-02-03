package help_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/help"
	"testing"
)

func TestPtr(t *testing.T) {
	assert.Equal(t, "hello", *help.Ptr[string]("hello"))
	assert.Equal(t, int64(123), *help.Ptr[int64](123))
	assert.Equal(t, false, *help.Ptr[bool](false))
}

func TestIsEmpty(t *testing.T) {
	assert.True(t, help.IsEmpty(nil))
	assert.True(t, help.IsEmpty(""))
	assert.True(t, help.IsEmpty(0))
	assert.True(t, help.IsEmpty(false))
	assert.True(t, help.IsEmpty([]string{}))
	assert.False(t, help.IsEmpty(help.Ptr[int64](0)))
	var a *string
	assert.True(t, help.IsEmpty(a))
	var b struct{}
	assert.True(t, help.IsEmpty(b))
}
