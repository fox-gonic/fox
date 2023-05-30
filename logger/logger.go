package logger

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type key int // key is unexported and used for Context

const (
	// TraceIDKey key
	TraceIDKey key = 0
)

// NewContext return context with reqid
func NewContext(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, reqID)
}

var pid = uint32(time.Now().UnixNano() % 4294967291)

// TraceID is the key for the x-request-id header.
var TraceID = "x-request-id"

// DefaultGenRequestID default generate request id
var DefaultGenRequestID func() string = func() string {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:], pid)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	return base64.URLEncoding.EncodeToString(b[:])
}

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

	// context
	WithContext(ctx ...context.Context) context.Context
}

// New return logger
var New func(traceID ...string) Logger = newLogger

// NewWithoutCaller new log without caller field
func NewWithoutCaller(reqID ...string) Logger {
	return newLogger(reqID...)
}

// NewWithContext return logger with context
func NewWithContext(ctx context.Context) Logger {
	reqID := ""

	reqid, ok := ctx.Value(TraceIDKey).(string)
	if ok {
		reqID = reqid
	}

	log := newLogger(reqID)
	l := log.(*Log)
	zl := l.log.With().CallerWithSkipFrameCount(3).Logger()
	l.log = &zl

	return l
}

// newLogger return Logger
func newLogger(traceID ...string) Logger {

	var trace string
	if len(traceID) > 0 {
		trace = traceID[0]
	}

	var writers []io.Writer

	if config.ConsoleLoggingEnabled {
		if config.EncodeLogsAsJSON {
			writers = append(writers, os.Stderr)
		} else {
			writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: DefaultLogTimeFormat})
		}
	}

	if config.FileLoggingEnabled {
		if config.EncodeLogsAsJSON {
			writers = append(writers, config.rollingWrite)
		} else {
			writers = append(writers, zerolog.ConsoleWriter{Out: config.rollingWrite, TimeFormat: DefaultLogTimeFormat})
		}
	}

	mw := io.MultiWriter(writers...)

	c := zerolog.New(mw).With().Timestamp().CallerWithSkipFrameCount(3)

	if trace != "" {
		c = c.Str(TraceID, trace)
	}

	l := c.Logger().Level(zerolog.Level(DefaultLogLevel))

	log := &Log{log: &l, traceID: trace}

	return log
}

var std = New("Std").Caller(4)

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

// WithContext return context with log
func (l *Log) WithContext(ctx ...context.Context) context.Context {
	if len(ctx) > 0 {
		return context.WithValue(ctx[0], TraceIDKey, l.traceID)
	}
	return context.WithValue(context.Background(), TraceIDKey, l.traceID)
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

// Debug debug level
func Debug(arguments ...interface{}) {
	std.Debug(arguments)
}

// Info info level
func Info(arguments ...interface{}) {
	std.Info(arguments)
}

// Warn warn level
func Warn(arguments ...interface{}) {
	std.Warn(arguments)
}

// Error error level
func Error(arguments ...interface{}) {
	std.Error(arguments)
}

// Fatal fatal level
func Fatal(arguments ...interface{}) {
	std.Fatal(arguments)
}

// Panic panic level
func Panic(arguments ...interface{}) {
	std.Panic(arguments)
}

// Debugf debug format
func Debugf(format string, arguments ...interface{}) {
	std.Debugf(format, arguments...)
}

// Infof info format
func Infof(format string, arguments ...interface{}) {
	std.Infof(format, arguments...)
}

// Warnf warn format
func Warnf(format string, arguments ...interface{}) {
	std.Warnf(format, arguments...)
}

// Errorf error format
func Errorf(format string, arguments ...interface{}) {
	std.Errorf(format, arguments...)
}

// Fatalf fatal format
func Fatalf(format string, arguments ...interface{}) {
	std.Fatalf(format, arguments...)
}

// Panicf panic format
func Panicf(format string, arguments ...interface{}) {
	std.Panicf(format, arguments...)
}

// WithField add new field
func WithField(key string, value interface{}) Logger {
	return std.WithField(key, value)
}

// WithFields add new fields
func WithFields(fields map[string]interface{}) Logger {
	return std.WithFields(fields)
}

// WithError adds the field "error" with serialized err to the logger context.
func WithError(err error) Logger {
	return std.WithError(err)
}
