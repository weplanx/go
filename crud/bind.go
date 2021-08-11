package crud

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Bind 统一控制器函数返回
//	参数:
//	 handlerFn: func(c *gin.Context) interface{}
//	返回类型:
//	 (字符串) => 200 {"error":0,"msg":<字符串>}
//	 (错误) => 200 {"error":1,"msg":<err.Error()>}
//	 (默认) => 200 {"error":0,"data":<interface{}>}
//	自定义错误码: c.Set("code", 1000)
func Bind(handlerFn interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if fn, ok := handlerFn.(func(c *gin.Context) interface{}); ok {
			switch result := fn(c).(type) {
			case string:
				c.JSON(http.StatusOK, gin.H{
					"error": 0,
					"msg":   result,
				})
				break
			case error:
				code, exists := c.Get("code")
				if !exists {
					code = 1
				}
				c.JSON(http.StatusOK, gin.H{
					"error": code,
					"msg":   result.Error(),
				})
				break
			default:
				if result != nil {
					c.JSON(http.StatusOK, gin.H{
						"error": 0,
						"data":  result,
					})
				} else {
					c.Status(http.StatusNotFound)
				}
			}
		}
	}
}
