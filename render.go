package fox

import (
	"encoding/json"
	"net/http"

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
		code = c.engine.DefaultRenderErrorStatusCode
	}

	if r, ok := err.(render.Render); ok {
		c.Render(code, r)
		return
	}

	// TODO(m): render by writer content-type
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
	case render.Redirect:
		c.Redirect(r.Code, r.Location)
	case render.Render:
		c.Render(http.StatusOK, r)
	default:
		// TODO(m): render by writer content-type
		c.JSON(http.StatusOK, r)
	}

	c.Abort()
}
