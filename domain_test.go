package fox

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

// Test NewDomainEngine

func TestNewDomainEngine_Default(t *testing.T) {
	de := NewDomainEngine()

	require.NotNil(t, de)
	require.NotNil(t, de.Engine)
	require.NotNil(t, de.GetEngine)
	assert.Empty(t, de.domains)
}

func TestNewDomainEngine_CustomGetter(t *testing.T) {
	customEngine := New()
	getter := func() *Engine {
		return customEngine
	}

	de := NewDomainEngine(getter)

	require.NotNil(t, de)
	require.NotNil(t, de.GetEngine)
	assert.Same(t, customEngine, de.Engine)
}

func TestNewDefaultDomainEngine(t *testing.T) {
	de := NewDefaultDomainEngine()

	require.NotNil(t, de)
	require.NotNil(t, de.Engine)
	require.NotNil(t, de.GetEngine)
}

func newDomainRoutingTestEngine() *DomainEngine {
	de := NewDomainEngine()
	de.Domain("api.example.com", func(subEngine *Engine) {
		subEngine.GET("/", func() string {
			return "api"
		})
	})
	de.Domain("admin.example.com", func(subEngine *Engine) {
		subEngine.GET("/", func() string {
			return "admin"
		})
	})
	de.GET("/", func() string {
		return "default"
	})
	return de
}

func runDomainEngineListener(t *testing.T, engine *DomainEngine) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Go(func() {
		_ = engine.RunListener(listener)
	})
	t.Cleanup(func() {
		_ = listener.Close()
		wg.Wait()
	})

	return listener.Addr().String()
}

func requestDomainRoute(t *testing.T, client *http.Client, address, host string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+address+"/", nil)
	require.NoError(t, err)
	req.Host = host

	response, err := client.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = response.Body.Close()
	})
	return response
}

func TestDomainEngine_HandlerUsesDomainRouter(t *testing.T) {
	de := newDomainRoutingTestEngine()
	handler := de.Handler()
	require.Same(t, de, handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "api.example.com"
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "api", w.Body.String())
}

func TestDomainEngine_RunListenerUsesDomainRouter(t *testing.T) {
	de := newDomainRoutingTestEngine()
	address := runDomainEngineListener(t, de)
	client := &http.Client{Timeout: 5 * time.Second}

	tests := []struct {
		host string
		want string
	}{
		{host: "api.example.com", want: "api"},
		{host: "admin.example.com", want: "admin"},
		{host: "www.example.com", want: "default"},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			response := requestDomainRoute(t, client, address, tt.host)
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Equal(t, tt.want, string(body))
		})
	}
}

func TestDomainEngine_RunListenerSupportsH2C(t *testing.T) {
	de := newDomainRoutingTestEngine()
	de.UseH2C = true
	address := runDomainEngineListener(t, de)
	transport := &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, address string, _ *tls.Config) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, address)
		},
	}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	response := requestDomainRoute(t, client, address, "api.example.com")
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	assert.Equal(t, 2, response.ProtoMajor)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "api", string(body))
}

func TestDomainEngine_Run_InvalidAddress(t *testing.T) {
	de := NewDomainEngine()

	assert.Error(t, de.Run("invalid"))
}

func TestDomainEngine_RunTLS_InvalidAddress(t *testing.T) {
	de := NewDomainEngine()

	assert.Error(t, de.RunTLS("invalid", "cert.pem", "key.pem"))
}

func TestDomainEngine_RunUnix_InvalidPath(t *testing.T) {
	de := NewDomainEngine()

	assert.Error(t, de.RunUnix(filepath.Join(t.TempDir(), "missing", "fox.sock")))
}

func TestDomainEngine_RunUnix_DelegatesAndCleansUp(t *testing.T) {
	de := NewDomainEngine()
	socketFile, err := os.CreateTemp("", "fox-*.sock")
	require.NoError(t, err)
	socketPath := socketFile.Name()
	require.NoError(t, socketFile.Close())
	require.NoError(t, os.Remove(socketPath))
	t.Cleanup(func() {
		_ = os.Remove(socketPath)
	})
	serveErr := errors.New("stop serving")

	err = de.runUnix(socketPath, func(listener net.Listener) error {
		assert.Equal(t, "unix", listener.Addr().Network())
		_, err := os.Stat(socketPath)
		assert.NoError(t, err)
		return serveErr
	})

	require.ErrorIs(t, err, serveErr)
	_, err = os.Stat(socketPath)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestDomainEngine_RunFd_InvalidDescriptor(t *testing.T) {
	de := NewDomainEngine()

	require.Error(t, de.RunFd(-1))
	assert.Error(t, de.RunFd(1<<30))
}

func TestDomainEngine_RunFd_DelegatesToListener(t *testing.T) {
	de := NewDomainEngine()
	tcpListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	file, err := tcpListener.(*net.TCPListener).File()
	require.NoError(t, err)
	require.NoError(t, tcpListener.Close())
	t.Cleanup(func() {
		_ = file.Close()
	})
	serveErr := errors.New("stop serving")

	err = de.runFd(int(file.Fd()), func(listener net.Listener) error {
		assert.Equal(t, "tcp", listener.Addr().Network())
		return serveErr
	})

	assert.ErrorIs(t, err, serveErr)
}

func TestDomainEngine_RunQUIC_MissingCertificate(t *testing.T) {
	de := NewDomainEngine()

	assert.Error(t, de.RunQUIC("127.0.0.1:0", "missing-cert.pem", "missing-key.pem"))
}

func TestResolveDomainAddress(t *testing.T) {
	t.Setenv("PORT", "")
	assert.Equal(t, ":8080", resolveDomainAddress(nil))

	t.Setenv("PORT", "9000")
	assert.Equal(t, ":9000", resolveDomainAddress(nil))
	assert.Equal(t, ":8081", resolveDomainAddress([]string{":8081"}))
	assert.Panics(t, func() {
		resolveDomainAddress([]string{":8081", ":8082"})
	})
}

// Test Domain method

func TestDomainEngine_Domain_SingleDomain(t *testing.T) {
	de := NewDomainEngine()

	de.Domain("example.com", func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "example.com"
		})
	})

	assert.Len(t, de.domains, 1)
	assert.Equal(t, "example.com", de.domains[0].Name)
	assert.False(t, de.domains[0].IsRegexp)
	assert.Nil(t, de.domains[0].Regexp)
	assert.NotNil(t, de.domains[0].Handler)
}

func TestDomainEngine_Domain_MultipleDomains(t *testing.T) {
	de := NewDomainEngine()

	de.Domain("example.com", func(subEngine *Engine) {
		subEngine.GET("/", func(ctx *Context) string {
			return "example"
		})
	})

	de.Domain("test.com", func(subEngine *Engine) {
		subEngine.GET("/", func(ctx *Context) string {
			return "test"
		})
	})

	de.Domain("demo.com", func(subEngine *Engine) {
		subEngine.GET("/", func(ctx *Context) string {
			return "demo"
		})
	})

	assert.Len(t, de.domains, 3)
}

// Test DomainRegexp method

func TestDomainEngine_DomainRegexp_ValidPattern(t *testing.T) {
	de := NewDomainEngine()

	de.DomainRegexp(`^.*\.example\.com$`, func(subEngine *Engine) {
		subEngine.GET("/", func(ctx *Context) string {
			return "subdomain"
		})
	})

	assert.Len(t, de.domains, 1)
	assert.True(t, de.domains[0].IsRegexp)
	assert.NotNil(t, de.domains[0].Regexp)
}

func TestDomainEngine_DomainRegexp_InvalidPattern(t *testing.T) {
	de := NewDomainEngine()

	assert.Panics(t, func() {
		de.DomainRegexp(`[invalid(`, func(subEngine *Engine) {
			subEngine.GET("/", func(ctx *Context) string {
				return "invalid"
			})
		})
	})
}

func TestDomainEngine_DomainRegexp_MultiplePatterns(t *testing.T) {
	de := NewDomainEngine()

	de.DomainRegexp(`^api\..*\.com$`, func(subEngine *Engine) {
		subEngine.GET("/", func(ctx *Context) string {
			return "api"
		})
	})

	de.DomainRegexp(`^admin\..*\.com$`, func(subEngine *Engine) {
		subEngine.GET("/", func(ctx *Context) string {
			return "admin"
		})
	})

	assert.Len(t, de.domains, 2)
}

// Test ServeHTTP - no domains

func TestDomainEngine_ServeHTTP_NoDomains(t *testing.T) {
	de := NewDomainEngine()
	de.GET("/test", func(ctx *Context) string {
		return "default"
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	de.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "default", w.Body.String())
}

// Test ServeHTTP - exact domain match

func TestDomainEngine_ServeHTTP_ExactDomainMatch(t *testing.T) {
	de := NewDomainEngine()

	de.Domain("example.com", func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "example.com"
		})
	})

	de.Domain("test.com", func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "test.com"
		})
	})

	tests := []struct {
		host     string
		expected string
	}{
		{"example.com", "example.com"},
		{"test.com", "test.com"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Host = tt.host
			de.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expected, w.Body.String())
		})
	}
}

func TestDomainEngine_ServeHTTP_ExactDomainNoMatch_FallbackToDefault(t *testing.T) {
	de := NewDomainEngine()

	de.Domain("example.com", func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "example"
		})
	})

	de.GET("/test", func(ctx *Context) string {
		return "default"
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "other.com"
	de.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "default", w.Body.String())
}

// Test ServeHTTP - regex domain match

func TestDomainEngine_ServeHTTP_RegexMatch(t *testing.T) {
	de := NewDomainEngine()

	de.DomainRegexp(`^.*\.example\.com$`, func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "subdomain"
		})
	})

	tests := []struct {
		host     string
		expected string
	}{
		{"api.example.com", "subdomain"},
		{"admin.example.com", "subdomain"},
		{"www.example.com", "subdomain"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Host = tt.host
			de.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expected, w.Body.String())
		})
	}
}

func TestDomainEngine_ServeHTTP_RegexNoMatch_FallbackToDefault(t *testing.T) {
	de := NewDomainEngine()

	de.DomainRegexp(`^.*\.example\.com$`, func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "subdomain"
		})
	})

	de.GET("/test", func(ctx *Context) string {
		return "default"
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "example.org"
	de.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "default", w.Body.String())
}

// Test ServeHTTP - host with port

func TestDomainEngine_ServeHTTP_HostWithPort(t *testing.T) {
	de := NewDomainEngine()

	de.Domain("example.com", func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "example"
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "example.com:8080"
	de.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "example", w.Body.String())
}

func TestDomainEngine_ServeHTTP_RegexWithPort(t *testing.T) {
	de := NewDomainEngine()

	de.DomainRegexp(`^api\..*\.com$`, func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "api"
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "api.example.com:3000"
	de.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "api", w.Body.String())
}

// Test ServeHTTP - mixed exact and regex

func TestDomainEngine_ServeHTTP_MixedExactAndRegex(t *testing.T) {
	de := NewDomainEngine()

	// Exact match
	de.Domain("example.com", func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "exact"
		})
	})

	// Regex match
	de.DomainRegexp(`^api\..*\.com$`, func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "regex"
		})
	})

	// Default
	de.GET("/test", func(ctx *Context) string {
		return "default"
	})

	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{"exact match", "example.com", "exact"},
		{"regex match", "api.test.com", "regex"},
		{"no match", "other.com", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Host = tt.host
			de.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expected, w.Body.String())
		})
	}
}

// Test ServeHTTP - priority (first match wins)

func TestDomainEngine_ServeHTTP_FirstMatchWins(t *testing.T) {
	de := NewDomainEngine()

	// First domain
	de.Domain("example.com", func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "first"
		})
	})

	// Second domain (same host - should not be reached)
	de.Domain("example.com", func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "second"
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "example.com"
	de.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// First matching domain should be used
	assert.Equal(t, "first", w.Body.String())
}

func TestDomainEngine_ServeHTTP_RegexPriority(t *testing.T) {
	de := NewDomainEngine()

	// More specific regex first
	de.DomainRegexp(`^api\.example\.com$`, func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "specific"
		})
	})

	// More general regex second
	de.DomainRegexp(`^.*\.example\.com$`, func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "general"
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "api.example.com"
	de.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// First matching pattern should be used
	assert.Equal(t, "specific", w.Body.String())
}

// Test different routes per domain

func TestDomainEngine_ServeHTTP_DifferentRoutes(t *testing.T) {
	de := NewDomainEngine()

	de.Domain("api.example.com", func(subEngine *Engine) {
		subEngine.GET("/users", func(ctx *Context) string {
			return "api users"
		})
		subEngine.POST("/users", func(ctx *Context) string {
			return "api create user"
		})
	})

	de.Domain("admin.example.com", func(subEngine *Engine) {
		subEngine.GET("/dashboard", func(ctx *Context) string {
			return "admin dashboard"
		})
	})

	tests := []struct {
		name     string
		method   string
		host     string
		path     string
		expected string
	}{
		{"api get", http.MethodGet, "api.example.com", "/users", "api users"},
		{"api post", http.MethodPost, "api.example.com", "/users", "api create user"},
		{"admin get", http.MethodGet, "admin.example.com", "/dashboard", "admin dashboard"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Host = tt.host
			de.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expected, w.Body.String())
		})
	}
}

// Test 404 handling per domain

func TestDomainEngine_ServeHTTP_404PerDomain(t *testing.T) {
	de := NewDomainEngine()

	de.Domain("example.com", func(subEngine *Engine) {
		subEngine.GET("/exists", func(ctx *Context) string {
			return "found"
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	req.Host = "example.com"
	de.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Test empty host

func TestDomainEngine_ServeHTTP_EmptyHost(t *testing.T) {
	de := NewDomainEngine()

	de.Domain("example.com", func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "example"
		})
	})

	de.GET("/test", func(ctx *Context) string {
		return "default"
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = ""
	de.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "default", w.Body.String())
}

// Benchmark tests

func BenchmarkDomainEngine_ServeHTTP_NoDomains(b *testing.B) {
	de := NewDomainEngine()
	de.GET("/test", func(ctx *Context) string {
		return "test"
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		de.ServeHTTP(w, req)
	}
}

func BenchmarkDomainEngine_ServeHTTP_ExactMatch(b *testing.B) {
	de := NewDomainEngine()

	de.Domain("example.com", func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "example"
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		de.ServeHTTP(w, req)
	}
}

func BenchmarkDomainEngine_ServeHTTP_RegexMatch(b *testing.B) {
	de := NewDomainEngine()

	de.DomainRegexp(`^.*\.example\.com$`, func(subEngine *Engine) {
		subEngine.GET("/test", func(ctx *Context) string {
			return "subdomain"
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "api.example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		de.ServeHTTP(w, req)
	}
}

func BenchmarkDomainEngine_ServeHTTP_MultipleDomains(b *testing.B) {
	de := NewDomainEngine()

	// Add 10 domains
	for i := 0; i < 10; i++ {
		de.Domain("example"+string(rune('0'+i))+".com", func(subEngine *Engine) {
			subEngine.GET("/test", func(ctx *Context) string {
				return "test"
			})
		})
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "example5.com" // Middle of the list

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		de.ServeHTTP(w, req)
	}
}
