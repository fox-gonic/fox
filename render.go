package fox

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/fox-gonic/fox/render"
)

// StatusCoder is a interface for http status code
type StatusCoder interface {
	StatusCode() int
}

// renderError render error
func (c *Context) renderError(err error) {
	if err == nil {
		return
	}

	var code int
	if e, ok := err.(StatusCoder); ok {
		code = e.StatusCode()
	}
	if code == 0 {
		code = DefaultRenderErrorStatusCode
	}

	if r, ok := err.(render.Render); ok {
		c.Render(code, r)
		return
	}

	value := reflect.TypeOf(err)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// TODO(m): render by writer content-type

	switch value.Kind() {
	case reflect.Struct, reflect.Map, reflect.Array, reflect.Slice:
		c.JSON(code, err)
		return
	}

	if e, ok := err.(json.Marshaler); ok {
		c.JSON(code, e)
	} else {
		c.String(code, err.Error())
	}
}

// render auto render
func (c *Context) render(res any) {
	if res == nil {
		return
	}

	switch r := res.(type) {
	case error:
		c.renderError(r)
	case string:
		c.String(http.StatusOK, r)
	case render.Render:
		c.Render(http.StatusOK, r)
	default:
		// TODO(m): render by writer content-type
		c.JSON(http.StatusOK, r)
	}

	c.Abort()
}
