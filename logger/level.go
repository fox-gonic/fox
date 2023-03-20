package logger

// Level type
type Level int8

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	NoLevel
	// Disabled disables the logger.
	Disabled
	// TraceLevel defines trace log level.
	TraceLevel Level = -1
)

var (
	// DefaultLogLevel log level
	DefaultLogLevel Level = TraceLevel

	// DefaultLogTimeFormat default log time format
	DefaultLogTimeFormat = "2006-01-02 15:04:05.000000"
)
