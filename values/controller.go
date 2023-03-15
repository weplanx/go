package values

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
)

type Controller struct {
	Service *Service
}

type SetDto struct {
	Data map[string]interface{} `json:"data,required" vd:"len($)>0 && range($,regexp('^[a-z_]+$',#k));msg:'Keys must be lowercase letters with underscores'"`
}

func (x *Controller) Set(_ context.Context, c *app.RequestContext) {
	var dto SetDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if err := x.Service.Set(dto.Data); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

type GetDto struct {
	Keys map[string]int64 `query:"keys" vd:"len($)==0 || range($,regexp('^[a-z_]+$',#k));msg:'Keys must be lowercase letters with underscores'"`
}

func (x *Controller) Get(_ context.Context, c *app.RequestContext) {
	var dto GetDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	data, err := x.Service.Get(dto.Keys)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, data)
}

type RemoveDto struct {
	Key string `path:"key,required" vd:"regexp('^[a-z_]+$');msg:'Keys must be lowercase letters with underscores'"`
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

	c.Status(http.StatusNoContent)
}
