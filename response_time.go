package fox

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const headerXResponseTime = "X-Response-Time"

// XResponseTimer wrap gin response writer add start time
type XResponseTimer struct {
	gin.ResponseWriter
	start time.Time
	key   string
}

// WriteHeader implement http.ResponseWriter
func (w *XResponseTimer) WriteHeader(statusCode int) {
	buf := make([]byte, 0, 32)
	buf = strconv.AppendInt(buf, w.start.UnixMilli(), 10)
	buf = append(buf, ',', ' ')
	buf = strconv.AppendInt(buf, time.Since(w.start).Nanoseconds(), 10)
	w.Header().Set(w.key, string(buf))
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write implement http.ResponseWriter
func (w *XResponseTimer) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

// NewXResponseTimer x-response-time middleware
func NewXResponseTimer(key ...string) gin.HandlerFunc {
	k := headerXResponseTime
	if len(key) > 0 {
		k = key[0]
	}
	return func(c *gin.Context) {
		c.Writer = &XResponseTimer{
			ResponseWriter: c.Writer,
			start:          time.Now(),
			key:            k,
		}
		c.Next()
	}
}
