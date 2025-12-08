package fox_test

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fox-gonic/fox"
	"github.com/fox-gonic/fox/httperrors"
)

type Foo struct {
	A string
	B string
}

type AuthInfo struct {
	Username string `json:"username"`
}

func MiddlewareFailed(c *fox.Context) (res any, err error) {
	c.Logger.Info("MiddlewareFailed")
	res = "Middleware"
	err = httperrors.ErrInvalidArguments
	return res, err
}

func MiddlewareSuccess(c *fox.Context) (res any, err error) {
	c.Logger.Info("MiddlewareSuccess")
	c.Set("user_id", int64(123))
	c.Set("auth_info", &AuthInfo{Username: "binder"})
	return res, err
}

type TestInput struct {
	Param string `uri:"param"`
	Query string `query:"query"`

	Body string `json:"body"`
}

func HandleSuccess(c *fox.Context, in TestInput) (res any, err error) {
	res = in
	return res, err
}

func HandleFailed(c *fox.Context, in *TestInput) (any, error) {
	err := &httperrors.Error{
		HTTPCode: http.StatusBadRequest,
		Code:     httperrors.ErrInvalidArguments.Code,
	}

	_ = err.AddField("message", map[string]any{
		"param": "invalid param " + in.Param,
	})

	return nil, err
}

func Ping(c *fox.Context) (res any, err error) {
	c.Logger.Info("PingHandler")
	res = Foo{"a", "b"}
	return res, err
}

type ContextBindingArgs struct {
	UserID   int64     `context:"user_id"   json:"user_id"`
	AuthInfo *AuthInfo `context:"auth_info" json:"auth_info"`
	Foo      *Foo      `context:"foo"       json:"foo,omitempty"`
}

func ContextBinding(c *fox.Context, args ContextBindingArgs) (res any, err error) {
	c.Logger.Info("ContextBinding", args)
	res = args
	return res, err
}

func TestEngine(t *testing.T) {
	r := assert.New(t)

	router := fox.New()
	router.GET("ping", MiddlewareFailed, Ping)
	router.GET("ping2", MiddlewareSuccess, Ping)
	router.GET("binding", MiddlewareSuccess, ContextBinding)

	router.POST("/handle/:param/success", HandleSuccess)
	router.POST("/handle/:param/failed", HandleFailed)

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		router.ServeHTTP(w, req)
		r.Equal(http.StatusBadRequest, w.Code)

		body := w.Body.String()
		r.JSONEq(`{"code":"INVALID_ARGUMENTS","error":"(400): invalid arguments","meta":"invalid arguments"}`, body)
	}

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping2", nil)
		router.ServeHTTP(w, req)
		r.Equal(http.StatusOK, w.Code)

		var response Foo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		r.Equal("a", response.A)
		r.Equal("b", response.B)
	}

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/binding", nil)
		router.ServeHTTP(w, req)
		r.Equal(http.StatusOK, w.Code)
		r.JSONEq(`{"user_id":123,"auth_info":{"username":"binder"}}`, w.Body.String())
	}

	{
		jsonData := map[string]string{"body": "fromBody"}
		data, _ := json.Marshal(jsonData)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/handle/hello/success?query=world", bytes.NewReader(data))
		router.ServeHTTP(w, req)
		r.Equal(http.StatusOK, w.Code)
		r.JSONEq(`{"Param":"hello","Query":"world","body":"fromBody"}`, w.Body.String())

		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/handle/hello/failed?query=world", bytes.NewReader(data))
		router.ServeHTTP(w, req)

		r.Equal(http.StatusBadRequest, w.Code)
		r.JSONEq(`{"code":"INVALID_ARGUMENTS","message":{"param":"invalid param hello"}}`, w.Body.String())
	}
}

type TestRequest struct {
	Name string
	Age  int
}

// CustomError is a custom error type for testing
type CustomError struct {
	Code    int
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}

// HTTPError is another custom error type for testing
type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func TestIsValidHandlerFunc(t *testing.T) {
	tests := []struct {
		name     string
		handler  fox.HandlerFunc
		expected bool
	}{
		{
			name:     "Empty function",
			handler:  func() {},
			expected: true,
		},
		{
			name:     "Only Context parameter",
			handler:  func(ctx *fox.Context) string { return "" },
			expected: true,
		},
		{
			name:     "Context parameter with error return",
			handler:  func(ctx *fox.Context) (int, error) { return 0, nil },
			expected: true,
		},
		{
			name:     "Context with struct parameter",
			handler:  func(ctx *fox.Context, args TestRequest) string { return "" },
			expected: true,
		},
		{
			name:     "Context with struct pointer parameter",
			handler:  func(ctx *fox.Context, args *TestRequest) int { return 0 },
			expected: true,
		},
		{
			name:     "Context with map parameter",
			handler:  func(ctx *fox.Context, args map[string]any) bool { return true },
			expected: true,
		},
		{
			name:     "Context with map parameter and error return",
			handler:  func(ctx *fox.Context, args map[string]any) ([]byte, error) { return nil, nil },
			expected: true,
		},
		{
			name:     "Too many parameters",
			handler:  func(ctx *fox.Context, args TestRequest, extra int) string { return "" },
			expected: false,
		},
		{
			name:     "Too many return values",
			handler:  func(ctx *fox.Context) (string, int, error) { return "", 0, nil },
			expected: false,
		},
		{
			name:     "First parameter is not Context",
			handler:  func(ctx string) string { return "" },
			expected: false,
		},
		{
			name:     "Second parameter is not struct or map",
			handler:  func(ctx *fox.Context, args int) string { return "" },
			expected: false,
		},
		{
			name:     "Second return value is not error",
			handler:  func(ctx *fox.Context) (string, string) { return "", "" },
			expected: false,
		},
		{
			name:     "Not a function type",
			handler:  "not a function",
			expected: false,
		},
		{
			name:     "Context parameter is not pointer type",
			handler:  func(ctx fox.Context) string { return "" },
			expected: false,
		},
		{
			name:     "Context parameter is not pointer type with error return",
			handler:  func(ctx fox.Context) (string, error) { return "", nil },
			expected: false,
		},
		{
			name:     "Context parameter is not pointer type with struct parameter",
			handler:  func(ctx fox.Context, args TestRequest) string { return "" },
			expected: false,
		},
		{
			name:     "Context parameter is not pointer type with map parameter",
			handler:  func(ctx fox.Context, args map[string]any) string { return "" },
			expected: false,
		},
		{
			name:     "Custom error return type",
			handler:  func(ctx *fox.Context) (string, *CustomError) { return "", nil },
			expected: true,
		},
		{
			name:     "HTTP error return type",
			handler:  func(ctx *fox.Context) (int, *HTTPError) { return 0, nil },
			expected: true,
		},
		{
			name:     "Custom error with struct parameter",
			handler:  func(ctx *fox.Context, args TestRequest) (bool, *CustomError) { return true, nil },
			expected: true,
		},
		{
			name:     "HTTP error with map parameter",
			handler:  func(ctx *fox.Context, args map[string]any) ([]byte, *HTTPError) { return nil, nil },
			expected: true,
		},
		{
			name:     "HTTP error with interface parameter",
			handler:  func(ctx *fox.Context, args any) ([]byte, *HTTPError) { return nil, nil },
			expected: true,
		},
		{
			name:     "Non-error second return type",
			handler:  func(ctx *fox.Context) (string, string) { return "", "" },
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fox.IsValidHandlerFunc(tt.handler)
			if result != tt.expected {
				t.Errorf("IsValidHandlerFunc() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCustomErrorRendering(t *testing.T) {
	assert := assert.New(t)

	customErr := &CustomError{
		Code:    500,
		Message: "custom error message",
	}

	handler := func(c *fox.Context) (any, error) {
		return nil, customErr
	}

	router := fox.New()

	router.RenderErrorFunc = func(ctx *fox.Context, err error) {
		var customErr *CustomError
		if errors.As(err, &customErr) {
			ctx.JSON(customErr.Code, map[string]string{
				"error": customErr.Message,
			})
		}
	}

	router.GET("/error", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(500, w.Code)
	assert.JSONEq(`{"error":"custom error message"}`, w.Body.String())
}

func TestDefaultEnableContextWithFallback(t *testing.T) {
	assert := assert.New(t)

	router := fox.New()

	assert.True(router.ContextWithFallback)

	type ctxKey struct{}
	router.Use(func(c *fox.Context) {
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxKey{}, "context value"))
		c.Next()
	})
	router.GET("/test", func(c *fox.Context) {
		val := c.Value(ctxKey{})
		if val != nil {
			c.String(200, val.(string))
		} else {
			c.String(200, "no context value")
		}
	})
	router.GET("/testGin", func(c *fox.Context) {
		val := c.Context.Value(ctxKey{})
		if val != nil {
			c.String(200, val.(string))
		} else {
			c.String(200, "no context value")
		}
	})

	t.Run("default with context value", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(200, w.Code)
		assert.Equal("context value", w.Body.String())
	})

	t.Run("disable ContextWithFallback then without context value", func(t *testing.T) {
		router.ContextWithFallback = false

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/testGin", nil)
		router.ServeHTTP(w, req)

		assert.Equal(200, w.Code)
		assert.Equal("no context value", w.Body.String())
	})
}

// TestEngine_NotFound tests custom 404 handler
func TestEngine_NotFound(t *testing.T) {
	router := fox.New()

	router.NotFound(func(c *fox.Context) {
		c.JSON(404, map[string]string{
			"error": "custom not found",
		})
	})

	router.GET("/exists", func() string {
		return "exists"
	})

	t.Run("existing route", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/exists", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "exists", w.Body.String())
	})

	t.Run("not found route", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/not-exists", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
		assert.JSONEq(t, `{"error":"custom not found"}`, w.Body.String())
	})
}

// TestEngine_NoRoute tests NoRoute handler
func TestEngine_NoRoute(t *testing.T) {
	router := fox.New()

	router.NoRoute(func(c *fox.Context) {
		c.JSON(404, map[string]string{
			"message": "route not found",
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/undefined", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
	assert.JSONEq(t, `{"message":"route not found"}`, w.Body.String())
}

// TestEngine_NoMethod tests NoMethod handler
func TestEngine_NoMethod(t *testing.T) {
	router := fox.New()
	router.HandleMethodNotAllowed = true

	router.NoMethod(func(c *fox.Context) {
		c.JSON(405, map[string]string{
			"error": "method not allowed",
		})
	})

	router.GET("/test", func() string {
		return "get"
	})

	t.Run("allowed method", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "get", w.Body.String())
	})

	t.Run("not allowed method", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 405, w.Code)
		assert.JSONEq(t, `{"error":"method not allowed"}`, w.Body.String())
	})
}

// TestEngine_Load tests Load router config func
func TestEngine_Load(t *testing.T) {
	router := fox.New()

	configFunc := func(r *fox.Engine, embedFS ...embed.FS) {
		r.GET("/loaded", func() string {
			return "loaded"
		})
	}

	router.Load(configFunc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/loaded", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "loaded", w.Body.String())
}

// TestHandlersChain_Last tests Last() method with edge cases
func TestHandlersChain_Last(t *testing.T) {
	tests := []struct {
		name     string
		chain    fox.HandlersChain
		expected fox.HandlerFunc
	}{
		{
			name:     "empty chain",
			chain:    fox.HandlersChain{},
			expected: nil,
		},
		{
			name: "single handler",
			chain: fox.HandlersChain{
				func() string { return "first" },
			},
			expected: func() string { return "first" },
		},
		{
			name: "multiple handlers",
			chain: fox.HandlersChain{
				func() string { return "first" },
				func() string { return "second" },
				func() string { return "third" },
			},
			expected: func() string { return "third" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.chain.Last()
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestSetMode tests SetMode function
func TestSetMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{
			name: "debug mode",
			mode: fox.DebugMode,
		},
		{
			name: "release mode",
			mode: fox.ReleaseMode,
		},
		{
			name: "test mode",
			mode: fox.TestMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fox.SetMode(tt.mode)
			assert.Equal(t, tt.mode, fox.Mode())
		})
	}
}

// TestMode tests Mode function
func TestMode(t *testing.T) {
	// Set to a known mode
	fox.SetMode(fox.DebugMode)
	assert.Equal(t, fox.DebugMode, fox.Mode())

	// Change mode
	fox.SetMode(fox.ReleaseMode)
	assert.Equal(t, fox.ReleaseMode, fox.Mode())

	// Change to test mode
	fox.SetMode(fox.TestMode)
	assert.Equal(t, fox.TestMode, fox.Mode())
}

// TestEngine_CORS tests CORS configuration
func TestEngine_CORS(t *testing.T) {
	t.Run("valid CORS config", func(t *testing.T) {
		router := fox.New()

		// Configure CORS with valid settings - must be called before routes
		router.CORS(cors.Config{
			AllowOrigins:     []string{"http://example.com"},
			AllowMethods:     []string{"GET", "POST"},
			AllowHeaders:     []string{"Origin", "Content-Type"},
			AllowCredentials: true,
		})

		router.GET("/test", func() string {
			return "test"
		})

		// CORS middleware should process requests
		assert.NotPanics(t, func() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", "http://example.com")
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		})
	})

	t.Run("invalid CORS config should not panic", func(t *testing.T) {
		// Test that invalid CORS config doesn't cause panic
		assert.NotPanics(t, func() {
			router := fox.New()

			// Configure CORS with invalid settings (this should not apply CORS middleware)
			router.CORS(cors.Config{
				AllowAllOrigins:  true,
				AllowOrigins:     []string{"http://example.com"}, // Invalid: can't set both
				AllowCredentials: true,
			})

			router.GET("/test", func() string {
				return "test"
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", "http://example.com")
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
		})
	})

	t.Run("CORS with wildcard origin", func(t *testing.T) {
		router := fox.New()

		router.CORS(cors.Config{
			AllowAllOrigins: true,
			AllowMethods:    []string{"GET", "POST", "PUT", "DELETE"},
		})

		router.GET("/test", func() string {
			return "test"
		})

		// Should not panic
		assert.NotPanics(t, func() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", "http://anywhere.com")
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
		})
	})
}
