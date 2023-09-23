package help_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/help"
	"testing"
)

func TestHelp(t *testing.T) {
	assert.True(t, help.IsEmpty(nil))
	assert.True(t, help.IsEmpty(""))
	assert.True(t, help.IsEmpty(0))
	assert.False(t, help.IsEmpty(help.Ptr(0)))
	var a *string
	assert.True(t, help.IsEmpty(a))
	var b struct{}
	assert.True(t, help.IsEmpty(b))
}
