package cmd

import (
	"github.com/kainonly/go-bit/support/model"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
)

var Drive string
var DSN string

var Setup = &cobra.Command{
	Use:   "setup",
	Short: "初始化系统数据与模型，*已存在的数据同时会被清空",
	Run: func(cmd *cobra.Command, args []string) {
		var dialector gorm.Dialector
		switch Drive {
		case "mysql":
			dialector = mysql.Open(DSN)
			break
		case "postgres":
			dialector = postgres.Open(DSN)
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
		if db.Migrator().HasTable(&model.Datastore{}) {
			if err := db.Migrator().DropTable(&model.Datastore{}); err != nil {
				log.Fatalln(err)
			}
		}
		if err := db.AutoMigrate(&model.Datastore{}); err != nil {
			log.Fatalln(err)
		}
		data := []model.Datastore{
			{
				Key:  "resource",
				Type: "collection",
				Schema: model.Schema{
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
		if err := model.GenerateModel(db); err != nil {
			log.Fatalln(err)
		}
	},
}
