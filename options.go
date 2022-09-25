package fox

import (
	"net"
)

// DefaultRemoteIPHeaders default remote ip headers
var DefaultRemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}

// Options are used to configure and define how your application engine should run.
type Options struct {

	// Addr is the bind address provided to http.Server. Default is "127.0.0.1:9000"
	// Can be set using ENV vars "ADDR" and "PORT".
	Addr string `mapstructure:"addr"`

	// website cookie key pair
	Secret string `mapstructure:"secret"`

	// Env is the "environment" in which the application engine is running. Default is "development".
	Env string `mapstructure:"env"`

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

	// ForwardedByClientIP if enabled, client IP will be parsed from the request's headers that
	// match those stored at `(*gin.Engine).RemoteIPHeaders`. If no IP was
	// fetched, it falls back to the IP obtained from
	// `(*gin.Context).Request.RemoteAddr`.
	ForwardedByClientIP bool

	// RemoteIPHeaders list of headers used to obtain the client IP when
	// `(*gin.Engine).ForwardedByClientIP` is `true` and
	// `(*gin.Context).Request.RemoteAddr` is matched by at least one of the
	// network origins of list defined by `(*gin.Engine).SetTrustedProxies()`.
	RemoteIPHeaders []string

	// TrustedPlatform if set to a constant of value gin.Platform*, trusts the headers set by
	// that platform, for example to determine the client IP
	TrustedPlatform string

	DefaultContentType string

	trustedCIDRs []*net.IPNet
}

var defaultEngineOptions = &Options{
	RedirectTrailingSlash:  true,
	RedirectFixedPath:      true,
	HandleMethodNotAllowed: true,
	HandleOPTIONS:          true,

	ForwardedByClientIP: true,
	RemoteIPHeaders:     []string{"X-Forwarded-For", "X-Real-IP"},
	TrustedPlatform:     defaultPlatform,

	DefaultContentType: MIMEJSON,
}
