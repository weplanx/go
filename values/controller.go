package values

import "github.com/gin-gonic/gin"

type Controller struct {
	Service *Service
}

// Set 设置动态配置
func (x *Controller) Set(c *gin.Context) interface{} {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if err := x.Service.Set(body); err != nil {
		return err
	}
	return nil
}

// Get 获取动态配置
func (x *Controller) Get(c *gin.Context) interface{} {
	var query struct {
		Keys []string `form:"keys" binding:"required,dive,gt=0"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		return err
	}
	data, err := x.Service.Get(query.Keys)
	if err != nil {
		return err
	}
	return data
}
