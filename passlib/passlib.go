package passlib

import (
	"github.com/alexedwards/argon2id"
)

func Hash(password string) (string, error) {
	return argon2id.CreateHash(password, &argon2id.Params{
		Memory:      65536,
		Iterations:  4,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})
}

func Verify(password string, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
