package engine

import (
	"embed"

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

// SetMode sets gin mode according to input string.
func SetMode(value string) {
	gin.SetMode(value)
}

// HandlerFunc middleware
type HandlerFunc interface{}

// Engine for server
type Engine struct {
	*gin.Engine

	RouterGroup
}

// NewEngine return engine instance
func NewEngine() *Engine {

	// Change gin default validator
	binding.Validator = new(DefaultValidator)

	router := gin.New()
	router.Use(Logger(), gin.Recovery())

	engine := &Engine{}
	engine.Engine = router
	engine.RouterGroup.router = &engine.Engine.RouterGroup

	return engine
}

// Use middleware
func (engine *Engine) Use(middleware ...HandlerFunc) {
	engine.RouterGroup.Use(middleware...)
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
