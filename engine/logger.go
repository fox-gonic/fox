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

// Logger middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			start      = time.Now()
			xRequestID = c.GetHeader(logger.TraceID)
		)

		if len(xRequestID) == 0 {
			xRequestID = logger.DefaultGenRequestID()
		}

		c.Header(logger.TraceID, xRequestID)

		log := logger.New(xRequestID)
		c.Set(LoggerContextKey, log)

		fields := map[string]interface{}{
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"client_ip": c.ClientIP(),
			"type":      "REQ",
		}

		if !strings.HasPrefix(c.Request.URL.Path, "/static") {
			fields["action"] = "Start"
			log.WithFields(fields).Info("[Started]")
		}

		c.Next()

		if !strings.HasPrefix(c.Request.URL.Path, "/static") {
			fields["status"] = c.Writer.Status()
			fields["latency"] = time.Since(start).String()
			fields["action"] = "Finished"
			log.WithFields(fields).Info("[Completed]")
		}
	}
}
