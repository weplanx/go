package sessions

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

type Controller struct {
	Service *Service
}

func (x *Controller) Lists(ctx context.Context, c *app.RequestContext) {
	c.JSON(200, x.Service.Lists(ctx))
}

type RemoveDto struct {
	Uid string `path:"uid,required" vd:"mongodb"`
}

func (x *Controller) Remove(ctx context.Context, c *app.RequestContext) {
	var dto RemoveDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}
	c.JSON(200, utils.H{
		"DeletedCount": x.Service.Remove(ctx, dto.Uid),
	})
}

func (x *Controller) Clear(ctx context.Context, c *app.RequestContext) {
	c.JSON(200, utils.H{
		"DeletedCount": x.Service.Clear(ctx),
	})
}
