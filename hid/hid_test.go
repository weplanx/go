package hid_test

import (
	"github.com/speps/go-hashids/v2"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/utils/hid"
	"testing"
)

var x *hid.HID

func TestNewHID(t *testing.T) {
	var err error
	x, err = hid.New("6ixSiEXaqxsJTozbnxQ76CWdZXB2JazK", hashids.DefaultAlphabet)
	assert.NoError(t, err)
}

var hash string
var value = []int{651}

func TestHID_Encode(t *testing.T) {
	var err error
	hash, err = x.Encode(value)
	assert.NoError(t, err)
}

func TestHID_Decode(t *testing.T) {
	data, err := x.Decode(hash)
	assert.NoError(t, err)
	assert.Equal(t, value, data)
}
