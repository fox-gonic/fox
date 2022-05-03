package render

import "net/http"

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

func writeHeaderCode(w http.ResponseWriter, code int) {
	if code >= 100 && code <= 999 {
		w.WriteHeader(code)
	}
}

// writeHeaders writes custom Header.
func writeHeaders(w http.ResponseWriter, headers map[string]string) {
	header := w.Header()
	for k, v := range headers {
		header.Set(k, v)
	}
}
