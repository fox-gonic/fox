package engine

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fox-gonic/fox/logger"
)

var (
	// LoggerContextKey logger save in gin context
	LoggerContextKey = "_fox-goinc/fox/logger/context/key"
)

// LoggerConfig defines the config for Logger middleware.
type LoggerConfig struct {

	// SkipPaths is an url path array which logs are not written.
	// Optional.
	SkipPaths []string
}

// Logger middleware
func Logger(config ...LoggerConfig) gin.HandlerFunc {

	var conf LoggerConfig
	if len(config) > 0 {
		conf = config[0]
	}

	var skip map[string]struct{}

	if length := len(conf.SkipPaths); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range conf.SkipPaths {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		var (
			start      = time.Now()
			xRequestID = c.GetHeader(logger.TraceID)
			path       = c.Request.URL.Path
		)

		if len(xRequestID) == 0 {
			xRequestID = logger.DefaultGenRequestID()
			c.Request.Header.Set(logger.TraceID, xRequestID)
		}

		log := logger.New(xRequestID)
		c.Set(LoggerContextKey, log)

		c.Header(logger.TraceID, xRequestID)
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			if raw := c.Request.URL.RawQuery; raw != "" {
				path = path + "?" + raw
			}

			fields := map[string]interface{}{
				"method":    c.Request.Method,
				"path":      path,
				"client_ip": c.ClientIP(),
				"type":      "ENGINE",
				"status":    c.Writer.Status(),
				"latency":   time.Since(start).String(),
			}

			log.WithFields(fields).Info("[ROUTE]")
		}
	}
}
