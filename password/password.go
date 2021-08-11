package password

import (
	"github.com/alexedwards/argon2id"
)

// Make 创建密码
func Make(text string, options ...*argon2id.Params) (string, error) {
	option := argon2id.DefaultParams
	if len(options) != 0 {
		option = options[0]
	}
	return argon2id.CreateHash(text, option)
}

// Verify 验证密码
func Verify(text string, hashText string) (bool, error) {
	return argon2id.ComparePasswordAndHash(text, hashText)
}
