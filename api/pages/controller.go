package pages

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type Controller struct {
	PagesService *Service
}

// GetOne 获取页面数据
// @router /:id
func (x *Controller) GetOne(ctx context.Context, c *app.RequestContext) {
	var dto struct {
		// 页面 ID
		Id string `path:"id,required" vd:"mongoId($);msg:'页面 ID 不规范'"`
	}
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	id, _ := primitive.ObjectIDFromHex(dto.Id)
	page, err := x.PagesService.FindOneById(ctx, id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, page)
}
