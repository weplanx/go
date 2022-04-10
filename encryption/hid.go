package encryption

import "github.com/speps/go-hashids/v2"

type HID struct {
	HashID *hashids.HashID
}

func NewIDx(key string, alphabet string) (x *HID, err error) {
	x = new(HID)
	if x.HashID, err = hashids.NewWithData(&hashids.HashIDData{
		Alphabet: alphabet,
		Salt:     key,
	}); err != nil {
		return
	}
	return
}

// EncodeId ID encryption
func (x *HID) EncodeId(value []int) (string, error) {
	return x.HashID.Encode(value)
}

// DecodeId ID decryption
func (x *HID) DecodeId(value string) ([]int, error) {
	return x.HashID.DecodeWithError(value)
}
