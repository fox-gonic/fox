package fox

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fox-gonic/fox/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test Logger middleware with default config

func TestLogger_DefaultConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test", w.Body.String())

	// Check TraceID header is set
	traceID := w.Header().Get(logger.TraceID)
	assert.NotEmpty(t, traceID)
	assert.Len(t, traceID, 16) // Default ID length
}

func TestLogger_ExistingTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	existingID := "existing-trace-123"
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(logger.TraceID, existingID)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should preserve existing TraceID
	traceID := w.Header().Get(logger.TraceID)
	assert.Equal(t, existingID, traceID)
}

func TestLogger_GeneratedTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	// Should generate new TraceID
	traceID := w.Header().Get(logger.TraceID)
	assert.NotEmpty(t, traceID)
	assert.Len(t, traceID, 16)
}

// Test Logger context key

func TestLogger_LoggerInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	loggerChecked := false

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		// Get logger from context
		loggerVal, exists := c.Get(LoggerContextKey)
		assert.True(t, exists, "logger should exist in context")

		if exists {
			contextLogger, ok := loggerVal.(logger.Logger)
			assert.True(t, ok, "logger should be correct type")
			assert.NotNil(t, contextLogger, "logger should not be nil")
			loggerChecked = true
		}

		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, loggerChecked, "logger checks should have been performed")
}

func TestLogger_LoggerContextKey(t *testing.T) {
	assert.Equal(t, "_fox-goinc/fox/logger/context/key", LoggerContextKey)
}

// Test SkipPaths config

func TestLogger_SkipPaths_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger(LoggerConfig{
		SkipPaths: []string{},
	}))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(logger.TraceID))
}

func TestLogger_SkipPaths_SinglePath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger(LoggerConfig{
		SkipPaths: []string{"/health"},
	}))
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	// Request to skipped path
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(logger.TraceID))

	// Request to non-skipped path
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(logger.TraceID))
}

func TestLogger_SkipPaths_MultiplePaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger(LoggerConfig{
		SkipPaths: []string{"/health", "/metrics", "/ping"},
	}))

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	router.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "metrics")
	})
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	router.GET("/api", func(c *gin.Context) {
		c.String(http.StatusOK, "api")
	})

	tests := []struct {
		path         string
		expectedBody string
	}{
		{"/health", "ok"},
		{"/metrics", "metrics"},
		{"/ping", "pong"},
		{"/api", "api"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
			assert.NotEmpty(t, w.Header().Get(logger.TraceID))
		})
	}
}

// Test with query parameters

func TestLogger_WithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/search", func(c *gin.Context) {
		c.String(http.StatusOK, "results")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/search?q=test&page=1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(logger.TraceID))
}

func TestLogger_WithEmptyQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test different HTTP methods

func TestLogger_DifferentMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/get"},
		{http.MethodPost, "/post"},
		{http.MethodPut, "/put"},
		{http.MethodDelete, "/delete"},
		{http.MethodPatch, "/patch"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			router := gin.New()
			router.Use(Logger())
			router.Handle(tt.method, tt.path, func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.NotEmpty(t, w.Header().Get(logger.TraceID))
		})
	}
}

// Test different status codes

func TestLogger_DifferentStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		path   string
		status int
	}{
		{"200 OK", "/ok", http.StatusOK},
		{"201 Created", "/created", http.StatusCreated},
		{"400 Bad Request", "/bad", http.StatusBadRequest},
		{"404 Not Found", "/notfound", http.StatusNotFound},
		{"500 Internal Error", "/error", http.StatusInternalServerError},
	}

	router := gin.New()
	router.Use(Logger())

	for _, tt := range tests {
		router.GET(tt.path, func(c *gin.Context) {
			c.Status(tt.status)
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)
			assert.NotEmpty(t, w.Header().Get(logger.TraceID))
		})
	}
}

// Test with errors

func TestLogger_WithErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/error", func(c *gin.Context) {
		c.Error(assert.AnError) //nolint:errcheck
		c.String(http.StatusInternalServerError, "error")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotEmpty(t, w.Header().Get(logger.TraceID))
}

func TestLogger_WithPrivateErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/error", func(c *gin.Context) {
		_ = c.Error(assert.AnError).SetType(gin.ErrorTypePrivate)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(logger.TraceID))
}

// Test with nil request header

func TestLogger_NilRequestHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header = nil // Set to nil
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test middleware chain

func TestLogger_MiddlewareChain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var firstCalled, loggerCalled, thirdCalled bool

	router := gin.New()
	router.Use(func(c *gin.Context) {
		firstCalled = true
		c.Next()
	})
	router.Use(Logger())
	router.Use(func(c *gin.Context) {
		// Check logger was set
		_, exists := c.Get(LoggerContextKey)
		loggerCalled = exists
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		thirdCalled = true
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.True(t, firstCalled)
	assert.True(t, loggerCalled)
	assert.True(t, thirdCalled)
}

// Test client IP logging

func TestLogger_ClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(logger.TraceID))
}

func TestLogger_ClientIP_WithXForwardedFor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(logger.TraceID))
}

// Test logger output

func TestLogger_LogOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	loggerValid := false

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		// Get logger from context
		loggerVal, exists := c.Get(LoggerContextKey)
		assert.True(t, exists, "logger should exist in context")

		if exists {
			log, ok := loggerVal.(logger.Logger)
			assert.True(t, ok, "logger should be correct type")
			assert.NotNil(t, log, "logger should not be nil")
			loggerValid = ok && log != nil
		}

		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, loggerValid, "logger should have been validated")
}

// Test concurrent requests

func TestLogger_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	done := make(chan bool)
	numRequests := 10

	for i := 0; i < numRequests; i++ {
		go func() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.NotEmpty(t, w.Header().Get(logger.TraceID))
			done <- true
		}()
	}

	// Wait for all requests
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// Benchmark tests

func BenchmarkLogger_NoSkip(b *testing.B) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkLogger_WithSkip(b *testing.B) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger(LoggerConfig{
		SkipPaths: []string{"/health", "/metrics"},
	}))
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkLogger_WithQueryParams(b *testing.B) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/search", func(c *gin.Context) {
		c.String(http.StatusOK, "results")
	})

	req := httptest.NewRequest(http.MethodGet, "/search?q=test&page=1", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
