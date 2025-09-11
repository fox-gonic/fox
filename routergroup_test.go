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
	return
}

func boo(c *fox.Context) (res any, err error) {
	res = "boo"
	return
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
		router.GET("too-many-values", func(c *fox.Context) (res any, other any, err error) { return })
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
