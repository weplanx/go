package password

import (
	"errors"
	"github.com/alexedwards/argon2id"
)

var (
	Params   = argon2id.DefaultParams
	NotMatch = errors.New("the password does not match the hash value")
)

func Create(password string) (string, error) {
	return argon2id.CreateHash(password, Params)
}

func Verify(password string, hash string) error {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return err
	}
	if !match {
		return NotMatch
	}
	return nil
}
