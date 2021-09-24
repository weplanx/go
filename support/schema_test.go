package support

import "testing"

func TestGenerateSchema(t *testing.T) {
	if err := GenerateSchema(db); err != nil {
		t.Error(err)
	}
}
