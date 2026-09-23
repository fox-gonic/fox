package fox

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type domain struct {
	Name     string
	IsRegexp bool
	Regexp   *regexp.Regexp
	Handler  http.Handler
}

// DomainEngine subdomain engine
type DomainEngine struct {
	*Engine

	GetEngine func() *Engine

	domains []*domain
}

// NewDomainEngine new domain engine
func NewDomainEngine(get ...func() *Engine) *DomainEngine {
	de := &DomainEngine{}

	if len(get) > 0 {
		de.GetEngine = get[0]
	} else {
		de.GetEngine = Default
	}

	de.Engine = de.GetEngine()

	return de
}

// NewDefaultDomainEngine new default domain engine
func NewDefaultDomainEngine() *DomainEngine {
	return NewDomainEngine(Default)
}

// Handler returns the domain router, wrapped with h2c support when enabled.
func (engine *DomainEngine) Handler() http.Handler {
	if !engine.UseH2C {
		return engine
	}

	return h2c.NewHandler(engine, &http2.Server{})
}

// Run attaches the domain router to an HTTP server and starts serving requests.
func (engine *DomainEngine) Run(addr ...string) error {
	address := resolveDomainAddress(addr)
	debugPrint("Listening and serving HTTP on %s\n", address)
	return engine.newHTTPServer(address).ListenAndServe()
}

// RunTLS attaches the domain router to an HTTPS server and starts serving requests.
func (engine *DomainEngine) RunTLS(addr, certFile, keyFile string) error {
	debugPrint("Listening and serving HTTPS on %s\n", addr)
	return engine.newHTTPServer(addr).ListenAndServeTLS(certFile, keyFile)
}

// RunUnix starts serving requests on a Unix domain socket.
func (engine *DomainEngine) RunUnix(file string) error {
	debugPrint("Listening and serving HTTP on unix:/%s", file)
	return engine.runUnix(file, engine.RunListener)
}

func (engine *DomainEngine) runUnix(file string, serve func(net.Listener) error) error {
	listener, err := net.Listen("unix", file)
	if err != nil {
		return err
	}
	defer listener.Close()
	defer os.Remove(file)

	return serve(listener)
}

// RunFd starts serving requests on an existing file descriptor.
func (engine *DomainEngine) RunFd(fd int) error {
	debugPrint("Listening and serving HTTP on fd@%d", fd)
	return engine.runFd(fd, engine.RunListener)
}

func (engine *DomainEngine) runFd(fd int, serve func(net.Listener) error) error {
	file := os.NewFile(uintptr(fd), fmt.Sprintf("fd@%d", fd))
	if file == nil {
		return fmt.Errorf("fox: invalid file descriptor: %d", fd)
	}
	defer file.Close()

	listener, err := net.FileListener(file)
	if err != nil {
		return err
	}
	defer listener.Close()

	return serve(listener)
}

// RunQUIC starts serving HTTP/3 requests with the domain router.
func (engine *DomainEngine) RunQUIC(addr, certFile, keyFile string) error {
	debugPrint("Listening and serving QUIC on %s\n", addr)
	return http3.ListenAndServeQUIC(addr, certFile, keyFile, engine.Handler())
}

// RunListener starts serving requests on listener with the domain router.
func (engine *DomainEngine) RunListener(listener net.Listener) error {
	debugPrint("Listening and serving HTTP on listener bound to %s", listener.Addr())
	return engine.newHTTPServer("").Serve(listener)
}

func (engine *DomainEngine) newHTTPServer(addr string) *http.Server {
	return &http.Server{ // #nosec G112 -- matches gin.Engine.Run; callers needing timeouts can use DomainEngine as a Handler.
		Addr:    addr,
		Handler: engine.Handler(),
	}
}

func resolveDomainAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			return ":" + port
		}
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}

// Domain registers an exact-match domain handler.
//
// Domains are matched in registration order. The first match wins, regardless
// of whether it is exact or regexp. Register regexp patterns after exact
// domains to avoid accidental shadowing.
func (engine *DomainEngine) Domain(name string, engineFunc func(subEngine *Engine)) {
	engine.server(name, false, engineFunc)
}

// DomainRegexp registers a regexp domain handler.
//
// Domains are matched in registration order. The first match wins, regardless
// of whether it is exact or regexp. Register regexp patterns after exact
// domains to avoid accidental shadowing.
func (engine *DomainEngine) DomainRegexp(name string, engineFunc func(subEngine *Engine)) {
	engine.server(name, true, engineFunc)
}

// server add domain handler
func (engine *DomainEngine) server(name string, isRegexp bool, engineFunc func(*Engine)) {
	domain := &domain{
		Name:     name,
		IsRegexp: isRegexp,
	}

	if isRegexp {
		req, err := regexp.Compile(name)
		if err != nil {
			panic("fox: invalid domain regexp pattern: " + name + ": " + err.Error())
		}

		domain.Regexp = req
	}

	subEngine := engine.GetEngine()
	engineFunc(subEngine)

	domain.Handler = subEngine

	engine.domains = append(engine.domains, domain)
}

// ServeHTTP conforms to the http.Handler interface.
func (engine *DomainEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if len(engine.domains) == 0 {
		engine.Engine.ServeHTTP(w, req)
		return
	}

	host := req.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	for i := 0; i < len(engine.domains); i++ {
		domain := engine.domains[i]
		if domain.IsRegexp && domain.Regexp.MatchString(host) {
			domain.Handler.ServeHTTP(w, req)
			return
		} else if domain.Name == host {
			domain.Handler.ServeHTTP(w, req)
			return
		}
	}

	engine.Engine.ServeHTTP(w, req)
}
