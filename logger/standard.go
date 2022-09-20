package logger

var standard Logger

func init() {
	log := newLogger("Std")
	l := log.(*Log)
	zl := l.log.With().CallerWithSkipFrameCount(4).Logger()
	l.log = &zl
	standard = l
}

// Debug debug level
func Debug(arguments ...interface{}) {
	standard.Debug(arguments)
}

// Info info level
func Info(arguments ...interface{}) {
	standard.Info(arguments)
}

// Warn warn level
func Warn(arguments ...interface{}) {
	standard.Warn(arguments)
}

// Error error level
func Error(arguments ...interface{}) {
	standard.Error(arguments)
}

// Fatal fatal level
func Fatal(arguments ...interface{}) {
	standard.Fatal(arguments)
}

// Panic panic level
func Panic(arguments ...interface{}) {
	standard.Panic(arguments)
}

// Debugf debug format
func Debugf(format string, arguments ...interface{}) {
	standard.Debug(format, arguments)
}

// Infof info format
func Infof(format string, arguments ...interface{}) {
	standard.Info(format, arguments)
}

// Warnf warn format
func Warnf(format string, arguments ...interface{}) {
	standard.Warn(format, arguments)
}

// Errorf error format
func Errorf(format string, arguments ...interface{}) {
	standard.Error(format, arguments)
}

// Fatalf fatal format
func Fatalf(format string, arguments ...interface{}) {
	standard.Fatal(format, arguments)
}

// Panicf panic format
func Panicf(format string, arguments ...interface{}) {
	standard.Panic(format, arguments)
}

// WithField add new field
func WithField(key string, value interface{}) Logger {
	return standard.WithField(key, value)
}

// WithFields add new fields
func WithFields(fields map[string]interface{}) Logger {
	return standard.WithFields(fields)
}

// WithError adds the field "error" with serialized err to the logger context.
func WithError(err error) Logger {
	return standard.WithError(err)
}

// SetLevel set level
func SetLevel(level Level) Logger {
	return standard.SetLevel(level)
}

// Caller set caller frame
func Caller(frame int) Logger {
	return standard.Caller(frame)
}

// TraceID trace id
func TraceID() string {
	return standard.TraceID()
}
