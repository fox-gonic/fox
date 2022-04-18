// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
// https://github.com/gin-gonic/gin

package fox

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	_ http.ResponseWriter = &ResponseWriter{}
	_ http.Hijacker       = &ResponseWriter{}
	_ http.Flusher        = &ResponseWriter{}
	_ http.CloseNotifier  = &ResponseWriter{}
)

func init() {
	SetMode(TestMode)
}

func TestResponseWriterReset(t *testing.T) {
	testWriter := httptest.NewRecorder()
	writer := &ResponseWriter{}
	var w = writer

	writer.reset(testWriter)
	assert.Equal(t, -1, writer.size)
	assert.Equal(t, http.StatusOK, writer.status)
	assert.Equal(t, testWriter, writer.ResponseWriter)
	assert.Equal(t, -1, w.Size())
	assert.Equal(t, http.StatusOK, w.Status())
	assert.False(t, w.Written())
}

func TestResponseWriterWriteHeader(t *testing.T) {
	testWriter := httptest.NewRecorder()
	writer := &ResponseWriter{}
	writer.reset(testWriter)
	w := writer

	w.WriteHeader(http.StatusMultipleChoices)
	assert.False(t, w.Written())
	assert.Equal(t, http.StatusMultipleChoices, w.Status())
	assert.NotEqual(t, http.StatusMultipleChoices, testWriter.Code)

	w.WriteHeader(-1)
	assert.Equal(t, http.StatusMultipleChoices, w.Status())
}

func TestResponseWriterWriteHeadersNow(t *testing.T) {
	testWriter := httptest.NewRecorder()
	writer := &ResponseWriter{}
	writer.reset(testWriter)
	w := writer

	w.WriteHeader(http.StatusMultipleChoices)
	w.WriteHeaderNow()

	assert.True(t, w.Written())
	assert.Equal(t, 0, w.Size())
	assert.Equal(t, http.StatusMultipleChoices, testWriter.Code)

	writer.size = 10
	w.WriteHeaderNow()
	assert.Equal(t, 10, w.Size())
}

func TestResponseWriterWrite(t *testing.T) {
	testWriter := httptest.NewRecorder()
	writer := &ResponseWriter{}
	writer.reset(testWriter)
	w := writer

	n, err := w.Write([]byte("hola"))
	assert.Equal(t, 4, n)
	assert.Equal(t, 4, w.Size())
	assert.Equal(t, http.StatusOK, w.Status())
	assert.Equal(t, http.StatusOK, testWriter.Code)
	assert.Equal(t, "hola", testWriter.Body.String())
	assert.NoError(t, err)

	n, err = w.Write([]byte(" adios"))
	assert.Equal(t, 6, n)
	assert.Equal(t, 10, w.Size())
	assert.Equal(t, "hola adios", testWriter.Body.String())
	assert.NoError(t, err)
}

func TestResponseWriterHijack(t *testing.T) {
	testWriter := httptest.NewRecorder()
	writer := &ResponseWriter{}
	writer.reset(testWriter)
	w := writer

	assert.Panics(t, func() {
		_, _, err := w.Hijack()
		assert.NoError(t, err)
	})
	assert.True(t, w.Written())

	assert.Panics(t, func() {
		w.CloseNotify()
	})

	w.Flush()
}

func TestResponseWriterFlush(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &ResponseWriter{}
		writer.reset(w)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Flush()
	}))
	defer testServer.Close()

	// should return 500
	resp, err := http.Get(testServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
