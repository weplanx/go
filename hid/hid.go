package hid

import "github.com/speps/go-hashids/v2"

type HID struct {
	HashID *hashids.HashID
}

func New(key string, alphabet string) (x *HID, err error) {
	x = new(HID)
	if x.HashID, err = hashids.NewWithData(&hashids.HashIDData{
		Alphabet: alphabet,
		Salt:     key,
	}); err != nil {
		return
	}
	return
}

func (x *HID) Encode(value []int) (string, error) {
	return x.HashID.Encode(value)
}

func (x *HID) Decode(value string) ([]int, error) {
	return x.HashID.DecodeWithError(value)
}
