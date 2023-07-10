package sessions

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
)

type Controller struct {
	Service *Service
}

func (x *Controller) Lists(ctx context.Context, c *app.RequestContext) {
	data, err := x.Service.Lists(ctx)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, data)
}

func (x *Controller) Remove(ctx context.Context, c *app.RequestContext) {
	if err := x.Service.Remove(ctx, c.Param("uid")); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (x *Controller) Clear(ctx context.Context, c *app.RequestContext) {
	if err := x.Service.Clear(ctx); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
