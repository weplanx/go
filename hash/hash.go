package hash

import (
	"errors"
	"github.com/alexedwards/argon2id"
)

var Invalid = errors.New("invalid password verification")

// Make create password
func Make(text string) (string, error) {
	return argon2id.CreateHash(text, argon2id.DefaultParams)
}

// Verify verify password
func Verify(text string, hashText string) error {
	result, err := argon2id.ComparePasswordAndHash(text, hashText)
	if err != nil {
		return err
	}
	if !result {
		return Invalid
	}
	return nil
}
