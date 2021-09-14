package support

import "testing"

func TestGenerateResources(t *testing.T) {
	if err := GenerateResources(db); err != nil {
		t.Error(err)
	}
}
