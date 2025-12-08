package fox_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fox-gonic/fox"
)

func foo(c *fox.Context) (res any, err error) {
	res = "foo"
	return res, err
}

func boo(c *fox.Context) (res any, err error) {
	res = "boo"
	return res, err
}

func TestRouterGroup(t *testing.T) {
	assert := assert.New(t)

	router := fox.New()
	api := router.Group("/api")

	api.GET("foo", foo)

	api.GET("boo", boo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("foo", w.Body.String())

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/boo", nil)
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("boo", w.Body.String())
}

func TestRouterGroupHandleInvalidHandler(t *testing.T) {
	router := fox.New()

	assert.Panics(t, func() {
		router.GET("too-many-values", func(c *fox.Context) (res any, other any, err error) { return res, other, err })
	})

	assert.Panics(t, func() {
		router.GET("invalid", "not a function")
	})

	assert.Panics(t, func() {
		router.Handle(http.MethodGet, "/invalid", func(i int) string { return "" })
	})
}

func TestRouterGroup_Use(t *testing.T) {
	router := fox.New()
	type ctxKey struct{}
	router.Use(func(c *fox.Context) {
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxKey{}, "context value"))
		// do not call the c.Next()
	})
	router.GET("/test", func(c *fox.Context) {
		val := c.Value(ctxKey{})
		if val != nil {
			c.String(200, val.(string))
		} else {
			c.String(200, "no context value")
		}
	})

	t.Run("with context value", func(t *testing.T) {
		assert := assert.New(t)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(200, w.Code)
		assert.Equal("context value", w.Body.String())
	})
}

// TestRouterGroup_PresetLogger tests handler with preset logger in context
func TestRouterGroup_PresetLogger(t *testing.T) {
	router := fox.New()

	// Middleware that sets a custom logger
	router.Use(func(c *fox.Context) {
		// Preset a custom logger in context
		c.Set(fox.LoggerContextKey, c.Logger)
	})

	router.GET("/test", func(c *fox.Context) string {
		// Logger should be available from context
		return "test"
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "test", w.Body.String())
}

// TestRouterGroup_WithTraceID tests handler with existing trace ID
func TestRouterGroup_WithTraceID(t *testing.T) {
	router := fox.New()

	router.GET("/test", func(c *fox.Context) string {
		return "test"
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Pre-set a trace ID in the response writer
	w.Header().Set("X-Request-Id", "preset-trace-id")

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// TestHTTPMethods tests all HTTP method shortcuts
func TestHTTPMethods(t *testing.T) {
	router := fox.New()

	// Test DELETE
	router.DELETE("/delete", func() string {
		return "deleted"
	})

	// Test PUT
	router.PUT("/put", func() string {
		return "updated"
	})

	// Test PATCH
	router.PATCH("/patch", func() string {
		return "patched"
	})

	// Test OPTIONS
	router.OPTIONS("/options", func() string {
		return "options"
	})

	// Test HEAD
	router.HEAD("/head", func() {})

	// Test Any
	router.Any("/any", func() string {
		return "any method"
	})

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"DELETE method", http.MethodDelete, "/delete", 200, "deleted"},
		{"PUT method", http.MethodPut, "/put", 200, "updated"},
		{"PATCH method", http.MethodPatch, "/patch", 200, "patched"},
		{"OPTIONS method", http.MethodOptions, "/options", 200, "options"},
		{"HEAD method", http.MethodHead, "/head", 200, ""},
		{"Any - GET", http.MethodGet, "/any", 200, "any method"},
		{"Any - POST", http.MethodPost, "/any", 200, "any method"},
		{"Any - PUT", http.MethodPut, "/any", 200, "any method"},
		{"Any - DELETE", http.MethodDelete, "/any", 200, "any method"},
		{"Any - PATCH", http.MethodPatch, "/any", 200, "any method"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}
