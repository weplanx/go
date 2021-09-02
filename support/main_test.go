package main

import (
	"github.com/kainonly/go-bit/support/cmd"
	"os"
	"testing"
)

func TestSetup(t *testing.T) {
	cmd.Setup.SetArgs([]string{"postgres", os.Getenv("DB_DSN")})
	if err := cmd.Setup.Execute(); err != nil {
		t.Error(err)
	}
}
