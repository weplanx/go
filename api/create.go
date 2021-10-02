package api

import "github.com/gin-gonic/gin"

type CreateBody struct {
	Data map[string]interface{} `json:"data" binding:"required"`
}

// Create resources
func (x *API) Create(c *gin.Context) interface{} {
	uri, err := x.getUri(c)
	if err != nil {
		return err
	}
	var body CreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	// TODO: Load schema cache
	if err := x.Db.WithContext(c).
		Table(uri.Model).
		Create(body.Data).Error; err != nil {
		return err
	}
	return "ok"
}
