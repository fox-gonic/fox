package fox

import (
	"fmt"
	"reflect"

	"github.com/miclle/fox/easybind"
)

func call(ctx *Context, handler HandlerFunc) (any, int, error) {
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

	if numOut > 3 {
		panic("only support handler func returns max is three")
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
			args := reflect.New(funcType.In(i)).Interface()
			if err := easybind.Bind(ctx.Request, args, ctx.Params); err != nil {
				// TODO(m) err maybe 413 Payload Too Large
				return nil, 400, err
			}
			in = append(in, reflect.ValueOf(args).Elem())
		}
		values = funcValue.Call(in)
	}

	if numOut == 0 {
		return nil, 0, nil
	}

	switch numOut {
	case 1:
		res := values[0].Interface()
		if err, ok := res.(error); ok {
			return nil, 0, err
		}
		return res, 0, nil

	case 2:
		var res, status = values[0].Interface(), values[1].Interface()
		if code, ok := status.(int); ok {
			return res, code, nil
		}
		if status == nil {
			return res, 200, nil
		}
		return res, 200, status.(error)

	default:
		var res, code, err = values[0].Interface(), values[1].Interface(), values[2].Interface()
		if err == nil {
			return res, code.(int), nil
		}
		return res, code.(int), err.(error)
	}
}
