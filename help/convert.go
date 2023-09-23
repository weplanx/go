package help

import (
	"math/rand"
)

func Reverse[T any](v []T) {
	for n, m := 0, len(v)-1; n < len(v)/2; n, m = n+1, m-1 {
		v[n], v[m] = v[m], v[n]
	}
}

func Shuffle[T any](v []T) {
	m := 0
	for n := len(v) - 1; n > 0; n-- {
		m = rand.Intn(n + 1)
		if n != m {
			v[n], v[m] = v[m], v[n]
		}
	}
}

func ReverseString(v string) string {
	runes := []rune(v)
	for n, m := 0, len(runes)-1; n < len(runes)/2; n, m = n+1, m-1 {
		runes[n], runes[m] = runes[m], runes[n]
	}
	return string(runes)
}

func ShuffleString(v string) string {
	runes, m := []rune(v), 0
	for n := len(runes) - 1; n > 0; n-- {
		m = rand.Intn(n + 1)
		if n != m {
			runes[n], runes[m] = runes[m], runes[n]
		}
	}
	return string(runes)
}
