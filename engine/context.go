package engine

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/valyala/bytebufferpool"

	"github.com/fox-gonic/fox/logger"
)

// Context with engine
type Context struct {
	*gin.Context

	Logger logger.Logger
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

		buf := bytebufferpool.Get()

		defer func() {
			buf.Reset()
			bytebufferpool.Put(buf)
		}()

		defer c.Request.Body.Close()
		if _, err := io.CopyN(buf, c.Request.Body, c.Request.ContentLength); err != nil {
			return nil, err
		}

		body = buf.Bytes()[:c.Request.ContentLength]

		c.Set(gin.BodyBytesKey, body)
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
