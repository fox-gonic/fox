package fox

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fox-gonic/fox/httperrors"
	"github.com/fox-gonic/fox/render"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test StatusCoder interface

type testStatusCoder struct {
	code int
}

func (t *testStatusCoder) StatusCode() int {
	return t.code
}

func (t *testStatusCoder) Error() string {
	return "test error with status code"
}

// Test render.Render interface

type testRender struct {
	data string
}

func (t *testRender) Render(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(t.data))
	return err
}

func (t *testRender) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
}

// Test json.Marshaler interface

type testJSONMarshaler struct {
	Message string
	Code    int
}

func (t *testJSONMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"message": t.Message,
		"code":    t.Code,
	})
}

func (t *testJSONMarshaler) Error() string {
	return t.Message
}

// Test renderError function

func TestRenderError_Nil(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	ctx.renderError(nil)

	// Should do nothing for nil error
	// Note: Gin sets 200 by default, but no body written
	assert.Empty(t, w.Body.String())
}

func TestRenderError_CustomRenderErrorFunc(t *testing.T) {
	engine := New()
	engine.RenderErrorFunc = func(c *Context, err error) {
		c.JSON(http.StatusTeapot, gin.H{"custom": err.Error()})
	}

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	ctx.renderError(errors.New("test error"))

	assert.Equal(t, http.StatusTeapot, w.Code)
	assert.Contains(t, w.Body.String(), "custom")
	assert.Contains(t, w.Body.String(), "test error")
}

func TestRenderError_StatusCoder(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	err := &testStatusCoder{code: http.StatusNotFound}
	ctx.renderError(err)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "test error with status code")
}

func TestRenderError_DefaultStatusCode(t *testing.T) {
	engine := New()
	engine.DefaultRenderErrorStatusCode = http.StatusBadRequest

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	ctx.renderError(errors.New("test error"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "test error")
}

func TestRenderError_JSONMarshaler(t *testing.T) {
	engine := New()
	engine.DefaultRenderErrorStatusCode = http.StatusInternalServerError

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	err := &testJSONMarshaler{
		Message: "test json error",
		Code:    1001,
	}

	ctx.renderError(err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "test json error")
	assert.Contains(t, w.Body.String(), "1001")
}

func TestRenderError_PlainError(t *testing.T) {
	engine := New()
	engine.DefaultRenderErrorStatusCode = http.StatusInternalServerError

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	ctx.renderError(errors.New("plain error"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "plain error", w.Body.String())
}

// testRenderError implements both error and render.Render interfaces
type testRenderError struct {
	message string
	code    int
}

func (e *testRenderError) Error() string {
	return e.message
}

func (e *testRenderError) StatusCode() int {
	return e.code
}

func (e *testRenderError) Render(w http.ResponseWriter) error {
	_, err := w.Write([]byte("rendered: " + e.message))
	return err
}

func (e *testRenderError) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}

func TestRenderError_RenderInterface(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	err := &testRenderError{
		message: "test render error",
		code:    http.StatusBadRequest,
	}

	ctx.renderError(err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "rendered: test render error", w.Body.String())
}

func TestRenderError_HTTPError(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	httpErr := httperrors.New(http.StatusBadRequest, "validation failed").
		SetCode("VALIDATION_ERROR").
		AddField("field", "email")

	ctx.renderError(httpErr)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var result map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", result["code"])
	assert.Equal(t, "email", result["field"])
}

// Test render function

func TestRender_Nil(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	ctx.render(nil)

	// Should do nothing for nil result
	// Note: Gin sets 200 by default, but no body written
	assert.Empty(t, w.Body.String())
	assert.False(t, ctx.IsAborted())
}

func TestRender_Error(t *testing.T) {
	engine := New()
	engine.DefaultRenderErrorStatusCode = http.StatusInternalServerError

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	ctx.render(errors.New("test error"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "test error")
	assert.True(t, ctx.IsAborted())
}

func TestRender_String(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	ctx.render("hello world")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello world", w.Body.String())
	assert.True(t, ctx.IsAborted())
}

func TestRender_Redirect(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = req

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	redirect := render.Redirect{
		Code:     http.StatusFound,
		Location: "https://example.com",
	}

	ctx.render(redirect)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))
	assert.True(t, ctx.IsAborted())
}

func TestRender_RenderInterface(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	customRender := &testRender{data: "custom data"}
	ctx.render(customRender)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom data", w.Body.String())
	assert.True(t, ctx.IsAborted())
}

func TestRender_JSON(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	data := map[string]any{
		"name": "John",
		"age":  30,
	}

	ctx.render(data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var result map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "John", result["name"])
	assert.InDelta(t, 30, result["age"], 0.001) // JSON numbers are float64
	assert.True(t, ctx.IsAborted())
}

func TestRender_Struct(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	user := User{
		Name:  "Jane",
		Email: "jane@example.com",
	}

	ctx.render(user)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var result User
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "Jane", result.Name)
	assert.Equal(t, "jane@example.com", result.Email)
	assert.True(t, ctx.IsAborted())
}

func TestRender_Array(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	data := []string{"a", "b", "c"}
	ctx.render(data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var result []string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, result)
	assert.True(t, ctx.IsAborted())
}

func TestRender_Number(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	ctx.render(42)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	assert.Equal(t, "42", w.Body.String())
	assert.True(t, ctx.IsAborted())
}

func TestRender_Boolean(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	ctx.render(true)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	assert.Equal(t, "true", w.Body.String())
	assert.True(t, ctx.IsAborted())
}

// Test edge cases

func TestRender_EmptyString(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	ctx.render("")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
	assert.True(t, ctx.IsAborted())
}

func TestRender_EmptyStruct(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	type Empty struct{}
	ctx.render(Empty{})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{}", w.Body.String())
	assert.True(t, ctx.IsAborted())
}

func TestRender_Pointer(t *testing.T) {
	engine := New()
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	ctx := &Context{
		Context: ginCtx,
		engine:  engine,
	}

	type User struct {
		Name string `json:"name"`
	}

	user := &User{Name: "Bob"}
	ctx.render(user)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Bob")
	assert.True(t, ctx.IsAborted())
}

// Benchmark tests

func BenchmarkRenderError_Plain(b *testing.B) {
	engine := New()
	engine.DefaultRenderErrorStatusCode = http.StatusInternalServerError
	err := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ctx := &Context{Context: ginCtx, engine: engine}
		ctx.renderError(err)
	}
}

func BenchmarkRenderError_HTTPError(b *testing.B) {
	engine := New()
	httpErr := httperrors.New(http.StatusBadRequest, "validation failed")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ctx := &Context{Context: ginCtx, engine: engine}
		ctx.renderError(httpErr)
	}
}

func BenchmarkRender_String(b *testing.B) {
	engine := New()
	data := "hello world"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ctx := &Context{Context: ginCtx, engine: engine}
		ctx.render(data)
	}
}

func BenchmarkRender_JSON(b *testing.B) {
	engine := New()
	data := map[string]any{
		"name": "John",
		"age":  30,
		"tags": []string{"go", "testing"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ctx := &Context{Context: ginCtx, engine: engine}
		ctx.render(data)
	}
}

func BenchmarkRender_Struct(b *testing.B) {
	engine := New()
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}
	user := User{Name: "Jane", Email: "jane@example.com", Age: 25}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ctx := &Context{Context: ginCtx, engine: engine}
		ctx.render(user)
	}
}
