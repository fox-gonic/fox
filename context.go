package fox

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fox-gonic/fox/logger"
)

// Context with engine
type Context struct {
	*gin.Context

	engine *Engine
	Logger logger.Logger
	// Request is the http request copy from gin.Context.
	Request *http.Request
}

// RequestBody return request body bytes
// see c.ShouldBindBodyWith
func (c *Context) RequestBody() (body []byte, err error) {
	if cb, ok := c.Get(gin.BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}

	if body == nil && c.Request != nil && c.Request.Body != nil {
		var (
			buf   bytes.Buffer
			bodyR = io.TeeReader(c.Request.Body, &buf)
		)
		if body, err = io.ReadAll(bodyR); err != nil {
			return
		}

		c.Set(gin.BodyBytesKey, body)

		// copy the request body to the next handler
		c.Request.Body = io.NopCloser(&buf)
	}
	return
}

// TraceID return request id
func (c *Context) TraceID() string {
	if id, exists := c.Get(logger.TraceID); exists {
		return id.(string)
	}

	if id := c.GetHeader(logger.TraceID); len(id) > 0 {
		return id
	}

	if id := c.Context.Writer.Header().Get(logger.TraceID); len(id) > 0 {
		return id
	}

	id := logger.DefaultGenRequestID()

	c.Header(logger.TraceID, id)
	c.Set(logger.TraceID, id)

	return id
}

func (c *Context) Done() <-chan struct{} {
	return c.Request.Context().Done()
}

func (c *Context) Err() error {
	return c.Request.Context().Err()
}

func (c *Context) Value(key any) any {
	return c.Request.Context().Value(key)
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Request.Context().Deadline()
}

func (c *Context) Next() {
	c.Context.Request = c.Request
	c.Context.Next()
}
