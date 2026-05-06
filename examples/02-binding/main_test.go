package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBindingRoutes(t *testing.T) {
	router := newRouter()
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{
		"username":"alice",
		"email":"alice@example.com",
		"password":"secret1",
		"age":25
	}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"username":"alice"`)
}
