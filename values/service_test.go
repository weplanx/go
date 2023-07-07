package values_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"testing"
)

func TestService_Set(t *testing.T) {
	t.Log("ok")
	b, _ := msgpack.Marshal(M{
		"token": "WasY11AEfAVXZ68c",
	})
	err := service.Set(b)
	assert.NoError(t, err)
}
