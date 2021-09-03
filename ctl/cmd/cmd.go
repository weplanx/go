package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	Setup.Flags().StringVarP(&Drive, "drive", "d", "mysql", "数据库驱动可以是 mysql 或 postgres")
	Setup.MarkFlagRequired("drive")
	Setup.Flags().StringVarP(&DSN, "dsn", "", "", "数据库连接 （必须）")
	Setup.MarkFlagRequired("dsn")
}

var Root = &cobra.Command{
	Use: "bit",
	Long: `一个提高 golang web 开发效率的工具
	- 项目 https://github.com/kainonly/go-bit
	- 文档 https://www.yuque.com/kainonly/go-bit`,
}
