package help

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
)

func Ptr[T any](i T) *T {
	return &i
}

func IsEmpty(i any) bool {
	if i == nil || i == "" || i == false {
		return true
	}

	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.Len() == 0 || v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr, reflect.Func, reflect.Chan:
		return v.IsNil()
	}

	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

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
