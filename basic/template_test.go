package basic

import "testing"

func TestGenerateModel(t *testing.T) {
	buf, err := GenerateModel(db)
	if err != nil {
		t.Error(err)
	}
	t.Log(buf.String())
}
