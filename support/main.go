package main

import (
	"github.com/kainonly/go-bit/support/cmd"
	"log"
)

func main() {
	cmd.Root.AddCommand(cmd.Setup)
	if err := cmd.Root.Execute(); err != nil {
		log.Fatalln(err)
	}
}
