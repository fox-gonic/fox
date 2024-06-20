package fox

import (
	"net/http"
	"regexp"
	"strings"
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

	domains []*domain
}

// NewDomainEngine new domain engine
func NewDomainEngine() *DomainEngine {
	return &DomainEngine{
		Engine: New(),
	}
}

// Domain add domain handler
func (engine *DomainEngine) Domain(name string, engineFunc func(subEngine *Engine)) {
	engine.server(name, false, engineFunc)
}

// DomainRegexp add domain handler
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
			panic(err)
		}

		domain.Regexp = req
	}

	subEngine := New()
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
		var domain = engine.domains[i]
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
