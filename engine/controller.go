package engine

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/weplanx/go/route"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strconv"
)

type Controller struct {
	Engine *Engine
	*Service
}

func (x *Controller) DefaultRouters(r *gin.RouterGroup) {
	r.POST("/:model", route.Use(x.Actions))
	r.HEAD("/:model/_count", route.Use(x.Count))
	r.HEAD("/:model/_exists", route.Use(x.Exists))
	r.GET("/:model", route.Use(x.Get))
	r.GET("/:model/:id", route.Use(x.GetById))
	r.PATCH("/:model", route.Use(x.Patch))
	r.PATCH("/:model/:id", route.Use(x.PatchById))
	r.PUT("/:model/:id", route.Use(x.Put))
	r.DELETE("/:model/:id", route.Use(x.Delete))
}

func (x *Controller) NewContext(c *gin.Context) (ctx context.Context, err error) {
	var uri Uri
	if err = c.ShouldBindUri(&uri); err != nil {
		return
	}
	if model, exists := c.Get("model"); exists {
		uri.Model = model.(string)
	}
	var headers Headers
	if err = c.ShouldBindHeader(&headers); err != nil {
		return
	}
	ctx = context.WithValue(c.Request.Context(),
		"params", &Params{
			Uri:     &uri,
			Headers: &headers,
		},
	)
	return
}

func (x *Controller) Actions(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	params := x.Params(ctx)
	switch params.Action {
	case "create":
		// 创建文档
		var body M
		if err = c.ShouldBindJSON(&body); err != nil {
			return err
		}
		if len(body) == 0 {
			return BodyEmpty
		}
		result, err := x.InsertOne(ctx, body)
		if err != nil {
			return err
		}
		c.Set("status_code", http.StatusCreated)
		if err = x.Event(ctx, "create", result); err != nil {
			return err
		}
		return result
	case "bulk-create":
		// 批量创建文档
		var body []M
		if err = c.ShouldBindJSON(&body); err != nil {
			return err
		}
		if len(body) == 0 {
			return BodyEmpty
		}
		result, err := x.InsertMany(ctx, body)
		if err != nil {
			return err
		}
		c.Set("status_code", http.StatusCreated)
		if err = x.Event(ctx, "bulk-create", result); err != nil {
			return err
		}
		return result
	case "bulk-delete":
		// 批量删除文档
		var body M
		if err = c.ShouldBindJSON(&body); err != nil {
			return err
		}
		if len(body) == 0 {
			return BodyEmpty
		}
		var result interface{}
		if result, err = x.DeleteMany(ctx, body); err != nil {
			return err
		}
		if err = x.Event(ctx, "delete", result); err != nil {
			return err
		}
		return result
	case "sort":
		// 通用排序
		var body []primitive.ObjectID
		if err := c.ShouldBindJSON(&body); err != nil {
			return err
		}
		result, err := x.Service.Sort(ctx, body)
		if err != nil {
			return err
		}
		return result
	}
	return nil
}

// Count 获取集合文档总数
func (x *Controller) Count(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var query struct {
		Filter M `form:"filter"`
	}
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	var count int64
	if count, err = x.CountDocuments(ctx, query.Filter); err != nil {
		return err
	}
	c.Header("wpx-total", strconv.FormatInt(count, 10))
	return nil
}

// Exists 获取文档是否存在
func (x *Controller) Exists(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var query struct {
		Filter M `form:"filter" binding:"required,gt=0"`
	}
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	exists, err := x.ExistDocument(ctx, query.Filter)
	if err != nil {
		return err
	}
	c.Header("wpx-exists", strconv.FormatBool(exists))
	return nil
}

// Get 通过获取多个文档
func (x *Controller) Get(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	params := x.Params(ctx)
	var query struct {
		Filter M        `form:"filter"`
		Sort   []string `form:"sort" binding:"omitempty,gt=0,dive,sort"`
		Field  []string `form:"field" binding:"omitempty,gt=0,dive,key"`
	}
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	var data interface{}
	switch params.Type {
	case "find-one":
		if data, err = x.FindOne(ctx, query.Filter, query.Field); err != nil {
			return err
		}
		return data
	case "find-by-page":
		if data, err = x.FindByPage(ctx, query.Filter, query.Sort, query.Field); err != nil {
			return err
		}
		c.Header("wpx-total", strconv.FormatInt(params.Total, 10))
		return data
	}
	if data, err = x.Find(ctx, query.Filter, query.Sort, query.Field, nil); err != nil {
		return err
	}
	return data
}

// GetById 获取单个文档
func (x *Controller) GetById(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var query struct {
		Field []string `form:"field" binding:"omitempty,gt=0,dive,key"`
	}
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	var data interface{}
	if data, err = x.FindOneById(ctx, query.Field); err != nil {
		return err
	}
	return data
}

// Patch 局部更新文档
func (x *Controller) Patch(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var query struct {
		Filter  M            `form:"filter" binding:"required,gt=0"`
		Options QueryOptions `form:"options"`
	}
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	var body M
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if len(body) == 0 {
		return BodyEmpty
	}
	var result interface{}
	if result, err = x.UpdateMany(ctx, query.Filter, body, query.Options); err != nil {
		return err
	}
	if err = x.Event(ctx, "update", result); err != nil {
		return err
	}
	return result
}

// PatchById 指定 ID 局部更新文档
func (x *Controller) PatchById(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var query struct {
		Options QueryOptions `form:"options"`
	}
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	var body M
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if len(body) == 0 {
		return BodyEmpty
	}
	var result interface{}
	if result, err = x.UpdateOneById(ctx, body, query.Options); err != nil {
		return err
	}
	if err = x.Event(ctx, "update", result); err != nil {
		return err
	}
	return result
}

// Put 替换文档
func (x *Controller) Put(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var body M
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	if len(body) == 0 {
		return BodyEmpty
	}
	var result interface{}
	if result, err = x.ReplaceOneById(ctx, body); err != nil {
		return err
	}
	if err = x.Event(ctx, "update", result); err != nil {
		return err
	}
	return result
}

// Delete 删除文档
func (x *Controller) Delete(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var result interface{}
	if result, err = x.DeleteOneById(ctx); err != nil {
		return err
	}
	if err = x.Event(ctx, "delete", result); err != nil {
		return err
	}
	return result
}
