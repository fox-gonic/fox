// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
// https://github.com/gin-gonic/gin

package fox

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

// ResponseWriter with a http.ResponseWriter wrapper
type ResponseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

func (w *ResponseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = noWritten
	w.status = defaultStatus
}

// WriteHeader sends an HTTP response header with the provided
// status code.
func (w *ResponseWriter) WriteHeader(statusCode int) {
	if statusCode > 0 && w.status != statusCode {
		if w.Written() {
			// debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
		}
		w.status = statusCode
	}
}

// WriteHeaderNow forces to write the http header (status code + headers).
func (w *ResponseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *ResponseWriter) Write(data []byte) (n int, err error) {
	w.WriteHeaderNow()
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	return
}

// WriteString writes the string into the response body.
func (w *ResponseWriter) WriteString(s string) (n int, err error) {
	w.WriteHeaderNow()
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	return
}

// Status returns the HTTP response status code of the current request.
func (w *ResponseWriter) Status() int {
	return w.status
}

// Size returns the number of bytes already written into the response http body.
// See Written()
func (w *ResponseWriter) Size() int {
	return w.size
}

// Written returns true if the response body was already written.
func (w *ResponseWriter) Written() bool {
	return w.size != noWritten
}

// Hijack implements the http.Hijacker interface.
func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.size < 0 {
		w.size = 0
	}
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotifier interface.
func (w *ResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush implements the http.Flusher interface.
func (w *ResponseWriter) Flush() {
	w.WriteHeaderNow()
	w.ResponseWriter.(http.Flusher).Flush()
}

// Pusher implements the http.Pusher interface.
func (w *ResponseWriter) Pusher() (pusher http.Pusher) {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}
