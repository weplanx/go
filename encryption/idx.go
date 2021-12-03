package encryption

import "github.com/speps/go-hashids/v2"

type IDx struct {
	HashIDData *hashids.HashIDData
	HashID     *hashids.HashID
}

func NewIDx(key string) (x *IDx, err error) {
	x = new(IDx)
	x.HashIDData = hashids.NewData()
	x.HashIDData.Salt = key
	if x.HashID, err = hashids.NewWithData(x.HashIDData); err != nil {
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
