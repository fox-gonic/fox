package fox

import (
	"fmt"
	"strings"
)

// IsDebugging returns true if the framework is running in debug mode.
// Use SetMode(fox.ReleaseMode) to disable debug mode.
func IsDebugging() bool {
	return engineMode == DevelopmentMode
}

// DebugPrintRouteFunc indicates debug log output format.
var DebugPrintRouteFunc func(httpMethod, absolutePath, handlerName string, nuHandlers int)

func debugPrintRoute(httpMethod, absolutePath string, handlers HandlersChain) {
	if IsDebugging() {
		nuHandlers := len(handlers)
		handlerName := nameOfFunction(handlers.Last())

		if DebugPrintRouteFunc == nil {
			debugPrint("%-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
		} else {
			DebugPrintRouteFunc(httpMethod, absolutePath, handlerName, nuHandlers)
		}
	}
}

func debugPrint(format string, values ...any) {
	if IsDebugging() {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(DefaultWriter, "[DBG] "+format, values...)
	}
}

func debugPrintWARNINGNew() {
	debugPrint(`[WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export FOX_MODE=release
 - using code:	fox.SetMode(fox.ReleaseMode)

`)
}
