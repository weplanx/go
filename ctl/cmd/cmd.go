package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	Setup.Flags().StringVarP(&DSN, "dsn", "", "", "postgres 数据库连接")
	Setup.MarkFlagRequired("dsn")
}

var Root = &cobra.Command{
	Use: "bit",
	Long: `一个提高 golang web 开发效率的工具
	- 项目 https://github.com/kainonly/go-bit
	- 文档 https://www.yuque.com/kainonly/go-bit`,
}
