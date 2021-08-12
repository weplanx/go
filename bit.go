package bit

import (
	"errors"
	"github.com/kainonly/go-bit/authx"
	"github.com/kainonly/go-bit/cipher"
	"github.com/kainonly/go-bit/cookie"
	"github.com/kainonly/go-bit/crud"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
)

type Config map[string]interface{}

// LoadConfiguration 初始化应用配置
func LoadConfiguration() (config Config, err error) {
	if _, err = os.Stat("./config.yml"); os.IsNotExist(err) {
		err = errors.New("the configuration file path [./config.yml] does not exist")
		return
	}
	var buf []byte
	buf, err = ioutil.ReadFile("./config.yml")
	if err != nil {
		return
	}
	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return
	}
	return
}

// InitializeCrud 初始化 CRUD 工具
func InitializeCrud(db *gorm.DB) *crud.Crud {
	return &crud.Crud{Db: db}
}

// InitializeCipher 初始化数据加密
func InitializeCipher(config Config) (*cipher.Cipher, error) {
	return cipher.New(config["cipher"].(string))
}

// InitializeCookie 创建 Cookie 工具
func InitializeCookie(config Config) (x *cookie.Cookie, err error) {
	var option cookie.Option
	if err = mapstructure.Decode(config["cookie"], &option); err != nil {
		return
	}
	return cookie.New(option), nil
}

func InitializeAuthx(config Config) (x *authx.Authx, err error) {
	var option map[string]authx.Option
	if err = mapstructure.Decode(config["auth"], &option); err != nil {
		return
	}
	return authx.New(option), nil
}
