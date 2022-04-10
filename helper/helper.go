package helper

import (
	"errors"
	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"regexp"
)

// Uuid 生成 UUIDv4 字符串
func Uuid() string {
	return uuid.New().String()
}

var (
	PasswordParams = argon2id.DefaultParams
	NotMatch       = errors.New("the password does not match the hash value")
)

// PasswordHash  创建密码的散列
func PasswordHash(password string) (string, error) {
	return argon2id.CreateHash(password, PasswordParams)
}

// PasswordVerify 验证密码是否和散列值匹配
func PasswordVerify(password string, hash string) error {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return err
	}
	if !match {
		return NotMatch
	}
	return nil
}

// ExtendValidation 扩展验证
func ExtendValidation() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("objectId", func(fl validator.FieldLevel) bool {
			return primitive.IsValidObjectID(fl.Field().String())
		})
		v.RegisterValidation("key", func(fl validator.FieldLevel) bool {
			matched, _ := regexp.MatchString(`^[a-z_]+$`, fl.Field().String())
			return matched
		})
		v.RegisterValidation("sort", func(fl validator.FieldLevel) bool {
			matched, _ := regexp.MatchString(`^[a-z_]+\.(1|-1)$`, fl.Field().String())
			return matched
		})
	}
}
