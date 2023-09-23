package help

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegValidate() {
	binding.MustRegValidateFunc("mongoId", func(args ...interface{}) error {
		if len(args) != 1 {
			return fmt.Errorf("the args must be one")
		}
		if _, e := primitive.ObjectIDFromHex(args[0].(string)); e != nil {
			return e
		}
		return nil
	})
}
