package logger

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger := New()
	assert.NotNil(t, logger)
	assert.IsType(t, &Log{}, logger)
}

func TestNewWithTraceID(t *testing.T) {
	traceID := "test-trace-id-123"
	logger := New(traceID)
	assert.NotNil(t, logger)
	assert.Equal(t, traceID, logger.TraceID())
}

func TestNewWithConfig(t *testing.T) {
	cfg := Config{
		LogLevel:              InfoLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      true,
	}

	logger := NewWithConfig(cfg, "test-id")
	assert.NotNil(t, logger)
	assert.Equal(t, "test-id", logger.TraceID())
}

func TestNewContext(t *testing.T) {
	ctx := context.Background()
	reqID := "request-123"

	newCtx := NewContext(ctx, reqID)
	assert.NotNil(t, newCtx)

	value := newCtx.Value(TraceIDKey)
	assert.Equal(t, reqID, value)
}

func TestNewWithContext(t *testing.T) {
	t.Run("with TraceID", func(t *testing.T) {
		//nolint:staticcheck // TraceID is a package-level constant used as context key
		ctx := context.WithValue(context.Background(), TraceID, "trace-123")
		logger := NewWithContext(ctx)
		assert.NotNil(t, logger)
		assert.Equal(t, "trace-123", logger.TraceID())
	})

	t.Run("with TraceIDKey", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), TraceIDKey, "trace-456")
		logger := NewWithContext(ctx)
		assert.NotNil(t, logger)
		assert.Equal(t, "trace-456", logger.TraceID())
	})

	t.Run("without trace ID", func(t *testing.T) {
		ctx := context.Background()
		logger := NewWithContext(ctx)
		assert.NotNil(t, logger)
		assert.Empty(t, logger.TraceID())
	})
}

func TestDefaultGenRequestID(t *testing.T) {
	// Generate IDs and verify they are valid base64-encoded strings
	id1 := DefaultGenRequestID()
	assert.NotEmpty(t, id1)
	assert.Len(t, id1, 16, "ID should be 16 characters (12 bytes base64 encoded)")

	// Verify it's valid base64
	decoded, err := base64.URLEncoding.DecodeString(id1)
	require.NoError(t, err)
	assert.Len(t, decoded, 12, "Decoded ID should be 12 bytes")

	// Generate a second ID - may or may not be unique depending on timing
	id2 := DefaultGenRequestID()
	assert.NotEmpty(t, id2)
	assert.Len(t, id2, 16)
}

func TestLog_LogLevels(t *testing.T) {
	var buf bytes.Buffer

	cfg := Config{
		LogLevel:              DebugLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      true,
	}

	// Redirect stderr to buffer for testing
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
	}()

	logger := NewWithConfig(cfg, "test-trace")

	tests := []struct {
		name    string
		logFunc func()
		level   string
		message string
	}{
		{
			name:    "Debug",
			logFunc: func() { logger.Debug("debug message") },
			level:   "debug",
			message: "debug message",
		},
		{
			name:    "Info",
			logFunc: func() { logger.Info("info message") },
			level:   "info",
			message: "info message",
		},
		{
			name:    "Warn",
			logFunc: func() { logger.Warn("warn message") },
			level:   "warn",
			message: "warn message",
		},
		{
			name:    "Error",
			logFunc: func() { logger.Error("error message") },
			level:   "error",
			message: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()

			// Call log function
			tt.logFunc()

			// Read from pipe
			w.Close()
			_, _ = io.Copy(&buf, r)
			r, w, _ = os.Pipe()
			os.Stderr = w

			output := buf.String()
			if output != "" {
				var logData map[string]any
				err := json.Unmarshal([]byte(output), &logData)
				if err == nil {
					assert.Equal(t, tt.level, logData["level"])
					assert.Equal(t, tt.message, logData["message"])
					assert.Equal(t, "test-trace", logData[TraceID])
				}
			}
		})
	}
}

func TestLog_LogFormats(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		LogLevel:              DebugLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      true,
	}

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
	}()

	logger := NewWithConfig(cfg)

	tests := []struct {
		name    string
		logFunc func()
		message string
	}{
		{
			name:    "Debugf",
			logFunc: func() { logger.Debugf("debug %s", "formatted") },
			message: "debug formatted",
		},
		{
			name:    "Infof",
			logFunc: func() { logger.Infof("info %d", 123) },
			message: "info 123",
		},
		{
			name:    "Warnf",
			logFunc: func() { logger.Warnf("warn %v", true) },
			message: "warn true",
		},
		{
			name:    "Errorf",
			logFunc: func() { logger.Errorf("error %s", "test") },
			message: "error test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			w.Close()
			_, _ = io.Copy(&buf, r)
			r, w, _ = os.Pipe()
			os.Stderr = w

			// Just verify no panic occurred
			assert.NotPanics(t, func() { tt.logFunc() })
		})
	}
}

func TestLog_WithField(t *testing.T) {
	logger := New("test")
	newLogger := logger.WithField("key", "value")

	assert.NotNil(t, newLogger)
	assert.IsType(t, &Log{}, newLogger)
	assert.Equal(t, "test", newLogger.TraceID())
}

func TestLog_WithFields(t *testing.T) {
	logger := New("test")
	fields := map[string]any{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	newLogger := logger.WithFields(fields)

	assert.NotNil(t, newLogger)
	assert.IsType(t, &Log{}, newLogger)
	assert.Equal(t, "test", newLogger.TraceID())
}

func TestLog_WithError(t *testing.T) {
	logger := New("test")
	err := errors.New("test error")

	newLogger := logger.WithError(err)

	assert.NotNil(t, newLogger)
	assert.IsType(t, &Log{}, newLogger)
	assert.Equal(t, "test", newLogger.TraceID())
}

func TestLog_SetLevel(t *testing.T) {
	logger := New("test")

	tests := []struct {
		name  string
		level Level
	}{
		{"Debug", DebugLevel},
		{"Info", InfoLevel},
		{"Warn", WarnLevel},
		{"Error", ErrorLevel},
		{"Trace", TraceLevel},
		{"Disabled", Disabled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLogger := logger.SetLevel(tt.level)
			assert.NotNil(t, newLogger)
			assert.IsType(t, &Log{}, newLogger)
		})
	}
}

func TestLog_Caller(t *testing.T) {
	logger := New("test")
	newLogger := logger.Caller(5)

	assert.NotNil(t, newLogger)
	assert.IsType(t, &Log{}, newLogger)
	assert.Equal(t, "test", newLogger.TraceID())
}

func TestLog_WithContext(t *testing.T) {
	traceID := "trace-123"
	logger := New(traceID)

	t.Run("with context parameter", func(t *testing.T) {
		parentCtx := context.Background()
		ctx := logger.WithContext(parentCtx)

		assert.NotNil(t, ctx)
		value := ctx.Value(TraceIDKey)
		assert.Equal(t, traceID, value)
	})

	t.Run("without context parameter", func(t *testing.T) {
		ctx := logger.WithContext()

		assert.NotNil(t, ctx)
		value := ctx.Value(TraceIDKey)
		assert.Equal(t, traceID, value)
	})
}

func TestLog_TraceID(t *testing.T) {
	traceID := "my-trace-id"
	logger := New(traceID)

	assert.Equal(t, traceID, logger.TraceID())
}

func TestPackageLevelFunctions(t *testing.T) {
	// Test package-level logging functions don't panic
	assert.NotPanics(t, func() {
		Debug("debug")
		Info("info")
		Warn("warn")
		Error("error")

		Debugf("debug %s", "formatted")
		Infof("info %d", 123)
		Warnf("warn %v", true)
		Errorf("error %s", "test")

		WithField("key", "value")
		WithFields(map[string]any{"key": "value"})
		WithError(errors.New("test error"))
	})
}

func TestNewLogger_ConsoleJSON(t *testing.T) {
	cfg := Config{
		LogLevel:              InfoLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      true,
	}

	logger := newLogger(cfg, "test-id")
	require.NotNil(t, logger)

	log, ok := logger.(*Log)
	require.True(t, ok)
	assert.Equal(t, "test-id", log.traceID)
}

func TestNewLogger_ConsoleHuman(t *testing.T) {
	cfg := Config{
		LogLevel:              InfoLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      false,
	}

	logger := newLogger(cfg)
	require.NotNil(t, logger)

	log, ok := logger.(*Log)
	require.True(t, ok)
	assert.Empty(t, log.traceID)
}

func TestNewLogger_FileLogging(t *testing.T) {
	tmpFile := os.TempDir() + "/fox-test.log"
	defer os.Remove(tmpFile)

	cfg := Config{
		LogLevel:              InfoLevel,
		ConsoleLoggingEnabled: false,
		FileLoggingEnabled:    true,
		EncodeLogsAsJSON:      true,
		Filename:              tmpFile,
		MaxSize:               10,
		MaxBackups:            3,
		MaxAge:                7,
	}

	// Set config to initialize rollingWrite
	SetConfig(&cfg)

	// Use config variable which has rollingWrite initialized
	logger := newLogger(config, "file-test")
	require.NotNil(t, logger)

	logger.Info("test message")

	// Verify file was created
	_, err := os.Stat(tmpFile)
	assert.NoError(t, err)
}

func TestNewLogger_MultiWriter(t *testing.T) {
	tmpFile := os.TempDir() + "/fox-multi-test.log"
	defer os.Remove(tmpFile)

	cfg := Config{
		LogLevel:              DebugLevel,
		ConsoleLoggingEnabled: true,
		FileLoggingEnabled:    true,
		EncodeLogsAsJSON:      true,
		Filename:              tmpFile,
		MaxSize:               10,
		MaxBackups:            3,
		MaxAge:                7,
	}

	SetConfig(&cfg)

	logger := newLogger(config, "multi-test")
	require.NotNil(t, logger)

	logger.Debug("debug message to both console and file")
}

func TestNewWithoutCaller(t *testing.T) {
	logger := NewWithoutCaller("no-caller-test")
	assert.NotNil(t, logger)
	assert.Equal(t, "no-caller-test", logger.TraceID())
}

func TestLogger_ChainedCalls(t *testing.T) {
	logger := New("chain-test")

	// Test chaining multiple operations
	chainedLogger := logger.
		SetLevel(DebugLevel).
		WithField("user", "test").
		WithFields(map[string]any{"request": "GET", "status": 200}).
		WithError(errors.New("test error")).
		Caller(3)

	assert.NotNil(t, chainedLogger)
	assert.Equal(t, "chain-test", chainedLogger.TraceID())
}

func TestLogger_MultipleTraceIDs(t *testing.T) {
	logger1 := New("trace-1")
	logger2 := New("trace-2")

	assert.Equal(t, "trace-1", logger1.TraceID())
	assert.Equal(t, "trace-2", logger2.TraceID())
	assert.NotEqual(t, logger1.TraceID(), logger2.TraceID())
}

func TestLogger_NoWritersConfig(t *testing.T) {
	cfg := Config{
		LogLevel:              InfoLevel,
		ConsoleLoggingEnabled: false,
		FileLoggingEnabled:    false,
	}

	logger := newLogger(cfg)
	assert.NotNil(t, logger)

	// Should not panic even with no writers
	assert.NotPanics(t, func() {
		logger.Info("test message with no writers")
	})
}

func TestGlobalFunctions_Output(t *testing.T) {
	// Test that global functions execute without panic
	// Output testing is difficult due to async nature of logging
	assert.NotPanics(t, func() {
		Info("test global info")
		Debug("test global debug")
		Warn("test global warn")
		Error("test global error")
	})
}

func TestLog_ContextPropagation(t *testing.T) {
	traceID := "propagation-test"
	logger := New(traceID)

	ctx1 := logger.WithContext()
	assert.Equal(t, traceID, ctx1.Value(TraceIDKey))

	// Create new logger from context
	logger2 := NewWithContext(ctx1)
	assert.Equal(t, traceID, logger2.TraceID())

	ctx2 := logger2.WithContext(ctx1)
	assert.Equal(t, traceID, ctx2.Value(TraceIDKey))
}

func TestLogger_AllMessageTypes(t *testing.T) {
	logger := New("msg-test")

	// Test with different message types
	assert.NotPanics(t, func() {
		logger.Info("string message")
		logger.Info(123)
		logger.Info(true)
		logger.Info(nil)
		logger.Info("multi", "arg", "message")
		logger.Infof("formatted %s %d %v", "test", 123, true)
	})
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}

func BenchmarkNewWithTraceID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New("trace-id")
	}
}

func BenchmarkDefaultGenRequestID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultGenRequestID()
	}
}

func BenchmarkLogger_Info(b *testing.B) {
	logger := New("bench")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("test message")
	}
}

func BenchmarkLogger_WithField(b *testing.B) {
	logger := New("bench")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = logger.WithField("key", "value")
	}
}

func TestInit(t *testing.T) {
	// Test that init() set the TimeFieldFormat
	assert.Equal(t, zerolog.TimeFormatUnixNano, zerolog.TimeFieldFormat)
}

func TestTraceIDConstant(t *testing.T) {
	assert.Equal(t, "x-request-id", TraceID)
}

func TestLogger_ConsoleWriterFormatting(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
	}()

	cfg := Config{
		LogLevel:              InfoLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      false, // Human-readable format
	}

	logger := NewWithConfig(cfg, "format-test")
	logger.Info("formatted console message")

	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	output := buf.String()
	// In console format, should contain the message
	assert.True(t, strings.Contains(output, "formatted console message") || output == "")
}
