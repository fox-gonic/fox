package fox

import (
	"encoding/json"
	"errors"

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

	if c.engine.RenderErrorFunc != nil {
		c.engine.RenderErrorFunc(c, err)
		return
	}

	var code int
	var statusCoder StatusCoder
	if errors.As(err, &statusCoder) {
		e := statusCoder
		code = e.StatusCode()
	}
	if code == 0 {
		code = c.engine.DefaultRenderErrorStatusCode
	}

	var r render.Render
	if errors.As(err, &r) {
		c.Render(code, r)
		return
	}

	var marshaler json.Marshaler
	if errors.As(err, &marshaler) {
		c.JSON(code, marshaler)
	} else {
		c.String(code, err.Error())
	}
}

// render auto render
func (c *Context) render(res any) {
	if res == nil {
		return
	}

	status := c.Writer.Status()
	switch r := res.(type) {
	case error:
		c.renderError(r)
	case string:
		c.String(status, r)
	case render.Redirect:
		c.Redirect(r.Code, r.Location)
	case render.Render:
		c.Render(status, r)
	default:
		c.JSON(status, r)
	}

	c.Abort()
}
