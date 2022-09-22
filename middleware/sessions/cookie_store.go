package sessions

import "github.com/gorilla/sessions"

// NewCookieStore returns a new CookieStore.
//
// Keys are defined in pairs to allow key rotation, but the common case is
// to set a single authentication key and optionally an encryption key.
//
// The first key in a pair is used for authentication and the second for
// encryption. The encryption key can be set to nil or omitted in the last
// pair, but the authentication key is required in all pairs.
//
// It is recommended to use an authentication key with 32 or 64 bytes.
// The encryption key, if set, must be either 16, 24, or 32 bytes to select
// AES-128, AES-192, or AES-256 modes.
func NewCookieStore(keyPairs ...[]byte) Store {
	return &CookieStore{sessions.NewCookieStore(keyPairs...)}
}

var _ Store = &CookieStore{}

// CookieStore wrap gorilla cookie store
// default options:
//
//	Path: "/"
//	MaxAge: 86400 * 30
type CookieStore struct {
	*sessions.CookieStore
}

// Options implemented Store
func (c *CookieStore) Options(options Options) {
	c.CookieStore.Options = &options
}
