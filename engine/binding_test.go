package engine

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MixStruct struct {
	Page       int        `query:"page"`
	PageSize   int        `query:"page_size"`
	IDs        []int      `query:"ids[]"`
	Start      *time.Time `query:"start"         time_format:"unix"`
	Referer    string     `header:"referer"`
	XRequestID string     `header:"X-Request-Id"`
	Vary       []string   `header:"vary"`
	Name       string     `json:"name"`
	Content    *string    `json:"content"`
}

func TestBinding(t *testing.T) {
	var (
		obj        MixStruct
		url        = "/?page=1&page_size=30&ids[]=1&ids[]=2&ids[]=3&ids[]=4&ids[]=5&start=1669732749"
		referer    = "http://domain.name/posts"
		varyHeader = []string{"X-PJAX, X-PJAX-Container, Turbo-Visit, Turbo-Frame", "Accept-Encoding, Accept, X-Requested-With"}
		XRequestID = "l4dCIsjENo3QsCoX"
	)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Referer", referer)
	req.Header.Set("X-Request-Id", XRequestID)
	req.Header.Add("vary", varyHeader[0])
	req.Header.Add("vary", varyHeader[1])

	ctx := &Context{
		Context: &gin.Context{
			Request: req,
		},
	}

	err := bind(ctx, obj)
	assert.Equal(t, ErrBindNonPointerValue, err)

	err = bind(ctx, &obj)
	assert.NoError(t, err)
	assert.Equal(t, 1, obj.Page)
	assert.Equal(t, 30, obj.PageSize)
	assert.Equal(t, referer, obj.Referer)
	assert.Equal(t, XRequestID, obj.XRequestID)
	assert.Equal(t, varyHeader, obj.Vary)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, obj.IDs)
	assert.NotZero(t, obj.Start)
}

func TestBindingJSON(t *testing.T) {
	var (
		obj        MixStruct
		url        = "/?page=1&page_size=30&ids[]=1&ids[]=2&ids[]=3&ids[]=4&ids[]=5&start=1669732749"
		referer    = "http://domain.name/posts"
		varyHeader = []string{"X-PJAX, X-PJAX-Container, Turbo-Visit, Turbo-Frame", "Accept-Encoding, Accept, X-Requested-With"}
		XRequestID = "l4dCIsjENo3QsCoX"
	)

	req := requestWithBody(http.MethodPost, url, `{"name": "Binder"}`)
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Referer", referer)
	req.Header.Set("X-Request-Id", XRequestID)
	req.Header.Add("vary", varyHeader[0])
	req.Header.Add("vary", varyHeader[1])

	ctx := &Context{
		Context: &gin.Context{
			Request: req,
		},
	}

	err := bind(ctx, &obj)
	assert.NoError(t, err)
	assert.Equal(t, 1, obj.Page)
	assert.Equal(t, 30, obj.PageSize)
	assert.Equal(t, referer, obj.Referer)
	assert.Equal(t, XRequestID, obj.XRequestID)
	assert.Equal(t, varyHeader, obj.Vary)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, obj.IDs)
	assert.Equal(t, "Binder", obj.Name)
	assert.Nil(t, obj.Content)
	assert.NotZero(t, obj.Start)
}

func requestWithBody(method, path, body string) (req *http.Request) {
	req, _ = http.NewRequest(method, path, bytes.NewBufferString(body))
	return
}
