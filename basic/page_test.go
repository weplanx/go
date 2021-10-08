package basic

import (
	"context"
	"testing"
)

func TestGeneratePage(t *testing.T) {
	ctx := context.Background()
	if err := GeneratePage(ctx, db); err != nil {
		t.Error(err)
	}
}
