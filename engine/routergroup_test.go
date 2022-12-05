package engine_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fox-gonic/fox/engine"
	"github.com/fox-gonic/fox/testhelper"
)

func foo(c *engine.Context) (res interface{}, err error) {
	res = "foo"
	return
}

func boo(c *engine.Context) (res interface{}, err error) {
	res = "boo"
	return
}

func TestRouterGroup(t *testing.T) {

	assert := assert.New(t)

	router := engine.New()
	api := router.Group("/api")

	api.GET("foo", foo)

	api.GET("boo", boo)

	w := testhelper.PerformRequest(router, "GET", "/api/foo", nil)
	assert.Equal(http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Equal(body, "foo")

	w = testhelper.PerformRequest(router, "GET", "/api/boo", nil)
	assert.Equal(http.StatusOK, w.Code)

	body = w.Body.String()
	assert.Equal(body, "boo")
}
