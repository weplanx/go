package route

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Use(fn func(c *gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch x := fn(c).(type) {
		case error:
			statusCode, exists := c.Get("status_code")
			if !exists {
				statusCode = http.StatusBadRequest
			}
			code, exists := c.Get("code")
			if !exists {
				code = "INVALID"
			}
			c.JSON(statusCode.(int), gin.H{
				"code":    code,
				"message": x.Error(),
			})
			break
		default:
			if x != nil {
				statusCode, exists := c.Get("status_code")
				if !exists {
					statusCode = http.StatusOK
				}
				c.JSON(statusCode.(int), x)
			} else {
				c.Status(http.StatusNoContent)
			}
		}
	}
}
