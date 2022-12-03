package testhelper

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
)

// PerformRequest router test
func PerformRequest(r http.Handler, method, path string, header http.Header, body ...io.Reader) *httptest.ResponseRecorder {
	var data io.Reader

	if len(body) > 0 {
		data = body[0]
	}

	req, _ := http.NewRequest(method, path, data)
	req.Header = header
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// PerformRequestJSON 使用 interface{} 作为json参数，在函数内部去json Marshal
func PerformRequestJSON(r http.Handler, method, path string, data interface{}, header ...http.Header) *httptest.ResponseRecorder {
	b, _ := json.Marshal(data)
	req, _ := http.NewRequest(method, path, bytes.NewReader(b))

	h := http.Header{}
	if len(header) > 0 {
		h = header[0]
	}

	h.Set("Content-Type", "application/json")
	req.Header = h

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
