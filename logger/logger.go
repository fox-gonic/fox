package logger

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var pid = uint32(time.Now().UnixNano() % 4294967291)

// RequestIDKey is the key for the x-request-id header.
var RequestIDKey = "x-request-id"

// DefaultGenRequestID default generate request id
var DefaultGenRequestID func() string = func() string {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:], pid)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	return base64.URLEncoding.EncodeToString(b[:])
}

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

// Logger logger methods
type Logger interface {
	// STD log
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

	// Field logger
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger

	// Set level
	SetLevel(level Level) Logger

	// Caller skip frame count
	Caller(frame int) Logger

	// Trace ID
	TraceID() string
}

// NewLogger return logger
var NewLogger func(traceID string) Logger = newLogger

func newLogger(traceID string) Logger {

	w := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: DefaultLogTimeFormat}
	l := zerolog.New(w).
		With().
		Str(RequestIDKey, traceID).
		Timestamp().
		CallerWithSkipFrameCount(3).
		Logger().
		Level(zerolog.Level(DefaultLogLevel))

	log := &Log{log: &l, traceID: traceID}

	return log
}

// Log implement Logger
type Log struct {
	log     *zerolog.Logger
	traceID string
}

// Debug debug level
func (l *Log) Debug(arguments ...interface{}) {
	l.log.Debug().Msg(fmt.Sprint(arguments...))
}

// Info info level
func (l *Log) Info(arguments ...interface{}) {
	l.log.Info().Msg(fmt.Sprint(arguments...))
}

// Warn warn level
func (l *Log) Warn(arguments ...interface{}) {
	l.log.Warn().Msg(fmt.Sprint(arguments...))
}

// Error error level
func (l *Log) Error(arguments ...interface{}) {
	l.log.Error().Msg(fmt.Sprint(arguments...))
}

// Fatal fatal level
func (l *Log) Fatal(arguments ...interface{}) {
	l.log.Fatal().Msg(fmt.Sprint(arguments...))
}

// Panic panic level
func (l *Log) Panic(arguments ...interface{}) {
	l.log.Panic().Msg(fmt.Sprint(arguments...))
}

// Debugf debug format
func (l *Log) Debugf(format string, arguments ...interface{}) {
	l.log.Debug().Msg(fmt.Sprintf(format, arguments...))
}

// Infof info format
func (l *Log) Infof(format string, arguments ...interface{}) {
	l.log.Info().Msg(fmt.Sprintf(format, arguments...))
}

// Warnf warn format
func (l *Log) Warnf(format string, arguments ...interface{}) {
	l.log.Warn().Msg(fmt.Sprintf(format, arguments...))
}

// Errorf error format
func (l *Log) Errorf(format string, arguments ...interface{}) {
	l.log.Error().Msg(fmt.Sprintf(format, arguments...))
}

// Fatalf fatal format
func (l *Log) Fatalf(format string, arguments ...interface{}) {
	l.log.Fatal().Msg(fmt.Sprintf(format, arguments...))
}

// Panicf panic format
func (l *Log) Panicf(format string, arguments ...interface{}) {
	l.log.Panic().Msg(fmt.Sprintf(format, arguments...))
}

// WithField add new field
func (l *Log) WithField(key string, value interface{}) Logger {
	log := l.log.With().Fields(map[string]interface{}{key: value}).Logger()
	return &Log{
		log:     &log,
		traceID: l.traceID,
	}
}

// WithFields add new fields
func (l *Log) WithFields(fields map[string]interface{}) Logger {
	log := l.log.With().Fields(fields).Logger()
	return &Log{
		log:     &log,
		traceID: l.traceID,
	}
}

// WithError adds the field "error" with serialized err to the logger context.
func (l *Log) WithError(err error) Logger {
	log := l.log.With().Err(err).Logger()
	return &Log{
		log:     &log,
		traceID: l.traceID,
	}
}

// SetLevel set level
func (l *Log) SetLevel(level Level) Logger {
	zl := l.log.Level(zerolog.Level(DefaultLogLevel))
	return &Log{
		log:     &zl,
		traceID: l.traceID,
	}
}

// Caller set caller frame
func (l *Log) Caller(frame int) Logger {
	zl := l.log.With().CallerWithSkipFrameCount(frame).Logger()
	return &Log{
		log:     &zl,
		traceID: l.traceID,
	}
}

// TraceID trace id
func (l *Log) TraceID() string {
	return l.traceID
}
