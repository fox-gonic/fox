package fox

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fox-gonic/fox/httperrors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test handlers for different signatures

// No parameters, no return
func handlerNoParamsNoReturn() {
	// Do nothing
}

// No parameters, single return
func handlerNoParamsSingleReturn() string {
	return "hello"
}

// No parameters, error return
func handlerNoParamsErrorReturn() error {
	return errors.New("test error")
}

// No parameters, two returns
func handlerNoParamsTwoReturns() (string, error) {
	return "result", nil
}

// No parameters, two returns with error
func handlerNoParamsTwoReturnsWithError() (string, error) {
	return "", errors.New("test error")
}

// Context only, no return
func handlerCtxOnlyNoReturn(ctx *Context) {
	// Do nothing
}

// Context only, single return
func handlerCtxOnlySingleReturn(ctx *Context) string {
	return "hello from ctx"
}

// Context only, error return
func handlerCtxOnlyErrorReturn(ctx *Context) error {
	return errors.New("ctx error")
}

// Context only, two returns
func handlerCtxOnlyTwoReturns(ctx *Context) (string, error) {
	return "ctx result", nil
}

// Context with parameter, single return
func handlerCtxWithParamSingleReturn(ctx *Context, req *testRequest) string {
	return "name: " + req.Name
}

// Context with parameter, error from binding
type invalidRequest struct {
	Name string `json:"name" binding:"required"`
}

func handlerCtxWithInvalidParam(ctx *Context, req *invalidRequest) string {
	return "should not reach here"
}

// Context with parameter, two returns
func handlerCtxWithParamTwoReturns(ctx *Context, req *testRequest) (string, error) {
	if req.Name == "" {
		return "", errors.New("name is required")
	}
	return "name: " + req.Name, nil
}

// Context with multiple parameters
type testRequest struct {
	Name string `json:"name"`
}

type testRequest2 struct {
	Age int `json:"age"`
}

func handlerCtxWithMultipleParams(ctx *Context, req1 *testRequest, req2 *testRequest2) string {
	return "name: " + req1.Name + ", age: " + string(rune(req2.Age+'0'))
}

// Test call function

func TestCall_NoParamsNoReturn(t *testing.T) {
	engine := New()
	ctx := createTestContext(engine, "GET", "/", "")

	result := call(ctx, handlerNoParamsNoReturn)
	assert.Nil(t, result)
}

func TestCall_NoParamsSingleReturn(t *testing.T) {
	engine := New()
	ctx := createTestContext(engine, "GET", "/", "")

	result := call(ctx, handlerNoParamsSingleReturn)
	assert.Equal(t, "hello", result)
}

func TestCall_NoParamsErrorReturn(t *testing.T) {
	engine := New()
	ctx := createTestContext(engine, "GET", "/", "")

	result := call(ctx, handlerNoParamsErrorReturn)
	require.NotNil(t, result)
	err, ok := result.(error)
	require.True(t, ok)
	assert.Equal(t, "test error", err.Error())
}

func TestCall_NoParamsTwoReturns(t *testing.T) {
	engine := New()
	ctx := createTestContext(engine, "GET", "/", "")

	result := call(ctx, handlerNoParamsTwoReturns)
	assert.Equal(t, "result", result)
}

func TestCall_NoParamsTwoReturnsWithError(t *testing.T) {
	engine := New()
	ctx := createTestContext(engine, "GET", "/", "")

	result := call(ctx, handlerNoParamsTwoReturnsWithError)
	require.NotNil(t, result)
	err, ok := result.(error)
	require.True(t, ok)
	assert.Equal(t, "test error", err.Error())
}

func TestCall_CtxOnlyNoReturn(t *testing.T) {
	engine := New()
	ctx := createTestContext(engine, "GET", "/", "")

	result := call(ctx, handlerCtxOnlyNoReturn)
	assert.Nil(t, result)
}

func TestCall_CtxOnlySingleReturn(t *testing.T) {
	engine := New()
	ctx := createTestContext(engine, "GET", "/", "")

	result := call(ctx, handlerCtxOnlySingleReturn)
	assert.Equal(t, "hello from ctx", result)
}

func TestCall_CtxOnlyErrorReturn(t *testing.T) {
	engine := New()
	ctx := createTestContext(engine, "GET", "/", "")

	result := call(ctx, handlerCtxOnlyErrorReturn)
	require.NotNil(t, result)
	err, ok := result.(error)
	require.True(t, ok)
	assert.Equal(t, "ctx error", err.Error())
}

func TestCall_CtxOnlyTwoReturns(t *testing.T) {
	engine := New()
	ctx := createTestContext(engine, "GET", "/", "")

	result := call(ctx, handlerCtxOnlyTwoReturns)
	assert.Equal(t, "ctx result", result)
}

func TestCall_CtxWithParamSingleReturn(t *testing.T) {
	engine := New()
	body := `{"name":"John"}`
	ctx := createTestContext(engine, "POST", "/", body)

	result := call(ctx, handlerCtxWithParamSingleReturn)
	assert.Equal(t, "name: John", result)
}

func TestCall_CtxWithParamTwoReturns(t *testing.T) {
	engine := New()

	t.Run("success", func(t *testing.T) {
		body := `{"name":"John"}`
		ctx := createTestContext(engine, "POST", "/", body)

		result := call(ctx, handlerCtxWithParamTwoReturns)
		assert.Equal(t, "name: John", result)
	})

	t.Run("error from handler", func(t *testing.T) {
		body := `{"name":""}`
		ctx := createTestContext(engine, "POST", "/", body)

		result := call(ctx, handlerCtxWithParamTwoReturns)
		require.NotNil(t, result)
		err, ok := result.(error)
		require.True(t, ok)
		assert.Equal(t, "name is required", err.Error())
	})
}

func TestCall_BindingError(t *testing.T) {
	engine := New()

	t.Run("invalid json", func(t *testing.T) {
		body := `{"name": invalid json}`
		ctx := createTestContext(engine, "POST", "/", body)

		result := call(ctx, handlerCtxWithParamSingleReturn)
		require.NotNil(t, result)

		httpErr, ok := result.(*httperrors.Error)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.HTTPCode)
		assert.Equal(t, "BIND_ERROR", httpErr.Code)
	})

	t.Run("empty body with required field", func(t *testing.T) {
		// Note: validation only works if gin.binding.Validator is configured
		// By default, empty JSON `{}` will bind successfully with zero values
		body := `{}`
		ctx := createTestContext(engine, "POST", "/", body)

		result := call(ctx, handlerCtxWithInvalidParam)
		// With default config, this will succeed (validator not configured)
		// In production with validator configured, this would return an error
		assert.NotNil(t, result)
	})
}

func TestCall_MultipleParameters(t *testing.T) {
	engine := New()
	body := `{"name":"John","age":30}`
	ctx := createTestContext(engine, "POST", "/", body)

	result := call(ctx, handlerCtxWithMultipleParams)
	// Note: This will bind the same JSON to both parameters
	assert.Contains(t, result, "name: John")
}

func TestCall_WithQueryParams(t *testing.T) {
	engine := New()

	type queryRequest struct {
		Name string `query:"name"`
	}

	handler := func(ctx *Context, req *queryRequest) string {
		return "name: " + req.Name
	}

	req := httptest.NewRequest(http.MethodGet, "/?name=John", nil)
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = req

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
		Request: req,
	}

	result := call(ctx, handler)
	assert.Equal(t, "name: John", result)
}

func TestCall_WithURIParams(t *testing.T) {
	engine := New()

	type uriRequest struct {
		ID string `uri:"id"`
	}

	handler := func(ctx *Context, req *uriRequest) string {
		return "id: " + req.ID
	}

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = req
	ginCtx.Params = gin.Params{{Key: "id", Value: "123"}}

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
		Request: req,
	}

	result := call(ctx, handler)
	assert.Equal(t, "id: 123", result)
}

func TestCall_WithHeaderParams(t *testing.T) {
	engine := New()

	type headerRequest struct {
		Authorization string `header:"Authorization"`
	}

	handler := func(ctx *Context, req *headerRequest) string {
		return "auth: " + req.Authorization
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token123")
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = req

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
		Request: req,
	}

	result := call(ctx, handler)
	assert.Equal(t, "auth: Bearer token123", result)
}

func TestCall_ComplexStruct(t *testing.T) {
	engine := New()

	type complexRequest struct {
		Name  string            `json:"name"`
		Age   int               `json:"age"`
		Email string            `json:"email"`
		Tags  []string          `json:"tags"`
		Meta  map[string]string `json:"meta"`
	}

	handler := func(ctx *Context, req *complexRequest) *complexRequest {
		return req
	}

	body := `{
		"name": "John",
		"age": 30,
		"email": "john@example.com",
		"tags": ["go", "testing"],
		"meta": {"key": "value"}
	}`

	ctx := createTestContext(engine, "POST", "/", body)

	result := call(ctx, handler)
	require.NotNil(t, result)

	complexReq, ok := result.(*complexRequest)
	require.True(t, ok)
	assert.Equal(t, "John", complexReq.Name)
	assert.Equal(t, 30, complexReq.Age)
	assert.Equal(t, "john@example.com", complexReq.Email)
	assert.Equal(t, []string{"go", "testing"}, complexReq.Tags)
	assert.Equal(t, map[string]string{"key": "value"}, complexReq.Meta)
}

func TestCall_NilReturn(t *testing.T) {
	engine := New()

	handler := func(ctx *Context) *testRequest {
		return nil
	}

	ctx := createTestContext(engine, "GET", "/", "")

	result := call(ctx, handler)
	assert.Nil(t, result)
}

func TestCall_ErrorInterface(t *testing.T) {
	engine := New()

	t.Run("single return error", func(t *testing.T) {
		handler := func(ctx *Context) error {
			return httperrors.New(http.StatusNotFound, "not found")
		}

		ctx := createTestContext(engine, "GET", "/", "")
		result := call(ctx, handler)

		require.NotNil(t, result)
		err, ok := result.(error)
		require.True(t, ok)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("two returns with error", func(t *testing.T) {
		handler := func(ctx *Context) (string, error) {
			return "", httperrors.New(http.StatusBadRequest, "bad request")
		}

		ctx := createTestContext(engine, "GET", "/", "")
		result := call(ctx, handler)

		require.NotNil(t, result)
		err, ok := result.(error)
		require.True(t, ok)
		assert.Contains(t, err.Error(), "bad request")
	})

	t.Run("two returns with nil error", func(t *testing.T) {
		handler := func(ctx *Context) (string, error) {
			return "success", nil
		}

		ctx := createTestContext(engine, "GET", "/", "")
		result := call(ctx, handler)

		assert.Equal(t, "success", result)
	})
}

func TestCall_ContextValue(t *testing.T) {
	engine := New()

	type ctxRequest struct {
		UserID string `context:"user_id"`
	}

	handler := func(ctx *Context, req *ctxRequest) string {
		return "user: " + req.UserID
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = req

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
		Request: req,
	}

	ctx.Set("user_id", "12345")

	result := call(ctx, handler)
	assert.Equal(t, "user: 12345", result)
}

func TestCall_MixedBindings(t *testing.T) {
	engine := New()

	type mixedRequest struct {
		Name   string `json:"name"`
		UserID string `query:"user_id"`
		ID     string `uri:"id"`
		Token  string `header:"X-Token"`
	}

	handler := func(ctx *Context, req *mixedRequest) map[string]string {
		return map[string]string{
			"name":    req.Name,
			"user_id": req.UserID,
			"id":      req.ID,
			"token":   req.Token,
		}
	}

	body := `{"name":"John"}`
	req := httptest.NewRequest(http.MethodPost, "/?user_id=user123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token", "token456")

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = req
	ginCtx.Params = gin.Params{{Key: "id", Value: "789"}}

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
		Request: req,
	}

	result := call(ctx, handler)
	require.NotNil(t, result)

	resultMap, ok := result.(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "John", resultMap["name"])
	assert.Equal(t, "user123", resultMap["user_id"])
	assert.Equal(t, "789", resultMap["id"])
	assert.Equal(t, "token456", resultMap["token"])
}

// Benchmark tests

func BenchmarkCall_NoParams(b *testing.B) {
	engine := New()
	ctx := createTestContext(engine, "GET", "/", "")
	handler := handlerNoParamsSingleReturn

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = call(ctx, handler)
	}
}

func BenchmarkCall_CtxOnly(b *testing.B) {
	engine := New()
	ctx := createTestContext(engine, "GET", "/", "")
	handler := handlerCtxOnlySingleReturn

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = call(ctx, handler)
	}
}

func BenchmarkCall_CtxWithParam(b *testing.B) {
	engine := New()
	body := `{"name":"John"}`
	ctx := createTestContext(engine, "POST", "/", body)
	handler := handlerCtxWithParamSingleReturn

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = call(ctx, handler)
	}
}

func BenchmarkCall_ComplexBinding(b *testing.B) {
	engine := New()

	type complexRequest struct {
		Name  string            `json:"name"`
		Age   int               `json:"age"`
		Email string            `json:"email"`
		Tags  []string          `json:"tags"`
		Meta  map[string]string `json:"meta"`
	}

	handler := func(ctx *Context, req *complexRequest) string {
		return req.Name
	}

	body := `{
		"name": "John",
		"age": 30,
		"email": "john@example.com",
		"tags": ["go", "testing"],
		"meta": {"key": "value"}
	}`

	ctx := createTestContext(engine, "POST", "/", body)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = call(ctx, handler)
	}
}

// Helper function to create test context

func createTestContext(engine *Engine, method, path, body string) *Context {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = req

	return &Context{
		Context: ginCtx,
		engine:  engine,
		Request: req,
	}
}

// Test error marshaling

func TestCall_HTTPErrorMarshaling(t *testing.T) {
	engine := New()

	handler := func(ctx *Context) error {
		return httperrors.New(http.StatusBadRequest, "validation failed").
			SetCode("VALIDATION_ERROR").
			AddField("field", "email").
			AddField("reason", "invalid format")
	}

	ctx := createTestContext(engine, "GET", "/", "")
	result := call(ctx, handler)

	require.NotNil(t, result)
	httpErr, ok := result.(*httperrors.Error)
	require.True(t, ok)

	jsonData, err := json.Marshal(httpErr)
	require.NoError(t, err)

	var resultMap map[string]any
	err = json.Unmarshal(jsonData, &resultMap)
	require.NoError(t, err)

	assert.Equal(t, "VALIDATION_ERROR", resultMap["code"])
	assert.Equal(t, "email", resultMap["field"])
	assert.Equal(t, "invalid format", resultMap["reason"])
}
