package fox

import (
	"embed"
	"io"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	// DebugMode indicates gin mode is debug.
	DebugMode = gin.DebugMode
	// ReleaseMode indicates gin mode is release.
	ReleaseMode = gin.ReleaseMode
	// TestMode indicates gin mode is test.
	TestMode = gin.TestMode
)

var foxMode = DebugMode

// SetMode sets gin mode according to input string.
func SetMode(value string) {
	gin.SetMode(value)
	foxMode = value
}

// Mode returns current fox mode.
func Mode() string {
	return foxMode
}

// DefaultWriter is the default io.Writer used by Gin for debug output and
// middleware output like Logger() or Recovery().
// Note that both Logger and Recovery provides custom ways to configure their
// output io.Writer.
// To support coloring in Windows use:
//
//	import "github.com/mattn/go-colorable"
//	gin.DefaultWriter = colorable.NewColorableStdout()
var DefaultWriter io.Writer = os.Stdout

// DefaultErrorWriter is the default io.Writer used by Gin to debug errors
var DefaultErrorWriter io.Writer = os.Stderr

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

// Engine for server
type Engine struct {
	*gin.Engine

	RouterGroup

	// DefaultRenderErrorStatusCode is the default http status code used for automatic rendering
	DefaultRenderErrorStatusCode int
}

// New return engine instance
func New() *Engine {

	// Change gin default validator
	binding.Validator = new(DefaultValidator)

	engine := &Engine{
		Engine: gin.New(),
	}
	engine.RouterGroup.router = &engine.Engine.RouterGroup
	engine.RouterGroup.engine = engine

	engine.DefaultRenderErrorStatusCode = http.StatusBadRequest

	engine.Use(NewXResponseTimer(), Logger(), gin.Recovery())
	return engine
}

// Use middleware
func (engine *Engine) Use(middleware ...HandlerFunc) {
	engine.RouterGroup.Use(middleware...)
}

// NotFound adds handlers for NoRoute. It returns a 404 code by default.
func (engine *Engine) NotFound(handlers ...HandlerFunc) {
	handlersChain := engine.RouterGroup.handleWrapper(handlers...)
	engine.Engine.NoRoute(handlersChain...)
}

// CORS config
func (engine *Engine) CORS(config cors.Config) {
	if config.Validate() == nil {
		engine.Engine.Use(cors.New(config))
	}
}

// RouterConfigFunc engine load router config func
type RouterConfigFunc func(router *Engine, embedFS ...embed.FS)

// Load router config
func (engine *Engine) Load(f RouterConfigFunc, fs ...embed.FS) {
	f(engine, fs...)
}
