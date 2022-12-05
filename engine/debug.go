package engine

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/fox-gonic/fox/utils"
)

const ginSupportMinGoVer = 14

func init() {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {}
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
