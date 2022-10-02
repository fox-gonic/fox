package fox

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"

	"github.com/fox-gonic/fox/database"
	"github.com/fox-gonic/fox/internal/bytesconv"
	"github.com/fox-gonic/fox/logger"
	"github.com/fox-gonic/fox/middleware/sessions"
)

var (
	default404Body = []byte("404 page not found")
	default405Body = []byte("405 method not allowed")
)

var defaultPlatform string

var defaultTrustedCIDRs = []*net.IPNet{
	{ // 0.0.0.0/0 (IPv4)
		IP:   net.IP{0x0, 0x0, 0x0, 0x0},
		Mask: net.IPMask{0x0, 0x0, 0x0, 0x0},
	},
	{ // ::/0 (IPv6)
		IP:   net.IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Mask: net.IPMask{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	},
}

// HandlerFunc is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// wildcards (path variables).
// func(){}
// func(ctx *Context) any { ... }
// func(ctx *Context) (any, err) { ... }
// func(ctx *Context, args *AutoBindingArgType) (any) { ... }
// func(ctx *Context, args *AutoBindingArgType) (any, err) { ... }
type HandlerFunc interface{}

// HandlersChain defines a HandlerFunc slice.
type HandlersChain []HandlerFunc

// Last returns the last handler in the chain. i.e. the last handler is the main one.
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// RouteInfo represents a request route's specification which contains method and path and its handler.
type RouteInfo struct {
	Method      string
	Path        string
	Handler     string
	HandlerFunc HandlerFunc
	Handlers    HandlersChain
}

// RoutesInfo defines a RouteInfo slice.
type RoutesInfo []RouteInfo

// Engine is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Engine struct {
	Configurations *viper.Viper
	Database       *database.Database
	SessionName    string
	SessionStore   sessions.Store

	RouterGroup

	*Options

	trees methodTrees

	paramsPool sync.Pool

	pool        sync.Pool // pool of contexts that are used in a request
	maxParams   uint16
	maxSections uint16

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	noRoute    HandlersChain
	allNoRoute HandlersChain

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	noMethod    HandlersChain
	allNoMethod HandlersChain

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler HandlerFunc

	// cache is a key/value pair global for the engine.
	cache sync.Map
}

// Make sure the Router conforms with the http.Handler interface
var _ http.Handler = New(defaultEngineOptions)

// New returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func New(opts *Options) *Engine {
	engine := &Engine{
		Configurations: viper.New(),

		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},

		Options: opts,
	}

	engine.trustedCIDRs = defaultTrustedCIDRs
	engine.RouterGroup.engine = engine

	engine.pool.New = func() any {
		return engine.allocateContext()
	}
	return engine
}

// Default return engine with default options
func Default() *Engine {
	return New(defaultEngineOptions)
}

// NewWithConfig return new engine with config file
func NewWithConfig(path string) (*Engine, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	configurations := viper.New()
	configurations.SetConfigType("yaml")
	if err = configurations.ReadConfig(file); err != nil {
		return nil, err
	}

	{ // set default configurations
		configurations.SetDefault("addr", "127.0.0.1:9000")
		configurations.SetDefault("env", DevelopmentMode)

		configurations.SetDefault("logger", logger.Config{
			LogLevel:              logger.DebugLevel,
			ConsoleLoggingEnabled: true,
			EncodeLogsAsJSON:      false,
			FileLoggingEnabled:    false,
		})
	}

	engine := &Engine{
		Configurations: configurations,
	}

	// init database
	var databaseConfig *database.Config
	if err := engine.Configurations.UnmarshalKey("database", &databaseConfig); err != nil {
		return nil, err
	}
	if databaseConfig != nil {
		if engine.Database, err = database.New(databaseConfig); err != nil {
			return nil, err
		}
	}

	// set logger config
	var loggerConfig *logger.Config
	if err := configurations.UnmarshalKey("logger", &loggerConfig); err != nil {
		return nil, err
	}
	logger.SetConfig(loggerConfig)

	engine.RouterGroup = RouterGroup{
		Handlers: nil,
		basePath: "/",
		root:     true,
	}

	engine.Options = &Options{
		Addr: engine.Configurations.GetString("addr"),
		Env:  engine.Configurations.GetString("env"),

		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,

		ForwardedByClientIP: true,
		RemoteIPHeaders:     []string{"X-Forwarded-For", "X-Real-IP"},
		TrustedPlatform:     defaultPlatform,
		trustedCIDRs:        defaultTrustedCIDRs,

		DefaultContentType: MIMEJSON,
	}

	engine.trustedCIDRs = defaultTrustedCIDRs
	engine.RouterGroup.engine = engine

	engine.pool.New = func() any {
		return engine.allocateContext()
	}

	if engine.Env == DevelopmentMode {
		engine.GET("_engine_info", engine.Info)
	}

	if engine.PanicHandler == nil {
		engine.PanicHandler = Recovery()
	}
	engine.Use(Recovery())

	err = engine.InitSessionStore()
	if err != nil {
		return nil, err
	}

	err = engine.InitSessionMiddleware()
	if err != nil {
		return nil, err
	}

	return engine, nil
}

// SetOptions set engine options
func (engine *Engine) SetOptions(opts *Options) {
	engine.Options = opts
}

func (engine *Engine) allocateContext() *Context {
	params := make(Params, 0, engine.maxParams)
	skippedNodes := make([]skippedNode, 0, engine.maxSections)
	return &Context{engine: engine, Params: &params, skippedNodes: &skippedNodes}
}

// Store sets the value for a key.
func (engine *Engine) Store(key string, value any) {
	engine.cache.Store(key, value)
}

// Load returns the value stored in the map for a key, or nil if no value is present.
// The ok result indicates whether value was found in the map.
func (engine *Engine) Load(key string) (value any, exists bool) {
	return engine.cache.Load(key)
}

// MustLoad returns the value for the given key if it exists, otherwise it panics.
func (engine *Engine) MustLoad(key string) any {
	if value, exists := engine.cache.Load(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// Use attaches a global middleware to the router. i.e. the middleware attached through Use() will be
// included in the handlers chain for every single request. Even 404, 405, static files...
// For example, this is the right place for a logger or error management middleware.
func (engine *Engine) Use(middleware ...HandlerFunc) {
	engine.RouterGroup.Use(middleware...)
	engine.allNoRoute = engine.combineHandlers(engine.noRoute)
	engine.allNoMethod = engine.combineHandlers(engine.noMethod)
}

// NotFound configurable http.Handler which is called when no matching route is
// found. If it is not set, http.NotFound is used.
func (engine *Engine) NotFound(handlers ...HandlerFunc) {
	engine.noRoute = handlers
	engine.allNoRoute = engine.combineHandlers(engine.noRoute)
}

// NoMethod sets the handlers called when Engine.HandleMethodNotAllowed = true.
func (engine *Engine) NoMethod(handlers ...HandlerFunc) {
	engine.noMethod = handlers
	engine.allNoMethod = engine.combineHandlers(engine.noMethod)
}

func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(len(handlers) > 0, "there must be at least one handler")

	debugPrintRoute(method, path, handlers)

	for _, handler := range handlers {
		if handler == nil {
			panic("handler can not be nil")
		}
	}

	root := engine.trees.get(method)
	if root == nil {
		root = new(node)
		root.fullPath = "/"
		engine.trees = append(engine.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)

	// Update maxParams
	if paramsCount := countParams(path); paramsCount > engine.maxParams {
		engine.maxParams = paramsCount
	}

	if sectionsCount := countSections(path); sectionsCount > engine.maxSections {
		engine.maxSections = sectionsCount
	}
}

// Routes returns a slice of registered routes, including some useful information, such as:
// the http method, path and the handler name.
func (engine *Engine) Routes() (routes RoutesInfo) {
	for _, tree := range engine.trees {
		routes = iterate("", tree.method, routes, tree.root)
	}
	return routes
}

// isTrustedProxy will check whether the IP address is included in the trusted list according to Engine.trustedCIDRs
func (engine *Engine) isTrustedProxy(ip net.IP) bool {
	if engine.trustedCIDRs == nil {
		return false
	}
	for _, cidr := range engine.trustedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// validateHeader will parse X-Forwarded-For header and return the trusted client IP address
func (engine *Engine) validateHeader(header string) (clientIP string, valid bool) {
	if header == "" {
		return "", false
	}
	items := strings.Split(header, ",")
	for i := len(items) - 1; i >= 0; i-- {
		ipStr := strings.TrimSpace(items[i])
		ip := net.ParseIP(ipStr)
		if ip == nil {
			break
		}

		// X-Forwarded-For is appended by proxy
		// Check IPs in reverse order and stop when find untrusted proxy
		if (i == 0) || (!engine.isTrustedProxy(ip)) {
			return ipStr, true
		}
	}
	return "", false
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

	startTime := time.Now()

	ctx := engine.pool.Get().(*Context)
	ctx.reset(w, req)

	ctx.Logger.WithFields(map[string]interface{}{
		"method":    ctx.Request.Method,
		"path":      ctx.Request.URL.Path,
		"client_ip": ctx.ClientIP(),
		"type":      "REQUEST",
		"action":    "Start",
	}).Info("[Started]")

	engine.handleHTTPRequest(ctx)
	engine.pool.Put(ctx)

	ctx.Logger.WithFields(map[string]interface{}{
		"method":  ctx.Request.Method,
		"path":    ctx.Request.URL.Path,
		"status":  ctx.Writer.Status(),
		"latency": time.Since(startTime).String(),
		"type":    "REQUEST",
		"action":  "Finished",
	}).Info("[Completed]")
}

func (engine *Engine) handleHTTPRequest(ctx *Context) {
	httpMethod := ctx.Request.Method
	path := ctx.Request.URL.Path
	unescape := false

	// Find root of the tree for the given HTTP method
	t := engine.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != httpMethod {
			continue
		}
		root := t[i].root
		// Find route in tree
		value := root.getValue(path, ctx.Params, ctx.skippedNodes, unescape)
		if value.params != nil {
			ctx.Params = value.params
		}
		if value.handlers != nil {
			ctx.handlers = value.handlers
			ctx.fullPath = value.fullPath
			ctx.Next()
			ctx.Writer.WriteHeaderNow()
			return
		}
		if httpMethod != http.MethodConnect && path != "/" {
			if value.tsr && engine.RedirectTrailingSlash {
				redirectTrailingSlash(ctx)
				return
			}
			if engine.RedirectFixedPath && redirectFixedPath(ctx, root, engine.RedirectFixedPath) {
				return
			}
		}
		break
	}

	// Handle 405
	if engine.HandleMethodNotAllowed {
		for _, tree := range engine.trees {
			if tree.method == httpMethod {
				continue
			}
			if value := tree.root.getValue(path, nil, ctx.skippedNodes, unescape); value.handlers != nil {
				ctx.handlers = engine.allNoMethod
				serveError(ctx, http.StatusMethodNotAllowed, default405Body)
				return
			}
		}
	}
	ctx.handlers = engine.allNoRoute
	serveError(ctx, http.StatusNotFound, default404Body)
}

var mimePlain = []string{MIMEPlain}

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
		req.URL.Path = bytesconv.BytesToString(fixedPath)
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
		code = http.StatusPermanentRedirect
	}
	// debugPrint("redirecting request %d: %s --> %s", code, rPath, rURL)
	http.Redirect(ctx.Writer, req, rURL, code)
	ctx.Writer.WriteHeaderNow()
}
