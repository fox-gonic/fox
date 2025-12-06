package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetConfig(t *testing.T) {
	// Save original config
	originalConfig := config

	defer func() {
		// Restore original config
		config = originalConfig
		DefaultLogLevel = TraceLevel
	}()

	t.Run("basic config", func(t *testing.T) {
		cfg := &Config{
			LogLevel:              InfoLevel,
			ConsoleLoggingEnabled: true,
			EncodeLogsAsJSON:      true,
			FileLoggingEnabled:    false,
		}

		SetConfig(cfg)

		assert.Equal(t, InfoLevel, config.LogLevel)
		assert.True(t, config.ConsoleLoggingEnabled)
		assert.True(t, config.EncodeLogsAsJSON)
		assert.False(t, config.FileLoggingEnabled)
		assert.Equal(t, InfoLevel, DefaultLogLevel)
	})

	t.Run("with file logging", func(t *testing.T) {
		tmpFile := filepath.Join(os.TempDir(), "fox-config-test.log")
		defer os.Remove(tmpFile)

		cfg := &Config{
			LogLevel:              DebugLevel,
			ConsoleLoggingEnabled: false,
			FileLoggingEnabled:    true,
			Filename:              tmpFile,
			MaxSize:               50,
			MaxBackups:            5,
			MaxAge:                30,
		}

		SetConfig(cfg)

		assert.Equal(t, DebugLevel, config.LogLevel)
		assert.False(t, config.ConsoleLoggingEnabled)
		assert.True(t, config.FileLoggingEnabled)
		assert.Equal(t, tmpFile, config.Filename)
		assert.Equal(t, 50, config.MaxSize)
		assert.Equal(t, 5, config.MaxBackups)
		assert.Equal(t, 30, config.MaxAge)
		assert.NotNil(t, config.rollingWrite)
	})

	t.Run("file logging without filename", func(t *testing.T) {
		cfg := &Config{
			LogLevel:              WarnLevel,
			ConsoleLoggingEnabled: true,
			FileLoggingEnabled:    true,
			Filename:              "", // Empty filename
		}

		SetConfig(cfg)

		assert.True(t, config.FileLoggingEnabled)
		assert.NotEmpty(t, config.Filename)

		// Should generate default filename in temp directory
		expectedName := filepath.Base(os.Args[0]) + "-fox.log"
		expectedPath := filepath.Join(os.TempDir(), expectedName)
		assert.Equal(t, expectedPath, config.Filename)
		assert.NotNil(t, config.rollingWrite)
	})

	t.Run("multiple SetConfig calls", func(t *testing.T) {
		cfg1 := &Config{
			LogLevel:              ErrorLevel,
			ConsoleLoggingEnabled: true,
			EncodeLogsAsJSON:      false,
		}
		SetConfig(cfg1)
		assert.Equal(t, ErrorLevel, config.LogLevel)
		assert.False(t, config.EncodeLogsAsJSON)

		cfg2 := &Config{
			LogLevel:              InfoLevel,
			ConsoleLoggingEnabled: true,
			EncodeLogsAsJSON:      true,
		}
		SetConfig(cfg2)
		assert.Equal(t, InfoLevel, config.LogLevel)
		assert.True(t, config.EncodeLogsAsJSON)
	})
}

func TestConfig_DefaultValues(t *testing.T) {
	// Test default config values
	defaultCfg := Config{
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      false,
		FileLoggingEnabled:    false,
	}

	assert.True(t, defaultCfg.ConsoleLoggingEnabled)
	assert.False(t, defaultCfg.EncodeLogsAsJSON)
	assert.False(t, defaultCfg.FileLoggingEnabled)
}

func TestConfig_AllFields(t *testing.T) {
	cfg := Config{
		LogLevel:              DebugLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      true,
		FileLoggingEnabled:    true,
		Filename:              "/tmp/test.log",
		MaxSize:               100,
		MaxBackups:            10,
		MaxAge:                7,
	}

	assert.Equal(t, DebugLevel, cfg.LogLevel)
	assert.True(t, cfg.ConsoleLoggingEnabled)
	assert.True(t, cfg.EncodeLogsAsJSON)
	assert.True(t, cfg.FileLoggingEnabled)
	assert.Equal(t, "/tmp/test.log", cfg.Filename)
	assert.Equal(t, 100, cfg.MaxSize)
	assert.Equal(t, 10, cfg.MaxBackups)
	assert.Equal(t, 7, cfg.MaxAge)
}

func TestConfig_RollingWriter(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "fox-rolling-test.log")
	defer os.Remove(tmpFile)

	cfg := &Config{
		LogLevel:           InfoLevel,
		FileLoggingEnabled: true,
		Filename:           tmpFile,
		MaxSize:            10,
		MaxBackups:         3,
		MaxAge:             7,
	}

	SetConfig(cfg)

	require.NotNil(t, config.rollingWrite)

	// Test writing to rolling writer
	n, err := config.rollingWrite.Write([]byte("test log message\n"))
	require.NoError(t, err)
	assert.Positive(t, n)

	// Verify file was created
	_, err = os.Stat(tmpFile)
	assert.NoError(t, err)
}

func TestConfig_FileLoggingWithCustomPath(t *testing.T) {
	tmpDir := os.TempDir()
	customFile := filepath.Join(tmpDir, "custom", "path", "fox-test.log")

	// Create parent directories
	err := os.MkdirAll(filepath.Dir(customFile), 0o755)
	require.NoError(t, err)
	defer os.RemoveAll(filepath.Join(tmpDir, "custom"))

	cfg := &Config{
		LogLevel:           InfoLevel,
		FileLoggingEnabled: true,
		Filename:           customFile,
		MaxSize:            10,
	}

	SetConfig(cfg)

	assert.Equal(t, customFile, config.Filename)
	assert.NotNil(t, config.rollingWrite)
}

func TestDefaultLogTimeFormat(t *testing.T) {
	expected := "2006-01-02 15:04:05.000000"
	assert.Equal(t, expected, DefaultLogTimeFormat)
}

func TestConfig_LogLevelUpdate(t *testing.T) {
	originalLevel := DefaultLogLevel
	defer func() {
		DefaultLogLevel = originalLevel
	}()

	levels := []Level{
		DebugLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel,
		FatalLevel,
		TraceLevel,
	}

	for _, level := range levels {
		cfg := &Config{
			LogLevel:              level,
			ConsoleLoggingEnabled: true,
		}

		SetConfig(cfg)

		assert.Equal(t, level, DefaultLogLevel)
		assert.Equal(t, level, config.LogLevel)
	}
}

func TestConfig_BothConsoleAndFileLogging(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "fox-both-test.log")
	defer os.Remove(tmpFile)

	cfg := &Config{
		LogLevel:              InfoLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      true,
		FileLoggingEnabled:    true,
		Filename:              tmpFile,
		MaxSize:               10,
		MaxBackups:            3,
		MaxAge:                7,
	}

	SetConfig(cfg)

	assert.True(t, config.ConsoleLoggingEnabled)
	assert.True(t, config.FileLoggingEnabled)
	assert.NotNil(t, config.rollingWrite)

	// Create logger and test it writes to both
	logger := New("both-test")
	logger.Info("test message to both console and file")

	// Verify file was created
	_, err := os.Stat(tmpFile)
	assert.NoError(t, err)
}

func BenchmarkSetConfig(b *testing.B) {
	cfg := &Config{
		LogLevel:              InfoLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      true,
	}

	for i := 0; i < b.N; i++ {
		SetConfig(cfg)
	}
}

func BenchmarkSetConfig_WithFileLogging(b *testing.B) {
	tmpFile := filepath.Join(os.TempDir(), "fox-bench.log")
	defer os.Remove(tmpFile)

	cfg := &Config{
		LogLevel:           InfoLevel,
		FileLoggingEnabled: true,
		Filename:           tmpFile,
		MaxSize:            10,
	}

	for i := 0; i < b.N; i++ {
		SetConfig(cfg)
	}
}
