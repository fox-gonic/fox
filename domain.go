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

// NewSubdomainEngine new domain engine
func NewSubdomainEngine() *DomainEngine {
	return &DomainEngine{
		Engine: New(),
	}
}

// Domain add domain handler
func (engines *DomainEngine) Domain(name string, handler http.Handler, isRegexp ...bool) {

	domain := &domain{
		Name:    name,
		Handler: handler,
	}

	if len(isRegexp) > 0 && isRegexp[0] {
		if req, err := regexp.Compile(name); err == nil {
			domain.IsRegexp = true
			domain.Regexp = req
		}
	}

	engines.domains = append(engines.domains, domain)
}

// ServeHTTP conforms to the http.Handler interface.
func (engines *DomainEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	host := req.Host
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		host = parts[0]
	}

	for i := 0; i < len(engines.domains); i++ {
		var domain = engines.domains[i]

		if domain.IsRegexp && domain.Regexp.MatchString(host) {
			domain.Handler.ServeHTTP(w, req)
			return
		} else if domain.Name == host {
			domain.Handler.ServeHTTP(w, req)
			return
		}
	}

	engines.Engine.ServeHTTP(w, req)
}
