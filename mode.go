package fox

import (
	"io"
	"os"
)

const (
	// DebugMode indicates fox mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates fox mode is release.
	ReleaseMode = "release"
	// TestMode indicates fox mode is test.
	TestMode = "test"
)

// DefaultWriter is the default io.Writer used by Fox for debug output.
var DefaultWriter io.Writer = os.Stdout

// DefaultErrorWriter is the default io.Writer used by Fox to debug errors
var DefaultErrorWriter io.Writer = os.Stderr

var engineMode string

// SetMode sets engine mode according to input string.
func SetMode(value string) {
	switch value {
	case DebugMode, ReleaseMode, TestMode:
		engineMode = value
	default:
		engineMode = DebugMode
	}
}
