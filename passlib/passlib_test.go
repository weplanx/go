package passlib_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/go/passlib"
	"testing"
)

func TestHashAndVerify(t *testing.T) {
	hash, err := passlib.Hash("pass@VAN1234")
	assert.NoError(t, err)
	err = passlib.Verify("pass@VAN1234", hash)
	assert.NoError(t, err)
	err = passlib.Verify("pass@VAN1235", hash)
	assert.ErrorIs(t, err, passlib.ErrNotMatch)
}

const PASS1 = `$argon2i$v=19$m=65536,t=4,p=1$NPCjKIcoU2z6rg6p8glOfg$jrbRcvsTq/ITJP414/xhNNwOtVeHYa478hPn8M6uJLA`
const PASS2 = `$argon2id$v=x$m=65536,t=4,p=1$NPCjKIcoU2z6rg6p8glOfg$jrbRcvsTq/ITJP414/xhNNwOtVeHYa478hPn8M6uJLA`
const PASS3 = `$argon2id$v=18$m=65536,t=4,p=1$NPCjKIcoU2z6rg6p8glOfg$jrbRcvsTq/ITJP414/xhNNwOtVeHYa478hPn8M6uJLA`
const PASS4 = `$argon2id$v=19$xcxcsdsdwe$NPCjKIcoU2z6rg6p8glOfg$jrbRcvsTq/ITJP414/xhNNwOtVeHYa478hPn8M6uJLA`
const PASS5 = `$argon2id$v=19$m=65536,t=4,p=1$()$jrbRcvsTq/ITJP414/xhNNwOtVeHYa478hPn8M6uJLA`
const PASS6 = `$argon2id$v=19$m=65536,t=4,p=1$NPCjKIcoU2z6rg6p8glOfg$()`

func TestVerifyErrors(t *testing.T) {
	var err error
	err = passlib.Verify("pass@VAN1234", "asdaqweqwexcxzcqweqw")
	assert.ErrorIs(t, err, passlib.ErrInvalidHash)
	err = passlib.Verify("pass@VAN1234", PASS1)
	assert.ErrorIs(t, err, passlib.ErrIncompatibleVariant)
	err = passlib.Verify("pass@VAN1234", PASS2)
	assert.ErrorIs(t, err, passlib.ErrIncompatibleVersion)
	err = passlib.Verify("pass@VAN1234", PASS3)
	assert.ErrorIs(t, err, passlib.ErrIncompatibleVersion)
	err = passlib.Verify("pass@VAN1234", PASS4)
	assert.ErrorIs(t, err, passlib.ErrInvalidHash)
	err = passlib.Verify("pass@VAN1234", PASS5)
	assert.Error(t, err)
	err = passlib.Verify("pass@VAN1234", PASS6)
	assert.Error(t, err)
	t.Log(err)
}
