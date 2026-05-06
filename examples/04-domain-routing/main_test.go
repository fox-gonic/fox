package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDomainRoutes(t *testing.T) {
	router := newRouter()
	req := httptest.NewRequest(http.MethodGet, "http://api.example.com/status", nil)
	req.Host = "api.example.com"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"status":"API service running"`)
}
