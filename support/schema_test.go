package support

import "testing"

func TestGenerateSchema(t *testing.T) {
	if err := GenerateSchema(db); err != nil {
		t.Error(err)
	}
}

func TestGenerateModels(t *testing.T) {
	buf, err := GenerateModels(db)
	if err != nil {
		t.Error(err)
	}
	t.Log(buf.String())
}
