package support

import "testing"

func TestGeneratePlanx(t *testing.T) {
	if err := GeneratePlanx(db); err != nil {
		t.Error(err)
	}
}
