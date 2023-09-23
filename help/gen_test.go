package help_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/help"
	"testing"
)

func TestUuid(t *testing.T) {
	v := help.Uuid()
	_, err := uuid.Parse(v)
	assert.NoError(t, err)
}
