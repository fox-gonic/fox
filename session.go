package fox

import (
	"github.com/gorilla/context"

	"github.com/fox-gonic/fox/middleware/sessions"
)

// SessionOptions session options
type SessionOptions struct {
	Store    string `mapstructure:"store"`
	KeyPairs string `mapstructure:"key_pairs"`
	Name     string `mapstructure:"name"`
	sessions.Options
}

// NewSessionHandler returns a session middleware
func NewSessionHandler(name string, store sessions.Store) HandlerFunc {
	return func(c *Context) {
		s := sessions.New(name, store, c.Writer, c.Request)
		c.Set(sessions.DefaultKey, s)
		defer context.Clear(c.Request)
		c.Next()
	}
}

// ManySessionHandler returns multiple sessions middleware
func ManySessionHandler(names []string, store sessions.Store) HandlerFunc {
	return func(c *Context) {
		m := make(map[string]sessions.Session, len(names))
		for _, name := range names {
			m[name] = sessions.New(name, store, c.Writer, c.Request)
		}
		c.Set(sessions.DefaultKey, m)
		defer context.Clear(c.Request)
		c.Next()
	}
}

// DefaultSession shortcut to get session
func DefaultSession(c *Context) sessions.Session {
	return c.MustGet(sessions.DefaultKey).(sessions.Session)
}

// DefaultMany shortcut to get session with given name
func DefaultMany(c *Context, name string) sessions.Session {
	return c.MustGet(sessions.DefaultKey).(map[string]sessions.Session)[name]
}

// InitSessionMiddleware init session middleware
func (engine *Engine) InitSessionMiddleware() error {
	var opts *SessionOptions
	if err := engine.Configurations.UnmarshalKey("session", &opts); err != nil {
		return err
	}
	var store sessions.Store
	switch opts.Store {
	case "none":
		return nil
	case "gorm_store":
		if engine.Database != nil {
			store = sessions.NewGormStore(engine.Database.DB, true, []byte(opts.KeyPairs))
		}
	default: // cookie_store
		store = sessions.NewCookieStore([]byte(opts.KeyPairs))
	}
	store.Options(opts.Options)
	engine.Use(NewSessionHandler(opts.Name, store))
	return nil
}
