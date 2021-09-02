package main

import (
	"github.com/kainonly/go-bit/support/cmd"
	"log"
)

func main() {
	if err := cmd.Init().Execute(); err != nil {
		log.Fatalln(err)
	}
}
