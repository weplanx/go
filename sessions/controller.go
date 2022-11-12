package sessions

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
)

type Controller struct {
	SessionsService *Service
}

// Lists 列出所有会话用户 ID
// @router /sessions [GET]
func (x *Controller) Lists(ctx context.Context, c *app.RequestContext) {
	data, err := x.SessionsService.Lists(ctx)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, data)
}

// Remove 移除会话
// @router /sessions/:uid [DELETE]
func (x *Controller) Remove(ctx context.Context, c *app.RequestContext) {
	if err := x.SessionsService.Remove(ctx, c.Param("uid")); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Clear 清除所有会话
// @router /sessions [DELETE]
func (x *Controller) Clear(ctx context.Context, c *app.RequestContext) {
	if err := x.SessionsService.Clear(ctx); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
