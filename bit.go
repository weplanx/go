package bit

import (
	"github.com/gin-gonic/gin"
	"github.com/kainonly/go-bit/cipher"
	"github.com/kainonly/go-bit/cookie"
	"github.com/kainonly/go-bit/crud"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
	"net/http"
)

type Config map[string]interface{}

type Bit struct {
	Config Config
	Db     *gorm.DB
}

// Initialize 初始化辅助工具
func Initialize(db *gorm.DB, config Config) *Bit {
	return &Bit{config, db}
}

// Crud 创建控制器通用资源操作
//	参数:
//	 model: 模型名称
//	 options: 配置
func (x *Bit) Crud(model interface{}, options ...crud.Option) *crud.Crud {
	c := &crud.Crud{
		Db:    x.Db,
		Model: model,
	}
	for _, apply := range options {
		apply(c)
	}
	return c
}

// Cookie 创建 Cookie 工具
func (x *Bit) Cookie(ctx *gin.Context) (c *cookie.Cookie, err error) {
	var option cookie.Option
	if err = mapstructure.Decode(x.Config["cookie"], &option); err != nil {
		return
	}
	var samesite http.SameSite
	switch option.SameSite {
	case "lax":
		samesite = http.SameSiteLaxMode
		break
	case "strict":
		samesite = http.SameSiteStrictMode
		break
	case "none":
		samesite = http.SameSiteNoneMode
		break
	default:
		samesite = http.SameSiteDefaultMode
	}
	c = &cookie.Cookie{
		Option:       option,
		Ctx:          ctx,
		HttpSameSite: samesite,
	}
	return
}

// InitializeCipher 初始化数据加密工具
func InitializeCipher(config Config) (*cipher.Cipher, error) {
	return cipher.Make(config["cipher"].(string))
}
