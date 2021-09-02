package cmd

import (
	"github.com/kainonly/go-bit/support/core"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
)

var Root = &cobra.Command{
	Use: "bit",
	Long: `一个提高 golang web 开发效率的工具
	- 项目 https://github.com/kainonly/go-bit
	- 文档 https://www.yuque.com/kainonly/go-bit`,
}

var Setup = &cobra.Command{
	Use:              "setup [type] [dsn]",
	Short:            "初始化系统数据与模型，*已存在的数据同时会被清空",
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		var dialector gorm.Dialector
		switch args[0] {
		case "mysql":
			dialector = mysql.Open(args[1])
			break
		case "postgres":
			dialector = postgres.Open(args[1])
			break
		}
		db, err := gorm.Open(dialector, &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
		})
		if err != nil {
			log.Fatalln(err)
		}
		if db.Migrator().HasTable(&core.Datastore{}) {
			if err := db.Migrator().DropTable(&core.Datastore{}); err != nil {
				log.Fatalln(err)
			}
		}
		if err := db.AutoMigrate(&core.Datastore{}); err != nil {
			log.Fatalln(err)
		}
		data := []core.Datastore{
			{
				Key:  "resource",
				Type: "collection",
				Schema: core.Schema{
					{
						Key:     "parent",
						Label:   "父级",
						Type:    "varchar",
						Default: "root",
						Require: true,
						Length:  50,
					},
					{
						Key:     "path",
						Label:   "路径",
						Type:    "varchar",
						Require: true,
						Length:  50,
					},
					{
						Key:     "name",
						Label:   "资源名称",
						Type:    "varchar",
						Require: true,
						Length:  20,
					},
					{
						Key:     "nav",
						Label:   "是否为导航",
						Type:    "boolean",
						Default: "true",
						Require: true,
					},
					{
						Key:     "router",
						Label:   "是否为路由页面",
						Type:    "boolean",
						Default: "true",
						Require: true,
					},
					{
						Key:    "icon",
						Label:  "字体图标",
						Type:   "varchar",
						Length: 50,
					},
					{
						Key:     "sort",
						Label:   "排序",
						Type:    "smallint",
						Default: "0",
						Hide:    true,
					},
				},
			},
		}
		if err := db.Create(&data).Error; err != nil {
			log.Fatalln(err)
		}
		if err := core.GenerateModel(db); err != nil {
			log.Fatalln(err)
		}
	},
}

func Init() *cobra.Command {
	Root.AddCommand(Setup)
	return Root
}
