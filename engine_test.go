package fox_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fox-gonic/fox"
	"github.com/fox-gonic/fox/httperrors"
)

type Foo struct {
	A string
	B string
}

func MiddlewareFailed(c *fox.Context) (res interface{}, err error) {
	c.Logger.Info("MiddlewareFailed")
	res = "Middleware"
	err = httperrors.ErrInvalidArguments
	return
}

func MiddlewareSuccess(c *fox.Context) (res interface{}, err error) {
	c.Logger.Info("MiddlewareSuccess")
	return
}

type TestInput struct {
	Param string `uri:"param"`
	Query string `query:"query"`

	Body string `json:"body"`
}

func HandleSuccess(c *fox.Context, in TestInput) (res interface{}, err error) {
	res = in
	return
}

func HandleFailed(c *fox.Context, in *TestInput) (res interface{}, err error) {
	err = &httperrors.Error{
		HTTPCode: http.StatusBadRequest,
		Code:     httperrors.ErrInvalidArguments.Code,
		Message: map[string]interface{}{
			"param": "invalid param " + in.Param,
		},
	}
	return
}

func Ping(c *fox.Context) (res interface{}, err error) {
	c.Logger.Info("PingHandler")
	res = Foo{"a", "b"}
	return
}

func TestEngine(t *testing.T) {

	assert := assert.New(t)

	router := fox.New()
	router.GET("ping", MiddlewareFailed, Ping)
	router.GET("ping2", MiddlewareSuccess, Ping)

	router.POST("/handle/:param/success", HandleSuccess)
	router.POST("/handle/:param/failed", HandleFailed)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusBadRequest, w.Code)

	body := w.Body.String()
	assert.Equal(`{"code":"INVALID_ARGUMENTS","message":{"error":"invalid arguments"}}`, body)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/ping2", nil)
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)

	var response Foo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(err)

	assert.Equal(response.A, "a")
	assert.Equal(response.B, "b")

	jsonData := map[string]string{"body": "fromBody"}
	data, _ := json.Marshal(jsonData)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/handle/hello/success?query=world", bytes.NewReader(data))
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(`{"Param":"hello","Query":"world","body":"fromBody"}`, w.Body.String())

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/handle/hello/failed?query=world", bytes.NewReader(data))
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusBadRequest, w.Code)
	assert.Equal(`{"code":"INVALID_ARGUMENTS","message":{"param":"invalid param hello"}}`, w.Body.String())
}
