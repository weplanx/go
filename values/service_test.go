package values_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/weplanx/go-wpx/values"
	"testing"
)

func TestService_Fetch(t *testing.T) {
	//data := make(map[string]interface{})
	//err := service.Fetch(data)
	//assert.NoError(t, err)
	//t.Log(data)
	b, err := msgpack.Marshal(values.DEFAULT)
	ciphertext, err := cipherx.Encode(b)
	assert.NoError(t, err)
	_, err = keyvalue.PutString("values", ciphertext)
	assert.NoError(t, err)
}

func TestService_Fetch2(t *testing.T) {
	entry, err := keyvalue.Get("values")
	assert.NoError(t, err)
	b, err := cipherx.Decode(string(entry.Value()))
	assert.NoError(t, err)
	var data values.Values
	err = msgpack.Unmarshal(b, &data)
	assert.NoError(t, err)
	t.Log(data)
}

//func TestService_Set(t *testing.T) {
//	t.Log("ok")
//	b, _ := msgpack.Marshal(M{
//		"token": "WasY11AEfAVXZ68c",
//	})
//	err := service.Set(b)
//	assert.NoError(t, err)
//}
