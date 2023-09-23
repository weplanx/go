package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"sort"
	"strconv"
	"time"
)

var (
	ErrNotMatch = errors.New("code does not match")
)

type Totp struct {
	Secret        string
	Window        int
	Counter       int
	DisallowReuse []int
	ScratchCodes  []int
}

func (x *Totp) Authenticate(password string) (bool, error) {
	var scratch bool
	switch {
	case len(password) == 6 && password[0] >= '0' && password[0] <= '9':
		break
	case len(password) == 8 && password[0] >= '1' && password[0] <= '9':
		scratch = true
		break
	default:
		return false, ErrNotMatch
	}
	code, err := strconv.Atoi(password)
	if err != nil {
		return false, ErrNotMatch
	}
	if scratch {
		return x.CheckScratchCodes(code), nil
	}
	if x.Counter > 0 {
		return x.CheckCode(code), nil
	}
	ts := int(time.Now().UTC().Unix() / 30)
	return x.CheckTotpCode(ts, code), nil
}

func (x *Totp) CheckScratchCodes(code int) bool {
	for i, v := range x.ScratchCodes {
		if code == v {
			l := len(x.ScratchCodes) - 1
			x.ScratchCodes[i] = x.ScratchCodes[l]
			x.ScratchCodes = x.ScratchCodes[0:l]
			return true
		}
	}
	return false
}

func (x *Totp) CheckCode(code int) bool {
	for i := 0; i < x.Window; i++ {
		if Compute(x.Secret, int64(x.Counter+i)) == code {
			x.Counter += i + 1
			return true
		}
	}
	x.Counter++
	return false
}

func (x *Totp) CheckTotpCode(ts, code int) bool {
	minT := ts - (x.Window / 2)
	maxT := ts + (x.Window / 2)
	for t := minT; t <= maxT; t++ {
		if Compute(x.Secret, int64(t)) == code {
			if x.DisallowReuse != nil {
				for _, timeCode := range x.DisallowReuse {
					if timeCode == t {
						return false
					}
				}
				x.DisallowReuse = append(x.DisallowReuse, t)
				sort.Ints(x.DisallowReuse)
				m := 0
				for x.DisallowReuse[m] < minT {
					m++
				}
				x.DisallowReuse = x.DisallowReuse[m:]
			}
			return true
		}
	}
	return false
}

func Compute(secret string, value int64) int {
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return -1
	}

	hash := hmac.New(sha1.New, key)
	err = binary.Write(hash, binary.BigEndian, value)
	if err != nil {
		return -1
	}
	h := hash.Sum(nil)

	offset := h[19] & 0x0f

	truncated := binary.BigEndian.Uint32(h[offset : offset+4])

	truncated &= 0x7fffffff
	code := truncated % 1000000

	return int(code)
}
