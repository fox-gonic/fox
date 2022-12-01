package fox

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/miclle/binding"
)

func call(ctx *Context, handler HandlerFunc) (any, error) {
	var (
		funcValue = reflect.ValueOf(handler)
		funcType  = funcValue.Type()
		ctxValue  = reflect.ValueOf(ctx)
	)

	// TODO(m) check handler type when route registering
	if funcValue.Kind() != reflect.Func {
		panic(fmt.Sprintf("%#v is not a function", handler))
	}

	var (
		numIn  = funcType.NumIn()
		numOut = funcType.NumOut()
	)

	if numOut > 2 {
		panic("only support handler func returns max is two values")
	}

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
			if err := binding.Bind(ctx.Request, parameter, ctx.Params.Values()); err != nil {
				// TODO(m) err maybe 413 Payload Too Large
				msg := &Error{
					Status: http.StatusBadRequest,
					Err:    err,
				}
				return nil, msg
			}
			in = append(in, reflect.ValueOf(parameter).Elem())
		}
		values = funcValue.Call(in)
	}

	switch numOut {
	case 0:
		return nil, nil
	case 1:
		res := values[0].Interface()
		if err, ok := res.(error); ok {
			return nil, err
		}
		return res, nil

	default: // 2
		res, err := values[0].Interface(), values[1].Interface()
		if err, ok := err.(error); ok {
			return res, err
		}
		return res, nil
	}
}
