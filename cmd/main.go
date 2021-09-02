package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "bit",
	Long: `A tool to improve the efficiency of golang web development
	- Github https://github.com/kainonly/go-bit
	- Document https://www.yuque.com/kainonly/go-bit`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
