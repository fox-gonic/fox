package engine

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fox-gonic/fox/logger"
)

var (
	// LoggerContextKey logger save in gin context
	LoggerContextKey = "_fox-goinc/fox/logger/context/key"
)

// GinLoggerMiddleware gin web framework logger middleware
func GinLoggerMiddleware(ServiceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		xRequestID := c.GetHeader(logger.TraceID)

		if len(xRequestID) == 0 {
			xRequestID = logger.DefaultGenRequestID()
		}

		c.Header(logger.TraceID, xRequestID)

		log := logger.New(xRequestID)
		c.Set(LoggerContextKey, log)

		if !strings.HasPrefix(c.Request.URL.Path, "/static") {
			logReq(log, c)
		}

		start := time.Now()
		c.Next()

		if !strings.HasPrefix(c.Request.URL.Path, "/static") {
			logResponse(log, c, start)
		}
	}
}

func logReq(log logger.Logger, c *gin.Context) {
	fields := map[string]interface{}{
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
		"client_ip": c.ClientIP(),
		"type":      "REQ",
		"action":    "Start",
	}

	log.WithFields(fields).Info("[Started]")
}

func logResponse(log logger.Logger, c *gin.Context, startTime time.Time) {
	fields := map[string]interface{}{
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
		"status":    c.Writer.Status(),
		"latency":   time.Since(startTime).String(), // 耗时
		"client_ip": c.ClientIP(),
		"type":      "REQ",
		"action":    "Finished",
	}

	log.WithFields(fields).Info("[Completed]")
}
