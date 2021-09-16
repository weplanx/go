package mvc

import (
	"github.com/gin-gonic/gin"
	"github.com/kainonly/go-bit/crud"
	"net/http"
)

// Returns Unified controller function
//	 handlerFn: func(c *gin.Context) interface{}
//	return types:
//	 (string) => 200 {"error":0,"msg":<string>}
//	 (error) => 200 {"error":1,"msg":<err.Error()>}, custom error code: c.Set("code", 1000)
//	 (interface) => 200 {"error":0,"data":<interface{}>}
func Returns(handlerFn interface{}) gin.HandlerFunc {
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

func Crud(r *gin.RouterGroup, i crud.Operations) *gin.RouterGroup {
	r.POST("r/first", Returns(i.First))
	r.POST("r/find", Returns(i.Find))
	r.POST("r/page", Returns(i.Page))
	r.POST("w/create", Returns(i.Create))
	r.POST("w/update", Returns(i.Update))
	r.POST("w/delete", Returns(i.Delete))
	return r
}
