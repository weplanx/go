package basic

import (
	"context"
	"testing"
)

func TestGenerateSchema(t *testing.T) {
	ctx := context.Background()
	if err := GenerateSchema(ctx, db); err != nil {
		t.Error(err)
	}
}
