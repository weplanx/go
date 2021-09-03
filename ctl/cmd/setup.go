package cmd

import (
	"github.com/spf13/cobra"
)

var DSN string

var Setup = &cobra.Command{
	Use:   "setup",
	Short: "初始化系统数据与模型，*已存在的数据同时会被清空",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
