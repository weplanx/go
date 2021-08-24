package str

import (
	"github.com/google/uuid"
	"github.com/huandu/xstrings"
	"math/rand"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Random generate random string
func Random(n int) string {
	b := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

// Uuid UUID
func Uuid() uuid.UUID {
	return uuid.New()
}

// Camel CamelCase
func Camel(str string) string {
	return xstrings.ToCamelCase(str)
}

// Snake snake_case
func Snake(str string) string {
	return xstrings.ToSnakeCase(str)
}

// Kebab kebab-case
func Kebab(str string) string {
	return xstrings.ToKebabCase(str)
}

// Limit 省略
func Limit(str string, length int) string {
	return str[:length-1] + "..."
}
