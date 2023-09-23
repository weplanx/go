package totp_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/totp"
	"testing"
	"time"
)

type Value1 struct {
	code   string
	result bool
}

func TestAuthenticate(t *testing.T) {
	x := &totp.Totp{
		Secret:       "2SH3V3GDW7ZNMGYE",
		Window:       3,
		Counter:      1,
		ScratchCodes: []int{11112222, 22223333},
	}
	values := []Value1{
		{"foobar", false},
		{"1fooba", false},
		{"1111111", false},
		{"293240", true},
		{"293240", false},
		{"33334444", false},
		{"11112222", true},
		{"11112222", false},
	}
	for _, v := range values {
		r, _ := x.Authenticate(v.code)
		assert.Equal(t, v.result, r)
	}
	x.Counter = 0
	ts := time.Now().UTC().Unix() / 30
	code := fmt.Sprintf("%06d", totp.Compute(x.Secret, ts))
	values = []Value1{
		{code + "1", false},
		{code, true},
	}
	for _, v := range values {
		r, _ := x.Authenticate(v.code)
		assert.Equal(t, v.result, r)
	}
}

type Value2 struct {
	code   int
	ts     int
	result bool
}

type Value3 struct {
	code       int
	ts         int
	result     bool
	disallowed []int
}

func TestTotpCode(t *testing.T) {
	var x totp.Totp
	x.Secret = "2SH3V3GDW7ZNMGYE"
	x.Window = 5
	values := []Value2{
		{50548, 9997, false},
		{50548, 9998, true},
		{50548, 9999, true},
		{50548, 10000, true},
		{50548, 10001, true},
		{50548, 10002, true},
		{50548, 10003, false},
	}

	for _, v := range values {
		r := x.CheckTotpCode(v.ts, v.code)
		assert.Equal(t, v.result, r)
	}

	x.DisallowReuse = make([]int, 0)
	var noreuses = []Value3{
		{50548 /* 10000 */, 9997, false, []int{}},
		{50548 /* 10000 */, 9998, true, []int{10000}},
		{50548 /* 10000 */, 9999, false, []int{10000}},
		{478726 /* 10001 */, 10001, true, []int{10000, 10001}},
		{646986 /* 10002 */, 10002, true, []int{10000, 10001, 10002}},
		{842639 /* 10003 */, 10003, true, []int{10001, 10002, 10003}},
	}

	for _, v := range noreuses {
		r := x.CheckTotpCode(v.ts, v.code)
		assert.Equal(t, v.result, r)
		assert.Equal(t, len(x.DisallowReuse), len(v.disallowed))
		same := true
		for i := range v.disallowed {
			if v.disallowed[i] != x.DisallowReuse[i] {
				same = false
			}
		}
		assert.True(t, same)
	}
}
