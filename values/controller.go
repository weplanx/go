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
	Update map[string]interface{} `json:"update,required" vd:"len($)>0&&range($,regexp('^[a-zA-Z]+$',#k));msg:'the key must be alphabet'"`
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

	c.Status(http.StatusNoContent)
}

type GetDto struct {
	Keys []string `query:"keys" vd:"range($,regexp('^[a-zA-Z]+$',#v));msg:'the key must be alphabet'"`
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

	c.JSON(http.StatusOK, data)
}

type RemoveDto struct {
	Key string `path:"key,required" vd:"regexp('^[a-zA-Z]+$');msg:'the key must be alphabet'"`
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
