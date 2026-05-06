package fox

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func registeredRouteHandler(_ *Context) string {
	return "ok"
}

func TestHandlerRoutesReturnsOriginalHandlerMetadata(t *testing.T) {
	engine := New()
	engine.GET("/health", registeredRouteHandler)

	routes := engine.HandlerRoutes()

	require.Len(t, routes, 1)
	require.Equal(t, "GET", routes[0].Method)
	require.Equal(t, "/health", routes[0].Path)
	require.Equal(t, reflect.TypeOf(registeredRouteHandler), routes[0].HandlerType)
	require.Contains(t, routes[0].HandlerName, "registeredRouteHandler")
}

func TestHandlerRoutesAreSortedAndCanBeDisabled(t *testing.T) {
	engine := New()
	engine.POST("/zeta", registeredRouteHandler)
	engine.GET("/alpha", registeredRouteHandler)
	engine.POST("/alpha", registeredRouteHandler)

	routes := engine.HandlerRoutes()
	require.Len(t, routes, 3)
	require.Equal(t, "GET", routes[0].Method)
	require.Equal(t, "/alpha", routes[0].Path)
	require.Equal(t, "POST", routes[1].Method)
	require.Equal(t, "/alpha", routes[1].Path)
	require.Equal(t, "/zeta", routes[2].Path)

	engine.DisableRouteRegistry()
	require.Empty(t, engine.HandlerRoutes())

	engine.GET("/later", registeredRouteHandler)
	require.Empty(t, engine.HandlerRoutes())
}

func TestRegisterHandlerRouteIgnoresEmptyHandlerChain(t *testing.T) {
	engine := New()
	engine.registerHandlerRoute("GET", "/empty", nil)
	require.Empty(t, engine.HandlerRoutes())
}

func BenchmarkRegisterHandlerRoute_Disabled(b *testing.B) {
	engine := New()
	engine.DisableRouteRegistry()
	handlers := HandlersChain{registeredRouteHandler}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		engine.registerHandlerRoute("GET", "/x", handlers)
	}
}
