package fox

// XRequestIDKey is the key for the X-Request-ID header.
var XRequestIDKey = "X-Request-ID"

// Logger logger methods
type Logger interface {
	Debug(arguments ...interface{})
	Info(arguments ...interface{})
	Warn(arguments ...interface{})
	Error(arguments ...interface{})
	Fatal(arguments ...interface{})
	Panic(arguments ...interface{})
	Debugf(format string, arguments ...interface{})
	Infof(format string, arguments ...interface{})
	Warnf(format string, arguments ...interface{})
	Errorf(format string, arguments ...interface{})
	Fatalf(format string, arguments ...interface{})
	Panicf(format string, arguments ...interface{})

	TraceID() string
}
