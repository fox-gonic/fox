package engine

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valyala/bytebufferpool"

	"github.com/fox-gonic/fox/logger"
)

// Context with engine
type Context struct {
	Context *gin.Context

	Logger logger.Logger
}

// params return context params url.Values
func (c *Context) params() map[string][]string {
	m := map[string][]string{}
	for _, param := range c.Context.Params {
		m[param.Key] = []string{param.Value}
	}
	return m
}

// Abort 中断后续 handlers。 注：当前 handler 需要主动 return
func (c *Context) Abort() {
	c.Context.Abort()
}

// Next should be used only inside middleware.
func (c *Context) Next() {
	c.Context.Next()
}

// Set 在 context 上下文中一个新的键/值对
func (c *Context) Set(key string, value interface{}) {
	c.Context.Set(key, value)
}

// Get 从 context 上下文中获取指定键的值，可通过 exists 判断是否存在
func (c *Context) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Context.Get(key)
	return
}

// MustGet 返回给定键值(如果存在)，否则 panic
func (c *Context) MustGet(key string) interface{} {
	return c.Context.MustGet(key)
}

// Cookie 获取 cookie
func (c *Context) Cookie(name string) (string, error) {
	return c.Context.Cookie(name)
}

// GetHeader 获取请求头信息
// 等价于 http.Request.Header.Get(key)
func (c *Context) GetHeader(key string) string {
	return c.Context.GetHeader(key)
}

// Header 设置请求头信息
func (c *Context) Header(key, value string) {
	c.Context.Header(key, value)
}

// Bind 根据 Content-Type 自动解析绑定对象
//
//		Content-Type                      | Binding      | Struct tag
//	 ----------------------------------|--------------|--------------------
//	 application/json                  | JSON binding | `json:"field_name"`
//	 application/xml                   | XML binding  | `xml:"field_name"`
//	 application/x-www-form-urlencoded | FORM binding | `form:"field_name"`
//	 multipart/form-data               | FORM binding | `form:"field_name"`
//	 GET request method                | FORM binding | `form:"field_name"`
func (c *Context) Bind(obj interface{}) error {
	return c.Context.Bind(obj)
}

// BindJSON application/json
func (c *Context) BindJSON(obj interface{}) error {
	return c.Context.BindJSON(obj)
}

// BindQuery bind GET request or application/x-www-form-urlencoded, multipart/form-data
func (c *Context) BindQuery(obj interface{}) error {
	return c.Context.BindQuery(obj)
}

// ShouldBindQuery 只绑定查询参数
func (c *Context) ShouldBindQuery(obj interface{}) error {
	return c.Context.ShouldBindQuery(obj)
}

// ShouldBindURI 只绑定 URL 中参数
// https://github.com/gin-gonic/gin#bind-uri
func (c *Context) ShouldBindURI(obj interface{}) error {
	return c.Context.ShouldBindUri(obj)
}

// Query 获取 url 中的参数
//
//	    GET /path?id=1234&name=Manu&value=
//		   c.Query("id") == "1234"
//		   c.Query("name") == "Manu"
//		   c.Query("value") == ""
//		   c.Query("wtf") == ""
func (c *Context) Query(key string) string {
	return c.Context.Query(key)
}

// Param 获取 URL 中的参数
// GET /user/:id
// GET /user/qiniu	c.Param("id") => "qiniu"
// GET /user/12345	c.Param("id") => "12345"
func (c *Context) Param(key string) string {
	return c.Context.Param(key)
}

// FormFile 返回提供的表单键的第一个文件
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	return c.Context.FormFile(name)
}

// MultipartForm is the parsed multipart form, including file uploads.
func (c *Context) MultipartForm() (*multipart.Form, error) {
	return c.Context.MultipartForm()
}

// ClientIP 获取客户端 IP
// 如果 gin.ForwardedByClientIP == true (默认已开启)
//
//	ip 会从 X-Forwarded-For 或 X-Real-Ip Header 中获取
//
// 其他情况从 X-Appengine-Remote-Addr Header 或 request.RemoteAddr 中获取
func (c *Context) ClientIP() string {
	return c.Context.ClientIP()
}

// RequestBody return request body bytes
// see c.ShouldBindBodyWith
func (c *Context) RequestBody() (body []byte, err error) {

	if cb, ok := c.Get(gin.BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}

	if request := c.Request(); body == nil && request.Body != nil {

		buf := bytebufferpool.Get()
		defer bytebufferpool.Put(buf)

		defer request.Body.Close()
		if _, err := io.CopyN(buf, request.Body, request.ContentLength); err != nil {
			return nil, err
		}

		body = buf.Bytes()

		c.Set(gin.BodyBytesKey, body)
	}
	return
}

// --------------------------------------------------------------------

// Request with http.Request
func (c *Context) Request() *http.Request {
	return c.Context.Request
}

// Writer with gin.ResponseWriter
func (c *Context) Writer() gin.ResponseWriter {
	return c.Context.Writer
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

// HTML render html
func (c *Context) HTML(code int, name string, obj interface{}) {
	c.Context.HTML(code, name, obj)
}

// FileFromFS writes the specified file from http.FileSytem into the body stream in an efficient way.
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	c.Context.FileFromFS(filepath, fs)
}

/************************************/
/**** HTTPS://PKG.GO.DEV/CONTEXT ****/
/************************************/

func (c *Context) getRequestContext() context.Context {
	if c.Request() == nil || c.Request().Context() == nil {
		return context.Background()
	}
	return c.Request().Context()
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
		return nil
	}
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}
	return c.getRequestContext().Value(key)
}
