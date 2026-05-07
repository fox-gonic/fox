package fox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func registeredRouteHandler(_ *Context) string {
	return "ok"
}

type manifestUserRequest struct {
	ID     string `uri:"id" binding:"required"`
	Search string `query:"search"`
}

type manifestUserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type manifestNestedType struct {
	When time.Time `json:"when"`
}

type manifestCustomError struct {
	Code string `json:"code"`
}

func (err manifestCustomError) Error() string {
	return err.Code
}

type manifestComplexType struct {
	Items  []manifestNestedType           `json:"items"`
	Lookup map[string]*manifestNestedType `json:"lookup"`
	Self   *manifestComplexType           `json:"self"`
	Err    manifestCustomError            `json:"err"`
	hidden string
}

type manifestRepeatedCompositeType struct {
	First  []manifestNestedType `json:"first"`
	Second []manifestNestedType `json:"second"`
}

func manifestUserHandler(_ *Context, _ manifestUserRequest) (manifestUserResponse, error) {
	return manifestUserResponse{}, nil
}

func manifestClosureHandler() HandlerFunc {
	return func(_ *Context, _ manifestUserRequest) (manifestUserResponse, error) {
		return manifestUserResponse{}, nil
	}
}

func manifestSharedTypeClosureHandler() HandlerFunc {
	return func(_ *Context, _ manifestNestedType) (manifestNestedType, error) {
		return manifestNestedType{}, nil
	}
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

func TestWriteRouteManifest(t *testing.T) {
	engine := New()
	engine.GET("/users/:id", manifestUserHandler)

	path := filepath.Join(t.TempDir(), "api", "routes.manifest.json")
	require.NoError(t, WriteRouteManifest(engine, path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var manifest RouteManifest
	require.NoError(t, json.Unmarshal(data, &manifest))
	require.Equal(t, RouteManifestVersion, manifest.Version)
	require.Len(t, manifest.Routes, 1)
	require.Equal(t, "GET", manifest.Routes[0].Method)
	require.Equal(t, "/users/:id", manifest.Routes[0].Path)
	require.Contains(t, manifest.Routes[0].Handler, "manifestUserHandler")
	require.Empty(t, manifest.Routes[0].InputTypes)
	require.Empty(t, manifest.Routes[0].ResultTypes)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	routes := raw["routes"].([]any)
	route := routes[0].(map[string]any)
	require.Contains(t, route, "handler")
	require.NotContains(t, route, "handlerName")
	require.NotContains(t, route, "handlerType")
	require.NotContains(t, route, "inputs")
	require.NotContains(t, route, "results")
}

func TestWriteRouteManifestRejectsEmptyPath(t *testing.T) {
	require.EqualError(t, WriteRouteManifest(New(), ""), "route manifest path is required")
}

func TestWriteRouteManifestReportsWriteErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "routes.manifest.json")
	require.NoError(t, os.Mkdir(path, 0o755))

	err := WriteRouteManifest(New(), path)
	require.Error(t, err)
	require.Contains(t, err.Error(), "write route manifest")
}

func TestWriteRouteManifestReportsCreateDirErrors(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "api")
	require.NoError(t, os.WriteFile(filePath, []byte("not a directory"), 0o644))

	err := WriteRouteManifest(New(), filepath.Join(filePath, "routes.manifest.json"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "create route manifest dir")
}

func TestRouteManifestFromNilEngine(t *testing.T) {
	manifest := RouteManifestFromEngine(nil)
	require.Equal(t, RouteManifestVersion, manifest.Version)
	require.Empty(t, manifest.Routes)
}

func TestWriteRouteManifestKeepsClosureTypes(t *testing.T) {
	engine := New()
	engine.GET("/users/:id", manifestClosureHandler())

	manifest := RouteManifestFromEngine(engine)
	require.Len(t, manifest.Routes, 1)
	require.Contains(t, manifest.Routes[0].Handler, ".func")
	require.Len(t, manifest.Routes[0].InputTypes, 1)
	require.Equal(t, "manifestUserRequest", manifest.Routes[0].InputTypes[0].Name)
	require.Len(t, manifest.Routes[0].ResultTypes, 2)
	require.Equal(t, "manifestUserResponse", manifest.Routes[0].ResultTypes[0].Name)
}

func TestWriteRouteManifestExpandsClosureInputsAndResultsIndependently(t *testing.T) {
	engine := New()
	engine.GET("/nested", manifestSharedTypeClosureHandler())

	manifest := RouteManifestFromEngine(engine)
	require.Len(t, manifest.Routes, 1)
	require.Len(t, manifest.Routes[0].InputTypes, 1)
	require.Len(t, manifest.Routes[0].ResultTypes, 2)
	require.Len(t, manifest.Routes[0].InputTypes[0].Fields, 1)
	require.Equal(t, "manifestNestedType", manifest.Routes[0].ResultTypes[0].Name)
	require.Len(t, manifest.Routes[0].ResultTypes[0].Fields, 1)
}

func TestRouteManifestTypeSerializesComplexTypes(t *testing.T) {
	require.Empty(t, manifestComplexType{}.hidden)

	typ := routeManifestType(reflect.TypeOf(manifestComplexType{}), map[reflect.Type]bool{})

	require.Equal(t, "struct", typ.Kind)
	require.Equal(t, "manifestComplexType", typ.Name)
	require.Len(t, typ.Fields, 4)

	items := typ.Fields[0]
	require.Equal(t, "Items", items.Name)
	require.Equal(t, "slice", items.Type.Kind)
	require.NotNil(t, items.Type.Elem)
	require.Equal(t, "manifestNestedType", items.Type.Elem.Name)
	require.Len(t, items.Type.Elem.Fields, 1)
	require.Equal(t, "Time", items.Type.Elem.Fields[0].Type.Name)
	require.Empty(t, items.Type.Elem.Fields[0].Type.Fields)

	lookup := typ.Fields[1]
	require.Equal(t, "Lookup", lookup.Name)
	require.Equal(t, "map", lookup.Type.Kind)
	require.NotNil(t, lookup.Type.Key)
	require.NotNil(t, lookup.Type.Elem)
	require.Equal(t, "string", lookup.Type.Key.Kind)
	require.Equal(t, "ptr", lookup.Type.Elem.Kind)
	require.Equal(t, "manifestNestedType", lookup.Type.Elem.Elem.Name)

	self := typ.Fields[2]
	require.Equal(t, "Self", self.Name)
	require.Equal(t, "ptr", self.Type.Kind)
	require.Equal(t, "manifestComplexType", self.Type.Elem.Name)
	require.Empty(t, self.Type.Elem.Fields)

	errField := typ.Fields[3]
	require.Equal(t, "Err", errField.Name)
	require.Equal(t, "string", errField.Type.Kind)
	require.Equal(t, "string", errField.Type.Name)
}

func TestRouteManifestTypeKeepsRepeatedUnnamedCompositeStructure(t *testing.T) {
	typ := routeManifestType(reflect.TypeOf(manifestRepeatedCompositeType{}), map[reflect.Type]bool{})

	require.Len(t, typ.Fields, 2)
	for _, field := range typ.Fields {
		require.Equal(t, "slice", field.Type.Kind)
		require.NotNil(t, field.Type.Elem)
		require.Equal(t, "manifestNestedType", field.Type.Elem.Name)
	}
}

func TestRouteManifestTypeHandlesNil(t *testing.T) {
	require.Empty(t, routeManifestType(nil, map[reflect.Type]bool{}))
}

func TestRouteManifestTypePreservesErrorInterfaceResult(t *testing.T) {
	typ := routeManifestType(reflect.TypeOf((*error)(nil)).Elem(), map[reflect.Type]bool{})
	require.Equal(t, "interface", typ.Kind)
	require.Equal(t, "error", typ.Name)
}

func TestRouteManifestIsFoxContextUsesContextType(t *testing.T) {
	require.True(t, routeManifestIsFoxContext(reflect.TypeOf(&Context{})))
	require.True(t, routeManifestIsFoxContext(reflect.TypeOf(Context{})))
	require.False(t, routeManifestIsFoxContext(reflect.TypeOf(struct {
		Context Context
	}{})))
}

func TestRouteManifestTypeSerializesConcreteErrorsAsStrings(t *testing.T) {
	typ := routeManifestType(reflect.TypeOf(manifestCustomError{}), map[reflect.Type]bool{})
	require.Equal(t, "string", typ.Kind)
	require.Equal(t, "string", typ.Name)
	require.Empty(t, typ.Fields)
}

func TestRouteManifestRouteWithoutHandlerType(t *testing.T) {
	route := routeManifestRoute(RouteInfo{Method: "GET", Path: "/raw", HandlerName: "raw.func1"})
	require.Equal(t, "GET", route.Method)
	require.Equal(t, "/raw", route.Path)
	require.Equal(t, "raw.func1", route.Handler)
	require.Empty(t, route.InputTypes)
	require.Empty(t, route.ResultTypes)
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
