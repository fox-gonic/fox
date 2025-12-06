package utils

import (
	"path"
	"reflect"
	"runtime"
)

// NameOfFunction return function name
func NameOfFunction(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// JoinPaths join paths
func JoinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	// Check for trailing slash preservation
	if len(relativePath) > 0 && relativePath[len(relativePath)-1] == '/' &&
		len(finalPath) > 0 && finalPath[len(finalPath)-1] != '/' {
		return finalPath + "/"
	}
	return finalPath
}
