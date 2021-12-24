package api

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/wire"
	"regexp"
)

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)

type Pagination struct {
	Index int64 `header:"x-page" binding:"omitempty,gt=0,number"`
	Size  int64 `header:"x-page-size" binding:"omitempty,oneof=10 20 50 100"`
}

func RegisterValidation() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("sort", func(fl validator.FieldLevel) bool {
			matched, _ := regexp.MatchString(`^[a-z_]+\.(1|-1)$`, fl.Field().String())
			return matched
		})
	}
}
