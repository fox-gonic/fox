package engine_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fox-gonic/fox/engine"
	"github.com/fox-gonic/fox/testhelper"
)

func TestRouterGroup(t *testing.T) {

	assert := assert.New(t)

	router := engine.New()
	api := router.Group("/api")

	api.GET("foo", func(c *engine.Context) (res interface{}, err error) {
		res = "foo"
		return
	})

	api.GET("boo", func(c *engine.Context) (res interface{}, err error) {
		res = "boo"
		return
	})

	w := testhelper.PerformRequest(router, "GET", "/api/foo", nil)
	assert.Equal(http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Equal(body, "foo")

	w = testhelper.PerformRequest(router, "GET", "/api/boo", nil)
	assert.Equal(http.StatusOK, w.Code)

	body = w.Body.String()
	assert.Equal(body, "boo")
}
