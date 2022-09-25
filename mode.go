package fox

import (
	"io"
	"os"
)

const (
	// DevelopmentMode indicates fox mode is development.
	DevelopmentMode = "development"
	// ProductionMode indicates fox mode is production.
	ProductionMode = "production"
	// TestMode indicates fox mode is test.
	TestMode = "test"
)

// DefaultWriter is the default io.Writer used by Fox for development output.
var DefaultWriter io.Writer = os.Stdout

// DefaultErrorWriter is the default io.Writer used by Fox to development errors
var DefaultErrorWriter io.Writer = os.Stderr

var engineMode = DevelopmentMode

// SetMode sets engine mode according to input string.
func SetMode(value string) {
	switch value {
	case DevelopmentMode, ProductionMode, TestMode:
		engineMode = value
	default:
		engineMode = DevelopmentMode
	}
}
