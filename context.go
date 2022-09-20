package fox

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fox-gonic/fox/logger"
	"github.com/fox-gonic/fox/render"
)

// Context allows us to pass variables between middleware,
// manage the flow, using logger with context
type Context struct {
	Request *http.Request
	Writer  *ResponseWriter
	Params  *Params

	requestID string
	Logger    logger.Logger

	engine       *Engine
	skippedNodes *[]skippedNode

	handlers HandlersChain
	index    int
	fullPath string

	// This mutex protects Keys map.
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]any
}

func (c *Context) reset(w http.ResponseWriter, req *http.Request) {
	c.Writer = &ResponseWriter{
		ResponseWriter: w,
		size:           noWritten,
		status:         defaultStatus,
	}
	c.Request = req
	*c.Params = (*c.Params)[:0]

	c.requestID = c.GetHeader(logger.TraceIDKey)
	if c.requestID == "" {
		c.requestID = logger.DefaultGenRequestID()
		c.Header(logger.TraceIDKey, c.requestID)
		c.Set(logger.TraceIDKey, c.requestID)
	}

	if logger.New != nil {
		c.Logger = logger.New(c.requestID)
	}

	c.handlers = nil
	c.index = -1
	c.fullPath = ""
	c.Keys = nil
	*c.skippedNodes = (*c.skippedNodes)[:0]
}

// Next should be used only inside middleware.
func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		res, err := call(c, c.handlers[c.index])
		if err != nil {
			c.renderError(err)
			return
		}
		if res != nil {
			c.render(res)
		}
		c.index++
	}
}

// renderError ...
func (c *Context) renderError(err error) {

	var parsedError *Error
	if errors.As(err, &parsedError) {
		var accepts = []string{c.engine.DefaultContentType}
		accepts = append(accepts, c.Accepts()...)
		if e := parsedError.Render(c.Writer, accepts...); e != nil {
			panic(e)
		}
		return
	}

	if v, ok := err.(WriteHeader); ok {
		c.Writer.WriteHeader(v.StatusCode())
	} else {
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}

	if r, ok := err.(Render); ok {
		if e := r.Render(c.Writer); e != nil {
			panic(e)
		}
	} else {
		c.Writer.Write([]byte(err.Error())) // nolint: errcheck
	}
}

// render writes the response headers and calls render.Render to render data.
func (c *Context) render(res any) {

	var r Render
	switch v := res.(type) {
	case error:
		c.renderError(v)
		return
	case string:
		r = render.String{Format: v}
	case render.Redirect:
		v.Request = c.Request
		r = v
		c.Writer.WriteHeader(-1)
	case render.String, render.JSON, render.IndentedJSON, render.JsonpJSON, render.XML,
		render.Data, render.HTML, render.YAML, render.Reader, render.ASCIIJSON, render.ProtoBuf:
		r = v.(Render)
	default:
		if crender, ok := res.(Render); ok {
			r = crender
			break
		}
		switch c.engine.DefaultContentType {
		case MIMEJSON:
			r = render.JSON{Data: res}

		case MIMEXML, MIMEXML2:
			r = render.XML{Data: res}

		case MIMEPROTOBUF:
			r = render.ProtoBuf{Data: res}

		case MIMEYAML:
			r = render.YAML{Data: res}

		default: // MIMEJSON
			r = render.JSON{Data: res}
		}
	}

	r.WriteContentType(c.Writer)
	if err := r.Render(c.Writer); err != nil {
		panic(err)
	}
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}

	c.Keys[key] = value
	c.mu.Unlock()
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (c *Context) Get(key string) (value any, exists bool) {
	c.mu.RLock()
	value, exists = c.Keys[key]
	c.mu.RUnlock()
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (c *Context) MustGet(key string) any {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// Engine return the engine that was used to create this context.
func (c *Context) Engine() *Engine {
	return c.engine
}

// ClientIP implements one best effort algorithm to return the real client IP.
// It calls c.RemoteIP() under the hood, to check if the remote IP is a trusted proxy or not.
// If it is it will then try to parse the headers defined in Engine.RemoteIPHeaders (defaulting to [X-Forwarded-For, X-Real-Ip]).
// If the headers are not syntactically valid OR the remote IP does not correspond to a trusted proxy,
// the remote IP (coming from Request.RemoteAddr) is returned.
func (c *Context) ClientIP() string {
	// Check if we're running on a trusted platform, continue running backwards if error
	if c.engine.TrustedPlatform != "" {
		// Developers can define their own header of Trusted Platform or use predefined constants
		if addr := c.requestHeader(c.engine.TrustedPlatform); addr != "" {
			return addr
		}
	}

	// It also checks if the remoteIP is a trusted proxy or not.
	// In order to perform this validation, it will see if the IP is contained within at least one of the CIDR blocks
	// defined by Engine.SetTrustedProxies()
	remoteIP := net.ParseIP(c.RemoteIP())
	if remoteIP == nil {
		return ""
	}
	trusted := c.engine.isTrustedProxy(remoteIP)

	if trusted && c.engine.ForwardedByClientIP && c.engine.RemoteIPHeaders != nil {
		for _, headerName := range c.engine.RemoteIPHeaders {
			ip, valid := c.engine.validateHeader(c.requestHeader(headerName))
			if valid {
				return ip
			}
		}
	}
	return remoteIP.String()
}

// RemoteIP parses the IP from Request.RemoteAddr, normalizes and returns the IP (without the port).
func (c *Context) RemoteIP() string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

// ContentType returns the Content-Type header of the request.
func (c *Context) ContentType() string {
	return filterFlags(c.requestHeader("Content-Type"))
}

// Accepts returns the Accept header of the request.
func (c *Context) Accepts() []string {
	return parseAccept(c.requestHeader("Accept"))
}

// Header is a intelligent shortcut for c.Writer.Header().Set(key, value).
// It writes a header in the response.
// If value == "", this method removes the header `c.Writer.Header().Del(key)`
func (c *Context) Header(key, value string) {
	if value == "" {
		c.Writer.Header().Del(key)
		return
	}
	c.Writer.Header().Set(key, value)
}

// GetHeader returns value from request headers.
func (c *Context) GetHeader(key string) string {
	return c.requestHeader(key)
}

func (c *Context) requestHeader(key string) string {
	return c.Request.Header.Get(key)
}

/************************************/
/**** HTTPS://PKG.GO.DEV/CONTEXT ****/
/************************************/

func (c *Context) getRequestContext() context.Context {
	if c.Request == nil || c.Request.Context() == nil {
		return context.Background()
	}
	return c.Request.Context()
}

// Deadline returns that there is no deadline (ok==false) when c.Request has no Context.
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.getRequestContext().Deadline()
}

// Done returns nil (chan which will wait forever) when c.Request has no Context.
func (c *Context) Done() <-chan struct{} {
	return c.getRequestContext().Done()
}

// Err returns nil when c.Request has no Context.
func (c *Context) Err() error {
	return c.getRequestContext().Err()
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *Context) Value(key any) any {
	if key == 0 {
		return c.Request
	}
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}
	return c.getRequestContext().Value(key)
}
