package passlib

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

var (
	DEFAULT_MEMORY_COST uint32 = 65536
	DEFAULT_TIME_COST   uint32 = 4
	DEFAULT_THREADS     uint8  = 1
)

var (
	ErrInvalidHash         = errors.New("unable to parse the current hash value")
	ErrIncompatibleVariant = errors.New("hash variants are not compatible")
	ErrIncompatibleVersion = errors.New("hash version are not support")
	ErrNotMatch            = errors.New("password does not match hash")
)

func Hash(password string) (hash string, err error) {
	salt := make([]byte, 16)
	if _, err = rand.Read(salt); err != nil {
		return
	}
	key := argon2.IDKey([]byte(password), salt,
		DEFAULT_TIME_COST, DEFAULT_MEMORY_COST, DEFAULT_THREADS, 32)

	return fmt.Sprintf(`$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s`,
		argon2.Version, DEFAULT_MEMORY_COST, DEFAULT_TIME_COST, DEFAULT_THREADS,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func Verify(password string, hash string) (err error) {
	options := strings.Split(hash, "$")
	fmt.Println(options)
	if len(options) != 6 {
		return ErrInvalidHash
	}
	if options[1] != "argon2id" {
		return ErrIncompatibleVariant
	}
	var version int
	if _, err = fmt.Sscanf(options[2], "v=%d", &version); err != nil {
		return ErrIncompatibleVersion
	}
	if version != argon2.Version {
		return ErrIncompatibleVersion
	}
	var memory uint32
	var time uint32
	var threads uint8
	if _, err = fmt.Sscanf(options[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return ErrInvalidHash
	}
	var salt []byte
	if salt, err = base64.RawStdEncoding.Strict().DecodeString(options[4]); err != nil {
		return
	}
	var key []byte
	if key, err = base64.RawStdEncoding.Strict().DecodeString(options[5]); err != nil {
		return
	}
	otherKey := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(key)))
	if subtle.ConstantTimeEq(int32(len(key)), int32(len(otherKey))) == 0 {
		return ErrNotMatch
	}
	if subtle.ConstantTimeCompare(key, otherKey) == 1 {
		return
	}
	return ErrNotMatch
}
