package fox_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
	return
}

func MiddlewareSuccess(c *fox.Context) (res any, err error) {
	c.Logger.Info("MiddlewareSuccess")
	c.Set("user_id", int64(123))
	c.Set("auth_info", &AuthInfo{Username: "binder"})
	return
}

type TestInput struct {
	Param string `uri:"param"`
	Query string `query:"query"`

	Body string `json:"body"`
}

func HandleSuccess(c *fox.Context, in TestInput) (res any, err error) {
	res = in
	return
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
	return
}

type ContextBindingArgs struct {
	UserID   int64     `context:"user_id"   json:"user_id"`
	AuthInfo *AuthInfo `context:"auth_info" json:"auth_info"`
	Foo      *Foo      `context:"foo"       json:"foo,omitempty"`
}

func ContextBinding(c *fox.Context, args ContextBindingArgs) (res any, err error) {
	c.Logger.Info("ContextBinding", args)
	res = args
	return
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
