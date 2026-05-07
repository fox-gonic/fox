package fox

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

var (
	routeManifestContextType = reflect.TypeOf(Context{})
	routeManifestErrorType   = reflect.TypeOf((*error)(nil)).Elem()
)

// RouteManifestVersion is the current JSON route manifest format version.
const RouteManifestVersion = "fox.route-manifest/v1"

// RouteManifest is a stable JSON representation of routes registered on an
// Engine. Tooling can consume it without booting the application itself.
type RouteManifest struct {
	Version string               `json:"version"`
	Routes  []RouteManifestRoute `json:"routes"`
}

// RouteManifestRoute describes one Fox route and the original business handler
// captured at registration time.
type RouteManifestRoute struct {
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	Handler     string              `json:"handler,omitempty"`
	InputTypes  []RouteManifestType `json:"inputTypes,omitempty"`
	ResultTypes []RouteManifestType `json:"resultTypes,omitempty"`
}

// RouteManifestType is a serializable subset of reflect.Type.
type RouteManifestType struct {
	Kind    string               `json:"kind"`
	Name    string               `json:"name,omitempty"`
	PkgPath string               `json:"pkgPath,omitempty"`
	Key     *RouteManifestType   `json:"key,omitempty"`
	Elem    *RouteManifestType   `json:"elem,omitempty"`
	Fields  []RouteManifestField `json:"fields,omitempty"`
}

// RouteManifestField is a serializable subset of reflect.StructField.
type RouteManifestField struct {
	Name      string            `json:"name"`
	Tag       string            `json:"tag,omitempty"`
	Anonymous bool              `json:"anonymous,omitempty"`
	Type      RouteManifestType `json:"type"`
}

// RouteManifestFromEngine returns a serializable snapshot of the Engine route
// registry.
func RouteManifestFromEngine(engine *Engine) RouteManifest {
	manifest := RouteManifest{Version: RouteManifestVersion}
	if engine == nil {
		return manifest
	}
	for _, route := range engine.HandlerRoutes() {
		manifest.Routes = append(manifest.Routes, routeManifestRoute(route))
	}
	return manifest
}

// WriteRouteManifest writes the Engine route registry as indented JSON.
func WriteRouteManifest(engine *Engine, path string) error {
	if path == "" {
		return errors.New("route manifest path is required")
	}
	data, err := json.MarshalIndent(RouteManifestFromEngine(engine), "", "  ")
	if err != nil {
		return fmt.Errorf("marshal route manifest: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create route manifest dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write route manifest: %w", err)
	}
	return nil
}

func routeManifestRoute(route RouteInfo) RouteManifestRoute {
	result := RouteManifestRoute{
		Method:  route.Method,
		Path:    route.Path,
		Handler: route.HandlerName,
	}
	if route.HandlerType == nil {
		return result
	}
	if !routeManifestNeedsInlineTypes(route.HandlerName) {
		return result
	}
	result.InputTypes = manifestTypeList(route.HandlerType, true, map[reflect.Type]bool{})
	result.ResultTypes = manifestTypeList(route.HandlerType, false, map[reflect.Type]bool{})
	return result
}

func routeManifestNeedsInlineTypes(handlerName string) bool {
	return strings.Contains(handlerName, ".func")
}

func manifestTypeList(typ reflect.Type, inputs bool, seen map[reflect.Type]bool) []RouteManifestType {
	var count int
	if inputs {
		count = typ.NumIn()
	} else {
		count = typ.NumOut()
	}
	result := make([]RouteManifestType, 0, count)
	for i := 0; i < count; i++ {
		if inputs {
			input := typ.In(i)
			if routeManifestIsFoxContext(input) {
				continue
			}
			result = append(result, routeManifestType(input, seen))
		} else {
			result = append(result, routeManifestType(typ.Out(i), seen))
		}
	}
	return result
}

func routeManifestType(typ reflect.Type, seen map[reflect.Type]bool) RouteManifestType {
	if typ == nil {
		return RouteManifestType{}
	}
	result := RouteManifestType{
		Kind:    typ.Kind().String(),
		Name:    typ.Name(),
		PkgPath: typ.PkgPath(),
	}
	if routeManifestIsConcreteError(typ) {
		return RouteManifestType{Kind: "string", Name: "string"}
	}
	// Anonymous composite types cannot form recursive cycles without a named
	// type, so keep their structure instead of emitting empty []T/map shells.
	if seen[typ] && typ.Name() != "" {
		return result
	}
	if routeManifestOpaqueType(typ) {
		return result
	}
	seen[typ] = true

	switch typ.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		elem := routeManifestType(typ.Elem(), seen)
		result.Elem = &elem
	case reflect.Map:
		key := routeManifestType(typ.Key(), seen)
		elem := routeManifestType(typ.Elem(), seen)
		result.Key = &key
		result.Elem = &elem
	case reflect.Struct:
		result.Fields = make([]RouteManifestField, 0, typ.NumField())
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.PkgPath != "" {
				continue
			}
			result.Fields = append(result.Fields, RouteManifestField{
				Name:      field.Name,
				Tag:       string(field.Tag),
				Anonymous: field.Anonymous,
				Type:      routeManifestType(field.Type, seen),
			})
		}
	}
	return result
}

func routeManifestIsConcreteError(typ reflect.Type) bool {
	if typ == nil || typ == routeManifestErrorType {
		return false
	}
	if typ.Implements(routeManifestErrorType) {
		return true
	}
	if typ.Kind() != reflect.Pointer && reflect.PointerTo(typ).Implements(routeManifestErrorType) {
		return true
	}
	return false
}

func routeManifestIsFoxContext(typ reflect.Type) bool {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ == routeManifestContextType
}

func routeManifestOpaqueType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		return false
	}
	return typ.PkgPath() == "time"
}
