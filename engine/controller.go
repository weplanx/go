package engine

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Controller struct {
	Engine  *Engine
	Service *Service
}

func (x *Controller) NewContext(c *gin.Context) (ctx context.Context, err error) {
	var params Params
	if err = c.ShouldBindUri(&params); err != nil {
		return
	}
	if model, exists := c.Get("model"); exists {
		params.Model = model.(string)
	}
	ctx = context.WithValue(c.Request.Context(),
		"params", &params,
	)
	return
}

type CreateBody struct {
	Docs   []M `json:"docs" binding:"required,dive,gt=0"`
	Format M   `json:"format" binding:"omitempty,dive,gt=0"`
}

// Create 创建文档
func (x *Controller) Create(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var body CreateBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	result, err := x.Service.InsertMany(ctx, body.Docs, body.Format)
	if err != nil {
		return err
	}
	c.Set("status_code", http.StatusCreated)
	if err = x.Service.Event(ctx, result); err != nil {
		return err
	}
	return result
}

type FindQuery struct {
	Type   string   `form:"type" binding:"omitempty"`
	Id     []string `form:"id" binding:"omitempty,excluded_with=Filter,dive,objectId"`
	Filter M        `form:"filter" binding:"omitempty,excluded_with=Id"`
	Order  []string `form:"order" binding:"omitempty,dive,gt=0,order"`
	Field  []string `form:"field" binding:"omitempty,dive,gt=0"`
	Limit  int64    `form:"limit" binding:"omitempty,dive,gt=0,lt=10000"`
	Skip   int64    `form:"skip" binding:"omitempty,dive,gte=0"`
}

// Find 通过获取多个文档
func (x *Controller) Find(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var query FindQuery
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	var data interface{}
	switch query.Type {
	case "single":
		if data, err = x.Service.FindOne(ctx,
			query.Filter, query.Field,
		); err != nil {
			return err
		}
		return data
	case "page":
		var p Pagination
		if err = c.ShouldBindHeader(&p); err != nil {
			return err
		}
		ctx = context.WithValue(ctx,
			"page", &p,
		)
		if data, err = x.Service.FindByPage(ctx,
			query.Filter, query.Order, query.Field); err != nil {
			return err
		}
		c.Header("x-page-total", strconv.FormatInt(p.Total, 10))
		return data
	}
	if len(query.Id) != 0 {
		if data, err = x.Service.FindById(ctx,
			query.Id, query.Order, query.Field); err != nil {
			return err
		}
	} else {
		if data, err = x.Service.Find(ctx,
			query.Filter, query.Order, query.Field, query.Limit, query.Skip); err != nil {
			return err
		}
	}
	return data
}

type FindOneByIdQuery struct {
	Field []string `form:"field"`
}

// FindOneById 通过 ID 获取单个文档
func (x *Controller) FindOneById(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var query FindOneByIdQuery
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	data, err := x.Service.FindOneById(ctx, query.Field)
	if err != nil {
		return err
	}
	return data
}

type UpdateQuery struct {
	Id     []string `form:"id" binding:"required_without=Filter,dive,objectId"`
	Filter M        `form:"filter" binding:"required_without=Id,excluded_with=Id"`
}

type UpdateBody struct {
	Update M `json:"update" binding:"required"`
	Format M `json:"format" binding:"omitempty,dive,gt=0"`
}

// Update 更新文档
func (x *Controller) Update(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var query UpdateQuery
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	var body UpdateBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	var result interface{}
	if len(query.Id) != 0 {
		if result, err = x.Service.UpdateManyById(ctx,
			query.Id, body.Update, body.Format,
		); err != nil {
			return err
		}
	} else {
		if result, err = x.Service.UpdateMany(ctx,
			query.Filter, body.Update, body.Format,
		); err != nil {
			return err
		}
	}
	if err = x.Service.Event(ctx, result); err != nil {
		return err
	}
	return result
}

// UpdateOne 更新单个文档
func (x *Controller) UpdateOne(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var body UpdateBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	result, err := x.Service.UpdateOne(ctx, body.Update, body.Format)
	if err != nil {
		return err
	}
	if err = x.Service.Event(ctx, result); err != nil {
		return err
	}
	return result
}

type ReplaceOneBody struct {
	Doc    M `json:"doc" binding:"required"`
	Format M `json:"format" binding:"omitempty,dive,gt=0"`
}

// ReplaceOne 替换文档
func (x *Controller) ReplaceOne(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	var body ReplaceOneBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	result, err := x.Service.ReplaceOne(ctx, body.Doc, body.Format)
	if err != nil {
		return err
	}
	if err = x.Service.Event(ctx, result); err != nil {
		return err
	}
	return result
}

// DeleteOne 删除文档
func (x *Controller) DeleteOne(c *gin.Context) interface{} {
	ctx, err := x.NewContext(c)
	if err != nil {
		return err
	}
	result, err := x.Service.DeleteOne(ctx)
	if err != nil {
		return err
	}
	if err = x.Service.Event(ctx, result); err != nil {
		return err
	}
	return result
}