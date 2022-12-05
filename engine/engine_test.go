package engine_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fox-gonic/fox/engine"
	"github.com/fox-gonic/fox/errors"
	"github.com/fox-gonic/fox/testhelper"
)

type Foo struct {
	A string
	B string
}

func MiddlewareFailed(c *engine.Context) (res interface{}, err error) {
	c.Logger.Info("MiddlewareFailed")
	res = "Middleware"
	err = errors.ErrInvalidArguments
	return
}

func MiddlewareSuccess(c *engine.Context) (res interface{}, err error) {
	c.Logger.Info("MiddlewareSuccess")
	return
}

type TestInput struct {
	Param string `uri:"param"`
	Query string `query:"query"`

	Body string `json:"body"`
}

func HandleSuccess(c *engine.Context, in TestInput) (res interface{}, err error) {
	res = in
	return
}

func HandleFailed(c *engine.Context, in *TestInput) (res interface{}, err error) {
	err = &errors.Error{
		HTTPCode: http.StatusBadRequest,
		Code:     errors.ErrInvalidArguments.Code,
		Message: map[string]interface{}{
			"param": "invalid param " + in.Param,
		},
	}
	return
}

func Ping(c *engine.Context) (res interface{}, err error) {
	c.Logger.Info("PingHandler")
	res = Foo{"a", "b"}
	return
}

func TestEngine(t *testing.T) {

	assert := assert.New(t)

	router := engine.New()
	router.GET("ping", MiddlewareFailed, Ping)
	router.GET("ping2", MiddlewareSuccess, Ping)

	router.POST("/handle/:param/success", HandleSuccess)
	router.POST("/handle/:param/failed", HandleFailed)

	w := testhelper.PerformRequest(router, "GET", "/ping", nil)
	assert.Equal(http.StatusBadRequest, w.Code)

	body := w.Body.String()
	assert.Equal(body, `{"code":"INVALID_ARGUMENTS","message":null}`)

	w = testhelper.PerformRequest(router, "GET", "/ping2", nil)
	assert.Equal(http.StatusOK, w.Code)

	var response Foo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(err)

	assert.Equal(response.A, "a")
	assert.Equal(response.B, "b")

	jsonData := map[string]string{"body": "fromBody"}
	w = testhelper.PerformRequestJSON(router, http.MethodPost, "/handle/hello/success?query=world", jsonData)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(`{"Param":"hello","Query":"world","body":"fromBody"}`, w.Body.String())

	w = testhelper.PerformRequestJSON(router, http.MethodPost, "/handle/hello/failed?query=world", jsonData)
	assert.Equal(http.StatusBadRequest, w.Code)
	assert.Equal(`{"code":"INVALID_ARGUMENTS","message":{"param":"invalid param hello"}}`, w.Body.String())
}
