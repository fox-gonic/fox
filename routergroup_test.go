package fox_test

import (
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

func tooManyValues(c *fox.Context) (res any, other any, err error) {
	res = "too many values"
	return
}

func TestRouterGroup(t *testing.T) {

	assert := assert.New(t)

	router := fox.New()
	api := router.Group("/api")

	api.GET("foo", foo)

	api.GET("boo", boo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/foo", nil)
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("foo", w.Body.String())

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/boo", nil)
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("boo", w.Body.String())
}

func TestRouterGroupPanic(t *testing.T) {
	assert := assert.New(t)

	defer func() {
		if r := recover(); r != nil {
			assert.Contains(r.(string), "only support handler func returns max is two values")
		} else {
			t.Error("expected panic did not occur")
		}
	}()

	router := fox.New()
	api := router.Group("/api")
	api.GET("too-many-values", tooManyValues)
}
