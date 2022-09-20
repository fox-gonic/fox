package fox

import (
	"net/http"
	"regexp"
)

var (
	// regEnLetter matches english letters for http method name
	regEnLetter = regexp.MustCompile("^[A-Z]+$")

	// anyMethods for RouterGroup Any method
	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

// RouterGroup is used internally to configure router, a RouterGroup is associated with
// a prefix and an array of handlers (middleware).
type RouterGroup struct {
	Handlers HandlersChain
	basePath string
	engine   *Engine
	root     bool
}

// Use adds middleware to the group, see example code in GitHub.
func (group *RouterGroup) Use(middleware ...HandlerFunc) {
	group.Handlers = group.combineHandlers(middleware)
}

// Group creates a new router group. You should add all the routes that have common middlewares or the same path prefix.
// For example, all the routes that use a common middleware for authorization could be grouped.
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
	}
}

func (group *RouterGroup) handle(method, relativePath string, handlers HandlersChain) {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.engine.addRoute(method, absolutePath, handlers)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (group *RouterGroup) Handle(method, path string, handlers ...HandlerFunc) {
	if matched := regEnLetter.MatchString(method); !matched {
		panic("http method " + method + " is not valid")
	}

	group.handle(method, path, handlers)
}

// GET is a shortcut for router.Handle(http.MethodGet, path, handle)
func (group *RouterGroup) GET(path string, handlers ...HandlerFunc) {
	group.handle(http.MethodGet, path, handlers)
}

// HEAD is a shortcut for router.Handle(http.MethodHead, path, handle)
func (group *RouterGroup) HEAD(path string, handlers ...HandlerFunc) {
	group.handle(http.MethodHead, path, handlers)
}

// OPTIONS is a shortcut for router.Handle(http.MethodOptions, path, handle)
func (group *RouterGroup) OPTIONS(path string, handlers ...HandlerFunc) {
	group.handle(http.MethodOptions, path, handlers)
}

// POST is a shortcut for router.Handle(http.MethodPost, path, handle)
func (group *RouterGroup) POST(path string, handlers ...HandlerFunc) {
	group.handle(http.MethodPost, path, handlers)
}

// PUT is a shortcut for router.Handle(http.MethodPut, path, handle)
func (group *RouterGroup) PUT(path string, handlers ...HandlerFunc) {
	group.handle(http.MethodPut, path, handlers)
}

// PATCH is a shortcut for router.Handle(http.MethodPatch, path, handle)
func (group *RouterGroup) PATCH(path string, handlers ...HandlerFunc) {
	group.handle(http.MethodPatch, path, handlers)
}

// DELETE is a shortcut for router.Handle(http.MethodDelete, path, handle)
func (group *RouterGroup) DELETE(path string, handlers ...HandlerFunc) {
	group.handle(http.MethodDelete, path, handlers)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (group *RouterGroup) Any(relativePath string, handlers ...HandlerFunc) {
	for _, method := range anyMethods {
		group.handle(method, relativePath, handlers)
	}
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//
//	er.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (group *RouterGroup) ServeFiles(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	fileServer := http.FileServer(root)

	group.GET(path, func(c *Context) {
		c.Request.URL.Path = c.Params.ByName("filepath")
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

func (group *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(group.Handlers) + len(handlers)
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}
