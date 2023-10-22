package values

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
)

type Controller struct {
	Service *Service
}

type SetDto struct {
	Update map[string]interface{} `json:"update" vd:"gt=0,dive,keys,alphanum,endkeys,required"`
}

func (x *Controller) Set(_ context.Context, c *app.RequestContext) {
	var dto SetDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if err := x.Service.Set(dto.Update); err != nil {
		c.Error(err)
		return
	}

	c.Status(204)
}

type GetDto struct {
	Keys []string `query:"keys" vd:"omitempty,dive,alphanum"`
}

func (x *Controller) Get(_ context.Context, c *app.RequestContext) {
	var dto GetDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	data, err := x.Service.Get(dto.Keys...)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, data)
}

type RemoveDto struct {
	Key string `path:"key,required" vd:"alphanum"`
}

func (x *Controller) Remove(_ context.Context, c *app.RequestContext) {
	var dto RemoveDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if err := x.Service.Remove(dto.Key); err != nil {
		c.Error(err)
		return
	}

	c.Status(204)
}
