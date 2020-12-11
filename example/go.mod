module example

go 1.15

require (
	github.com/json-iterator/go v1.1.10
	github.com/kainonly/gin-curd v0.0.0-20201208112541-54f82c6f0375
	gorm.io/driver/mysql v1.0.3
	gorm.io/gorm v1.20.9-0.20201207023106-e1952924e2a8
)

replace github.com/kainonly/gin-curd v0.0.0-20201208112541-54f82c6f0375 => ../
