package logger

import "github.com/rs/zerolog"

// Level type
type Level string

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = "debug"
	// InfoLevel defines info log level.
	InfoLevel Level = "info"
	// WarnLevel defines warn log level.
	WarnLevel Level = "warn"
	// ErrorLevel defines error log level.
	ErrorLevel Level = "error"
	// FatalLevel defines fatal log level.
	FatalLevel Level = "fatal"
	// PanicLevel defines panic log level.
	PanicLevel Level = "panic"
	// NoLevel defines an absent log level.
	NoLevel Level = "no"
	// Disabled disables the logger.
	Disabled Level = "disabled"
	// TraceLevel defines trace log level.
	TraceLevel Level = "trace"
)

func (l Level) ZerologLevel() zerolog.Level {
	switch l {
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	case FatalLevel:
		return zerolog.FatalLevel
	case PanicLevel:
		return zerolog.PanicLevel
	case NoLevel:
		return zerolog.NoLevel
	case Disabled:
		return zerolog.Disabled
	case TraceLevel:
		return zerolog.TraceLevel
	default:
		// default to trace level
		return zerolog.TraceLevel
	}
}

var DefaultLogLevel = TraceLevel
