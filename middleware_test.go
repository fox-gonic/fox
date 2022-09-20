// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package fox

import (
	"encoding/xml"
	"errors"
	"net/http"
	"testing"

	"github.com/fox-gonic/fox/render"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareGeneralCase(t *testing.T) {
	signature := ""
	router := New()
	router.Use(func(c *Context) {
		signature += "A"
		c.Next()
		signature += "B"
	})
	router.Use(func(c *Context) {
		signature += "C"
	})
	router.GET("/", func(c *Context) {
		signature += "D"
	})
	router.NotFound(func(c *Context) {
		signature += " X "
	})
	router.NoMethod(func(c *Context) {
		signature += " XX "
	})

	w := PerformRequest(router, "GET", "/", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ACDB", signature)

	signature = ""
	w = PerformRequest(router, "GET", "/not_found", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "AC X B", signature)

	signature = ""
	w = PerformRequest(router, "POST", "/", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, "AC XX B", signature)
}

func TestMiddlewareNotFound(t *testing.T) {
	signature := ""
	router := New()
	router.Use(func(c *Context) {
		signature += "A"
		c.Next()
		signature += "B"
	})
	router.Use(func(c *Context) {
		signature += "C"
		c.Next()
		c.Next()
		c.Next()
		c.Next()
		signature += "D"
	})
	router.NotFound(func(c *Context) {
		signature += "E"
		c.Next()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.Next()
		signature += "H"
	})
	router.NoMethod(func(c *Context) {
		signature += " X "
	})

	w := PerformRequest(router, "GET", "/", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "ACEGHFDB", signature)
}

func TestMiddlewareNoMethodEnabled(t *testing.T) {
	signature := ""
	router := New()
	router.HandleMethodNotAllowed = true
	router.Use(func(c *Context) {
		signature += "A"
		c.Next()
		signature += "B"
	})
	router.Use(func(c *Context) {
		signature += "C"
		c.Next()
		signature += "D"
	})
	router.NoMethod(func(c *Context) {
		signature += "E"
		c.Next()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.Next()
		signature += "H"
	})
	router.NotFound(func(c *Context) {
		signature += " X "
	})
	router.POST("/", func(c *Context) {
		signature += " XX "
	})
	w := PerformRequest(router, "GET", "/", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, "ACEGHFDB", signature)
}

func TestMiddlewareNoMethodDisabled(t *testing.T) {
	signature := ""
	router := New()

	// NoMethod disabled
	router.HandleMethodNotAllowed = false

	router.Use(func(c *Context) {
		signature += "A"
		c.Next()
		signature += "B"
	})
	router.Use(func(c *Context) {
		signature += "C"
		c.Next()
		signature += "D"
	})
	router.NoMethod(func(c *Context) {
		signature += "E"
		c.Next()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.Next()
		signature += "H"
	})
	router.NotFound(func(c *Context) {
		signature += " X "
	})
	router.POST("/", func(c *Context) {
		signature += " XX "
	})

	w := PerformRequest(router, "GET", "/", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "AC X DB", signature)
}

// TestFailHandlersChain - ensure that Fail interrupt used middleware in fifo order as
// as well as Abort
func TestMiddlewareFailHandlersChain(t *testing.T) {
	signature := ""
	router := New()
	router.Use(func(c *Context) (interface{}, error) {
		signature += "A"
		return nil, &Error{
			Status: http.StatusInternalServerError,
			Err:    errors.New("foo"),
		}
	})
	router.Use(func(c *Context) {
		signature += "B"
		c.Next()
		signature += "C"
	})
	w := PerformRequest(router, "GET", "/", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "A", signature)
}

func TestMiddlewareWrite(t *testing.T) {
	router := New()
	router.Use(func(c *Context) (string, error) {
		return "hola\n", nil
	})
	router.Use(func(c *Context) (interface{}, error) {
		data := struct {
			XMLName xml.Name `xml:"map"`
			Foo     string   `xml:"foo"`
		}{Foo: "bar"}
		return render.XML{Data: data}, nil
	})
	router.Use(func(c *Context) (interface{}, error) {
		data := map[string]any{"foo": "bar"}
		return render.JSON{Data: data}, nil
	})
	router.GET("/", func(c *Context) (interface{}, error) {
		data := map[string]any{"foo": "bar"}
		return render.JSON{Data: data}, nil
	})

	w := PerformRequest(router, "GET", "/", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hola\n<map><foo>bar</foo></map>{\"foo\":\"bar\"}{\"foo\":\"bar\"}", w.Body.String())

	router = New()
	router.DefaultContentType = MIMEPlain
	router.Use(func(c *Context) (string, error) {
		return "hola\n", &Error{
			Status: http.StatusBadRequest,
		}
	})
	w = PerformRequest(router, "GET", "/", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, http.StatusText(http.StatusBadRequest), w.Body.String())
}
