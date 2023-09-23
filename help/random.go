package help

import "math/rand"

func Random(n int, charset ...string) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if len(charset) != 0 {
		letters = charset[0]
	}

	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func RandomNumber(n int) string {
	return Random(n, "0123456789")
}

func RandomLowercase(n int) string {
	return Random(n, "abcdefghijklmnopqrstuvwxyz")
}

func RandomUppercase(n int) string {
	return Random(n, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

func RandomAlphabet(n int) string {
	return Random(n, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
}
