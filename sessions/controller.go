package sessions

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
)

type Controller struct {
	Service *Service
}

// Lists all session user IDs
// @router /sessions [GET]
func (x *Controller) Lists(ctx context.Context, c *app.RequestContext) {
	data, err := x.Service.Lists(ctx)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, data)
}

// Remove Session
// @router /sessions/:uid [DELETE]
func (x *Controller) Remove(ctx context.Context, c *app.RequestContext) {
	if err := x.Service.Remove(ctx, c.Param("uid")); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Clear sessions
// @router /sessions [DELETE]
func (x *Controller) Clear(ctx context.Context, c *app.RequestContext) {
	if err := x.Service.Clear(ctx); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
