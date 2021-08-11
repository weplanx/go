package str

import (
	"github.com/google/uuid"
	"github.com/huandu/xstrings"
	"math/rand"
	"time"
)

// Random 生成随机数
func Random(length int, letterRunes ...rune) string {
	b := make([]rune, length)
	if len(letterRunes) == 0 {
		letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	}
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// Uuid 生成 UUID
func Uuid() uuid.UUID {
	return uuid.New()
}

// Camel 字符串风格 CamelCase
func Camel(str string) string {
	return xstrings.ToCamelCase(str)
}

// Snake 字符串风格 snake_case
func Snake(str string) string {
	return xstrings.ToSnakeCase(str)
}

// Kebab 字符串风格 kebab-case
func Kebab(str string) string {
	return xstrings.ToKebabCase(str)
}

// Limit 省略
func Limit(str string, length int) string {
	return str[:length-1] + "..."
}
