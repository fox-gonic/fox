package fox

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test XResponseTimer WriteHeader

func TestXResponseTimer_WriteHeader_DefaultKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ginCtx.Writer = &XResponseTimer{
		ResponseWriter: ginCtx.Writer,
		start:          time.Now(),
		key:            headerXResponseTime,
	}

	// Simulate handler writing status
	ginCtx.Writer.WriteHeader(http.StatusOK)

	// Check that header was set
	header := w.Header().Get(headerXResponseTime)
	assert.NotEmpty(t, header)

	// Header format should be: "startMillis, durationNanos"
	parts := strings.Split(header, ", ")
	require.Len(t, parts, 2)

	// Verify startMillis is a valid timestamp
	startMillis, err := strconv.ParseInt(parts[0], 10, 64)
	require.NoError(t, err)
	assert.Positive(t, startMillis)

	// Verify durationNanos is a valid number
	durationNanos, err := strconv.ParseInt(parts[1], 10, 64)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, durationNanos, int64(0))
}

func TestXResponseTimer_WriteHeader_CustomKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	customKey := "X-Custom-Time"

	ginCtx.Writer = &XResponseTimer{
		ResponseWriter: ginCtx.Writer,
		start:          time.Now(),
		key:            customKey,
	}

	ginCtx.Writer.WriteHeader(http.StatusOK)

	// Check that custom header was set
	header := w.Header().Get(customKey)
	assert.NotEmpty(t, header)

	// Default header should not be set
	defaultHeader := w.Header().Get(headerXResponseTime)
	assert.Empty(t, defaultHeader)
}

func TestXResponseTimer_WriteHeader_DifferentStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", http.StatusOK},
		{"201 Created", http.StatusCreated},
		{"400 Bad Request", http.StatusBadRequest},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ginCtx, _ := gin.CreateTestContext(w)

			timer := &XResponseTimer{
				ResponseWriter: ginCtx.Writer,
				start:          time.Now(),
				key:            headerXResponseTime,
			}

			timer.WriteHeader(tt.statusCode)

			// Verify header is set
			assert.NotEmpty(t, w.Header().Get(headerXResponseTime))
		})
	}
}

// Test XResponseTimer Write

func TestXResponseTimer_Write(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	timer := &XResponseTimer{
		ResponseWriter: ginCtx.Writer,
		start:          time.Now(),
		key:            headerXResponseTime,
	}

	data := []byte("test response body")
	n, err := timer.Write(data)

	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, "test response body", w.Body.String())
}

func TestXResponseTimer_Write_MultipleWrites(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	timer := &XResponseTimer{
		ResponseWriter: ginCtx.Writer,
		start:          time.Now(),
		key:            headerXResponseTime,
	}

	// Multiple writes
	_, _ = timer.Write([]byte("Hello "))
	_, _ = timer.Write([]byte("World"))

	assert.Equal(t, "Hello World", w.Body.String())
}

func TestXResponseTimer_Write_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	timer := &XResponseTimer{
		ResponseWriter: ginCtx.Writer,
		start:          time.Now(),
		key:            headerXResponseTime,
	}

	n, err := timer.Write([]byte{})

	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Empty(t, w.Body.String())
}

// Test NewXResponseTimer middleware

func TestNewXResponseTimer_DefaultKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NewXResponseTimer())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test", w.Body.String())

	// Check response time header
	header := w.Header().Get(headerXResponseTime)
	assert.NotEmpty(t, header)

	// Verify header format
	parts := strings.Split(header, ", ")
	require.Len(t, parts, 2)
}

func TestNewXResponseTimer_CustomKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	customKey := "X-My-Timer"
	router := gin.New()
	router.Use(NewXResponseTimer(customKey))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check custom header
	header := w.Header().Get(customKey)
	assert.NotEmpty(t, header)

	// Default header should not be set
	defaultHeader := w.Header().Get(headerXResponseTime)
	assert.Empty(t, defaultHeader)
}

func TestNewXResponseTimer_WithDelay(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NewXResponseTimer())
	router.GET("/test", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	start := time.Now()
	router.ServeHTTP(w, req)
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check response time header
	header := w.Header().Get(headerXResponseTime)
	require.NotEmpty(t, header)

	// Parse duration
	parts := strings.Split(header, ", ")
	require.Len(t, parts, 2)

	durationNanos, err := strconv.ParseInt(parts[1], 10, 64)
	require.NoError(t, err)

	// Duration should be at least 10ms (in nanoseconds)
	assert.GreaterOrEqual(t, durationNanos, int64(10*time.Millisecond))

	// Duration should be close to actual elapsed time
	assert.InDelta(t, elapsed.Nanoseconds(), float64(durationNanos), float64(5*time.Millisecond))
}

func TestNewXResponseTimer_MultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NewXResponseTimer())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	// Make multiple requests
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		header := w.Header().Get(headerXResponseTime)
		assert.NotEmpty(t, header)
	}
}

func TestNewXResponseTimer_DifferentStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		path       string
		statusCode int
	}{
		{"success", "/success", http.StatusOK},
		{"created", "/created", http.StatusCreated},
		{"bad request", "/bad", http.StatusBadRequest},
		{"not found", "/notfound", http.StatusNotFound},
		{"server error", "/error", http.StatusInternalServerError},
	}

	router := gin.New()
	router.Use(NewXResponseTimer())

	for _, tt := range tests {
		router.GET(tt.path, func(c *gin.Context) {
			c.Status(tt.statusCode)
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
			header := w.Header().Get(headerXResponseTime)
			assert.NotEmpty(t, header)
		})
	}
}

func TestNewXResponseTimer_WithJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NewXResponseTimer())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "message")
	assert.NotEmpty(t, w.Header().Get(headerXResponseTime))
}

func TestNewXResponseTimer_ChainedMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Add multiple middleware
	router.Use(func(c *gin.Context) {
		c.Set("first", "1")
		c.Next()
	})
	router.Use(NewXResponseTimer())
	router.Use(func(c *gin.Context) {
		c.Set("third", "3")
		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		first, _ := c.Get("first")
		third, _ := c.Get("third")
		c.String(http.StatusOK, "first=%v, third=%v", first, third)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "first=1")
	assert.Contains(t, w.Body.String(), "third=3")
	assert.NotEmpty(t, w.Header().Get(headerXResponseTime))
}

// Test time format parsing

func TestXResponseTimer_HeaderFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NewXResponseTimer())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	header := w.Header().Get(headerXResponseTime)
	require.NotEmpty(t, header)

	// Parse header parts
	parts := strings.Split(header, ", ")
	require.Len(t, parts, 2, "header should have format: startMillis, durationNanos")

	// Part 1: start time in milliseconds
	startMillis, err := strconv.ParseInt(parts[0], 10, 64)
	require.NoError(t, err, "first part should be valid integer (startMillis)")

	// Convert to time and verify it's recent
	startTime := time.Unix(0, startMillis*int64(time.Millisecond))
	assert.WithinDuration(t, time.Now(), startTime, 1*time.Second)

	// Part 2: duration in nanoseconds
	durationNanos, err := strconv.ParseInt(parts[1], 10, 64)
	require.NoError(t, err, "second part should be valid integer (durationNanos)")
	assert.GreaterOrEqual(t, durationNanos, int64(0), "duration should be non-negative")
}

// Benchmark tests

func BenchmarkXResponseTimer_WriteHeader(b *testing.B) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	timer := &XResponseTimer{
		ResponseWriter: ginCtx.Writer,
		start:          time.Now(),
		key:            headerXResponseTime,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w = httptest.NewRecorder()
		ginCtx, _ = gin.CreateTestContext(w)
		timer.ResponseWriter = ginCtx.Writer
		timer.WriteHeader(http.StatusOK)
	}
}

func BenchmarkXResponseTimer_Write(b *testing.B) {
	gin.SetMode(gin.TestMode)

	data := []byte("test response body")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)

		timer := &XResponseTimer{
			ResponseWriter: ginCtx.Writer,
			start:          time.Now(),
			key:            headerXResponseTime,
		}

		_, _ = timer.Write(data)
	}
}

func BenchmarkNewXResponseTimer_Middleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NewXResponseTimer())
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
