package fox

import (
	"net/http"

	"github.com/gorilla/context"

	"github.com/fox-gonic/fox/middleware/sessions"
)

// SessionOptions session options
type SessionOptions struct {
	Store    string        `mapstructure:"store"`
	KeyPairs string        `mapstructure:"key_pairs"`
	Name     string        `mapstructure:"name"`
	Path     string        `mapstructure:"path"`
	Domain   string        `mapstructure:"domain"`
	MaxAge   int           `mapstructure:"max_age"`
	Secure   bool          `mapstructure:"secure"`
	HTTPOnly bool          `mapstructure:"http_only"`
	SameSite http.SameSite `mapstructure:"same_site"`
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
	if engine.SessionStore == nil {
		return nil
	}
	engine.Use(NewSessionHandler(engine.SessionName, engine.SessionStore))
	return nil
}

// InitSessionStore init session store
func (engine *Engine) InitSessionStore() error {
	var opts *SessionOptions
	if err := engine.Configurations.UnmarshalKey("session", &opts); err != nil {
		return err
	}

	if opts == nil {
		return nil
	}

	engine.SessionName = opts.Name

	var store sessions.Store
	switch opts.Store {
	case "none":
		return nil
	default: // cookie_store
		store = sessions.NewCookieStore([]byte(opts.KeyPairs))
	}

	store.Options(sessions.Options{
		Path:     opts.Path,
		Domain:   opts.Domain,
		MaxAge:   opts.MaxAge,
		Secure:   opts.Secure,
		HttpOnly: opts.HTTPOnly,
		SameSite: opts.SameSite,
	})

	engine.SessionStore = store
	return nil
}
