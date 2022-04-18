package fox

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
)

var (
	default404Body = []byte("404 page not found")
	default405Body = []byte("405 method not allowed")
)

// HandlerFunc is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// wildcards (path variables).
// func(){}
// func(ctx *Context) any { ... }
// func(ctx *Context) (any, err) { ... }
// func(ctx *Context, args *AutoBindingArgType) (any, err) { ... }
// func(ctx *Context, args *AutoBindingArgType) (any, code, err) { ... }
// func(ctx *Context, args *AutoBindingArgType) (any, code) { ... }
type HandlerFunc interface{}

// HandlersChain defines a HandlerFunc slice.
type HandlersChain []HandlerFunc

// Engine is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Engine struct {
	trees map[string]*node

	paramsPool sync.Pool

	pool      sync.Pool // pool of contexts that are used in a request
	maxParams uint16

	RouterGroup

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 308 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 308 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// If enabled, the router automatically replies to OPTIONS requests.
	// Custom OPTIONS handlers take priority over automatic replies.
	HandleOPTIONS bool

	// An optional http.Handler that is called on automatic OPTIONS requests.
	// The handler is only called if HandleOPTIONS is true and no OPTIONS
	// handler for the specific path was set.
	// The "Allowed" header is set before calling the handler.
	GlobalOPTIONS http.Handler

	// Cached value of global (*) allowed methods
	globalAllowed string

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	notFoundHandlers HandlersChain

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	methodNotAllowedHandlers HandlersChain

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
}

// Make sure the Router conforms with the http.Handler interface
var _ http.Handler = New()

// New returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func New() *Engine {
	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
	engine.RouterGroup.engine = engine
	engine.pool.New = func() any {
		return engine.allocateContext()
	}
	return engine
}

func (engine *Engine) allocateContext() *Context {
	params := make(Params, 0, engine.maxParams)
	return &Context{engine: engine, Params: &params}
}

// Use attaches a global middleware to the router. i.e. the middleware attached through Use() will be
// included in the handlers chain for every single request. Even 404, 405, static files...
// For example, this is the right place for a logger or error management middleware.
func (engine *Engine) Use(middleware ...HandlerFunc) {
	engine.RouterGroup.Use(middleware...)
}

// NotFound configurable http.Handler which is called when no matching route is
// found. If it is not set, http.NotFound is used.
func (engine *Engine) NotFound(handlers ...HandlerFunc) {
	engine.notFoundHandlers = handlers
}

// NoMethod sets the handlers called when Engine.HandleMethodNotAllowed = true.
func (engine *Engine) NoMethod(handlers ...HandlerFunc) {
	engine.methodNotAllowedHandlers = handlers
}

func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {

	varsCount := uint16(0)

	if method == "" {
		panic("method must not be empty")
	}
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	for _, handler := range handlers {
		if handler == nil {
			panic("handle must not be nil")
		}
	}

	if engine.trees == nil {
		engine.trees = make(map[string]*node)
	}

	root := engine.trees[method]
	if root == nil {
		root = new(node)
		engine.trees[method] = root

		engine.globalAllowed = engine.allowed("*", "")
	}

	root.addRoute(path, handlers)

	// Update maxParams
	if paramsCount := countParams(path); paramsCount+varsCount > engine.maxParams {
		engine.maxParams = paramsCount + varsCount
	}

	// Lazy-init paramsPool alloc func
	if engine.paramsPool.New == nil && engine.maxParams > 0 {
		engine.paramsPool.New = func() interface{} {
			ps := make(Params, 0, engine.maxParams)
			return &ps
		}
	}
}

func (engine *Engine) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		engine.PanicHandler(w, req, rcv)
	}
}

func (engine *Engine) allowed(path, reqMethod string) (allow string) {
	allowed := make([]string, 0, 9)

	if path == "*" { // server-wide
		// empty method is used for internal calls to refresh the cache
		if reqMethod == "" {
			for method := range engine.trees {
				if method == http.MethodOptions {
					continue
				}
				// Add request method to list of allowed methods
				allowed = append(allowed, method)
			}
		} else {
			return engine.globalAllowed
		}
	} else { // specific path
		for method := range engine.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == http.MethodOptions {
				continue
			}

			handle, _, _ := engine.trees[method].getValue(path, nil)
			if handle != nil {
				// Add request method to list of allowed methods
				allowed = append(allowed, method)
			}
		}
	}

	if len(allowed) > 0 {
		// Add request method to list of allowed methods
		allowed = append(allowed, http.MethodOptions)

		// Sort allowed methods.
		// sort.Strings(allowed) unfortunately causes unnecessary allocations
		// due to allowed being moved to the heap and interface conversion
		for i, l := 1, len(allowed); i < l; i++ {
			for j := i; j > 0 && allowed[j] < allowed[j-1]; j-- {
				allowed[j], allowed[j-1] = allowed[j-1], allowed[j]
			}
		}

		// return as comma separated list
		return strings.Join(allowed, ", ")
	}

	return allow
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) Run(addr string) (err error) {
	defer func() {
		if err != nil {
			fmt.Fprintf(DefaultErrorWriter, "[ERROR] %v\n", err)
		}
	}()

	err = http.ListenAndServe(addr, engine)
	return
}

// ServeHTTP makes the router implement the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := engine.pool.Get().(*Context)
	ctx.reset(w, req)
	engine.handleHTTPRequest(ctx)
	engine.pool.Put(ctx)
}

func (engine *Engine) handleHTTPRequest(ctx *Context) {
	if engine.PanicHandler != nil {
		defer engine.recv(ctx.Writer, ctx.Request)
	}

	httpMethod := ctx.Request.Method
	path := ctx.Request.URL.Path

	if root := engine.trees[httpMethod]; root != nil {
		handlers, ps, tsr := root.getValue(path, ctx.Params)

		if handlers != nil {
			ctx.handlers = handlers
			if ps != nil {
				ctx.Params = ps
			}
			ctx.Next()
			ctx.Writer.WriteHeaderNow()
			return
		}

		if httpMethod != http.MethodConnect && path != "/" {
			if tsr && engine.RedirectTrailingSlash {
				redirectTrailingSlash(ctx)
				return
			}
			// Try to fix the request path
			if engine.RedirectFixedPath && redirectFixedPath(ctx, root, engine.RedirectFixedPath) {
				return
			}
		}
	}

	// Handle OPTIONS requests
	if httpMethod == http.MethodOptions && engine.HandleOPTIONS {
		if allow := engine.allowed(path, http.MethodOptions); allow != "" {
			ctx.Writer.Header().Set("Allow", allow)
			if engine.GlobalOPTIONS != nil {
				engine.GlobalOPTIONS.ServeHTTP(ctx.Writer, ctx.Request)
			}
			return
		}
	}

	// Handle 405
	if engine.HandleMethodNotAllowed {
		if allow := engine.allowed(path, httpMethod); allow != "" {
			ctx.Writer.Header().Set("Allow", allow)
			ctx.handlers = engine.methodNotAllowedHandlers
			serveError(ctx, http.StatusMethodNotAllowed, default405Body)
			return
		}
	}

	// Handle 404
	ctx.handlers = ctx.engine.notFoundHandlers
	serveError(ctx, http.StatusNotFound, default404Body)
}

var mimePlain = []string{"text/plain"}

func serveError(c *Context, code int, defaultMessage []byte) {
	c.Writer.status = code
	c.Next()
	if c.Writer.Written() {
		return
	}
	if c.Writer.Status() == code {
		c.Writer.Header()["Content-Type"] = mimePlain
		_, err := c.Writer.Write(defaultMessage)
		if err != nil {
			// debugPrint("cannot write message to writer during serve error: %v", err)
		}
		return
	}
	c.Writer.WriteHeaderNow()
}

func redirectTrailingSlash(ctx *Context) {
	req := ctx.Request
	p := req.URL.Path
	if prefix := path.Clean(ctx.Request.Header.Get("X-Forwarded-Prefix")); prefix != "." {
		p = prefix + "/" + req.URL.Path
	}
	req.URL.Path = p + "/"
	if length := len(p); length > 1 && p[length-1] == '/' {
		req.URL.Path = p[:length-1]
	}
	redirectRequest(ctx)
}

func redirectFixedPath(ctx *Context, root *node, trailingSlash bool) bool {
	req := ctx.Request
	rPath := req.URL.Path

	if fixedPath, ok := root.findCaseInsensitivePath(CleanPath(rPath), trailingSlash); ok {
		req.URL.Path = fixedPath
		redirectRequest(ctx)
		return true
	}
	return false
}

func redirectRequest(ctx *Context) {
	req := ctx.Request
	// rPath := req.URL.Path
	rURL := req.URL.String()

	code := http.StatusMovedPermanently // Permanent redirect, request with GET method
	if req.Method != http.MethodGet {
		code = http.StatusTemporaryRedirect
	}
	// debugPrint("redirecting request %d: %s --> %s", code, rPath, rURL)
	http.Redirect(ctx.Writer, req, rURL, code)
	ctx.Writer.WriteHeaderNow()
}
