package fox

import (
	"errors"
	"net"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// RecoveryFunc defines the function passable to CustomRecovery.
type RecoveryFunc func(c *Context, err any)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() HandlerFunc {
	return RecoveryWithHandle(defaultHandleRecovery)
}

// RecoveryWithHandle returns a middleware for a given writer that recovers from any panics and calls the provided handle func to handle it.
func RecoveryWithHandle(handle RecoveryFunc) HandlerFunc {
	return func(c *Context) (err any) {
		defer func() {
			if err = recover(); err != nil {

				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				stack := debug.Stack()
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				headers := strings.Split(string(httpRequest), "\r\n")
				for idx, header := range headers {
					current := strings.Split(header, ":")
					if current[0] == "Authorization" {
						headers[idx] = current[0] + ": *"
					}
				}

				headersToStr := strings.Join(headers, "\r\n")
				if brokenPipe {
					c.Logger.Errorf("%s\n%s", err, headersToStr)
				} else if IsDebugging() {
					c.Logger.Errorf("[Recovery] panic recovered:\n%s\n%s\n\x1b[31m%s\033[0m", headersToStr, err, stack)
				} else {
					c.Logger.Errorf("[Recovery] panic recovered:\n%s\n\x1b[31m%s\033[0m", err, stack)
				}

				if brokenPipe {
					c.Abort()
				} else {
					handle(c, err)
				}
			}
		}()
		c.Next()
		return
	}
}

func defaultHandleRecovery(c *Context, err any) {
	// c.AbortWithStatus(http.StatusInternalServerError)
}
