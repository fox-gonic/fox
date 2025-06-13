package fox

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/fox-gonic/fox/utils"
)

func init() {
	gin.DebugPrintRouteFunc = func(_, _, _ string, _ int) {}
}

// IsDebugging returns true if the framework is running in debug mode.
// Use SetMode(gin.ReleaseMode) to disable debug mode.
func IsDebugging() bool {
	return foxMode == DebugMode
}

func debugPrintRoute(group *RouterGroup, httpMethod, absolutePath string, handlers HandlersChain) {
	if IsDebugging() {
		nuHandlers := len(group.router.Handlers) + len(handlers)
		handlerName := utils.NameOfFunction(handlers.Last())
		debugPrint("%-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}
}

func debugPrint(format string, values ...any) {
	if IsDebugging() {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(DefaultWriter, "[FOX-debug] "+format, values...)
	}
}
