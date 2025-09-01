package fox_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

func TestRouterGroupCompatibleWithGinHandler(t *testing.T) {
	middleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("foo", "bar")
			c.Next()
		}
	}
	assert := assert.New(t)

	router := fox.New()
	api := router.Group("/api")
	api.Use(middleware())
	api.GET("foo", func(c *gin.Context) {
		foo, _ := c.Get("foo")
		c.String(http.StatusOK, foo.(string))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("bar", w.Body.String())
}
