package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorHandlingRoutes(t *testing.T) {
	router := newRouter()
	req := httptest.NewRequest(http.MethodGet, "/error/http", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), `"code":"BAD_REQUEST"`)
}
