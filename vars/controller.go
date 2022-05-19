package vars

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	Vars *Service
}

// Get 获取指定变量
func (x *Controller) Get(c *gin.Context) interface{} {
	var query struct {
		Keys []string `form:"keys" binding:"required"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		return err
	}
	ctx := c.Request.Context()
	values, err := x.Vars.toMap(ctx, query.Keys)
	if err != nil {
		return err
	}
	for k, v := range values {
		if SecretText(k) {
			if v == "" || v == nil {
				values[k] = "-"
			} else {
				values[k] = "***"
			}
		}
	}
	return values
}

// Set 设置变量
func (x *Controller) Set(c *gin.Context) interface{} {
	var uri struct {
		Key string `uri:"key" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		return err
	}
	var body struct {
		Value interface{} `json:"value"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	if err := x.Vars.Set(ctx, uri.Key, body.Value); err != nil {
		return err
	}
	return nil
}

func (x *Controller) Options(c *gin.Context) interface{} {
	var query struct {
		Type string `form:"type" binding:"required"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		return err
	}
	ctx := c.Request.Context()
	switch query.Type {
	case "upload":
		platform, err := x.Vars.GetCloudPlatform(ctx)
		if err != nil {
			return err
		}
		switch platform {
		case "tencent":
			bucket, err := x.Vars.GetTencentCosBucket(ctx)
			if err != nil {
				return err
			}
			region, err := x.Vars.GetTencentCosRegion(ctx)
			if err != nil {
				return err
			}
			limit, err := x.Vars.GetTencentCosLimit(ctx)
			if err != nil {
				return err
			}
			return gin.H{
				"type":  "cos",
				"url":   fmt.Sprintf(`https://%s.cos.%s.myqcloud.com`, bucket, region),
				"limit": limit,
			}
		}
	case "office":
		platform, err := x.Vars.GetOfficePlatform(ctx)
		if err != nil {
			return err
		}
		redirect, err := x.Vars.GetRedirectUrl(ctx)
		if err != nil {
			return err
		}
		switch platform {
		case "feishu":
			id, err := x.Vars.GetFeishuAppId(ctx)
			if err != nil {
				return err
			}
			return gin.H{
				"url":      "https://open.feishu.cn/open-apis/authen/v1/index",
				"redirect": redirect,
				"app_id":   id,
			}
		}
	}
	return nil
}
