package bit

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Bind(handlerFn interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if fn, ok := handlerFn.(func(c *gin.Context) interface{}); ok {
			switch result := fn(c).(type) {
			case string:
				c.JSON(http.StatusOK, gin.H{
					"code": 0,
					"msg":  result,
				})
				break
			case error:
				code, exists := c.Get("code")
				if !exists {
					code = 1
				}
				c.JSON(http.StatusOK, gin.H{
					"code": code,
					"msg":  result.Error(),
				})
				break
			default:
				if result != nil {
					c.JSON(http.StatusOK, gin.H{
						"code": 0,
						"data": result,
					})
				} else {
					c.Status(http.StatusNotFound)
				}
			}
		}
	}
}
