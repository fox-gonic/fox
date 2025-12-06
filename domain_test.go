package fox

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
