package fox

import (
	"io"
	"os"
)

// DefaultWriter is the default io.Writer used by Fox for debug output.
var DefaultWriter io.Writer = os.Stdout

// DefaultErrorWriter is the default io.Writer used by Fox to debug errors
var DefaultErrorWriter io.Writer = os.Stderr
