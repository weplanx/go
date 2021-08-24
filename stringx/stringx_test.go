package stringx

import (
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandom(t *testing.T) {
	s := Random(8)
	t.Log(s)
	assert.Nil(t, validator.New().Var(s, "alpha,len=8"))
}

func TestUuid(t *testing.T) {
	assert.Nil(t, validator.New().Var(Uuid(), "uuid"))
}

func TestCamel(t *testing.T) {
	assert.Equal(t, Camel("my_lab"), "MyLab")
	assert.Equal(t, Camel("my-lab"), "MyLab")
	assert.Equal(t, Camel("my lab"), "MyLab")
}

func TestSnake(t *testing.T) {
	assert.Equal(t, Snake("MyLab"), "my_lab")
	assert.Equal(t, Snake("my-lab"), "my_lab")
	assert.Equal(t, Snake("my lab"), "my_lab")
}

func TestKebab(t *testing.T) {
	assert.Equal(t, Kebab("MyLab"), "my-lab")
	assert.Equal(t, Kebab("my_lab"), "my-lab")
	assert.Equal(t, Kebab("my lab"), "my-lab")
}

func TestLimit(t *testing.T) {
	assert.Equal(t,
		Limit("The quick brown fox jumps over the lazy dog", 20),
		"The quick brown fox...",
	)
}
