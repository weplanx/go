package kv

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
)

type Controller struct {
	KVService *Service
}

type SetDto struct {
	Data map[string]interface{} `json:"data,required" vd:"len($)>0 && range($,regexp('^[a-z_]+$',#k));msg:'key 必须是小写字母与下划线'"`
}

// Set 设置动态配置
// @router /values [PATCH]
func (x *Controller) Set(_ context.Context, c *app.RequestContext) {
	var dto SetDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if err := x.KVService.Set(dto.Data); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

type GetDto struct {
	// 动态配置键
	Keys []string `query:"keys" vd:"len($)==0 || range($,regexp('^[a-z_]+$',#k));msg:'key 必须是小写字母与下划线'"`
}

// Get 获取动态配置
// @router /values [GET]
func (x *Controller) Get(_ context.Context, c *app.RequestContext) {
	var dto GetDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	data, err := x.KVService.Get(dto.Keys...)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, data)
}

type RemoveDto struct {
	Key string `path:"key,required" vd:"regexp('^[a-z_]+$');msg:'key 必须是小写字母与下划线'"`
}

// Remove 移除动态配置
// @router /values/:id [DELETE]
func (x *Controller) Remove(_ context.Context, c *app.RequestContext) {
	var dto RemoveDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if err := x.KVService.Remove(dto.Key); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
