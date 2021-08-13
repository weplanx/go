package hash

import (
	"errors"
	"github.com/alexedwards/argon2id"
)

var Invalid = errors.New("invalid password verification")

// Make 创建密码
func Make(text string, options ...*argon2id.Params) (string, error) {
	option := argon2id.DefaultParams
	if len(options) != 0 {
		option = options[0]
	}
	return argon2id.CreateHash(text, option)
}

// Verify 验证密码
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
