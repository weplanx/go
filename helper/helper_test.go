package helper

import (
	"github.com/google/uuid"
	"testing"
)

func TestUuid(t *testing.T) {
	u, err := uuid.Parse(Uuid())
	if err != nil {
		t.Error(err)
	}
	t.Log(u)
}
