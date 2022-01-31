package encryption

import "github.com/speps/go-hashids/v2"

type IDx struct {
	HashID *hashids.HashID
}

func NewIDx(key string, alphabet string) (x *IDx, err error) {
	x = new(IDx)
	if x.HashID, err = hashids.NewWithData(&hashids.HashIDData{
		Alphabet: alphabet,
		Salt:     key,
	}); err != nil {
		return
	}
	return
}

// EncodeId ID encryption
func (x *IDx) EncodeId(value []int) (string, error) {
	return x.HashID.Encode(value)
}

// DecodeId ID decryption
func (x *IDx) DecodeId(value string) ([]int, error) {
	return x.HashID.DecodeWithError(value)
}
