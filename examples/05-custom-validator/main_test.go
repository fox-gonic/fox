package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomValidatorRoutes(t *testing.T) {
	router := newRouter()
	req := httptest.NewRequest(http.MethodPost, "/validate-password", strings.NewReader(`{"password":"Strong1!"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "Password is strong!", w.Body.String())
}
