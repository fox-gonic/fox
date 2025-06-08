package fox

import (
	"net/http"
	"reflect"

	"github.com/fox-gonic/fox/httperrors"
)

func call(ctx *Context, handler HandlerFunc) any {

	var (
		funcValue = reflect.ValueOf(handler)
		funcType  = funcValue.Type()
		ctxValue  = reflect.ValueOf(ctx)
	)

	var (
		numIn  = funcType.NumIn()
		numOut = funcType.NumOut()
	)

	var values []reflect.Value

	switch numIn {
	case 0:
		values = funcValue.Call([]reflect.Value{})
	case 1:
		values = funcValue.Call([]reflect.Value{ctxValue})
	default:
		in := make([]reflect.Value, 0, numIn)
		in = append(in, ctxValue)
		for i := 1; i < numIn; i++ {
			// Bind handler params
			parameter := reflect.New(funcType.In(i)).Interface()
			if err := bind(ctx, parameter); err != nil {
				msg := &httperrors.Error{
					HTTPCode: http.StatusBadRequest,
					Err:      err,
					Code:     "BIND_ERROR",
				}
				return msg
			}
			in = append(in, reflect.ValueOf(parameter).Elem())
		}
		values = funcValue.Call(in)
	}

	switch numOut {
	case 0:
		return nil
	case 1:
		res := values[0].Interface()
		if err, ok := res.(error); ok {
			return err
		}
		return res

	default: // 2
		res, err := values[0].Interface(), values[1].Interface()
		if err, ok := err.(error); ok {
			return err
		}
		return res
	}
}
