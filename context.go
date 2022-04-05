package fox

import (
	"net/http"
	"sync"
)

// Context allows us to pass variables between middleware,
// manage the flow, using logger with context
type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
	Handler HandlerFunc
	Params  *Params

	engine *Engine

	// This mutex protects Keys map.
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]any
}

func (c *Context) reset() {
	*c.Params = (*c.Params)[:0]
	c.Handler = nil
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// TODO(m) Using Generics

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
