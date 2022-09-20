package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
)

// Config for logging
type Config struct {
	// log level
	LogLevel int `mapstructure:"log_level"`

	// Enable console logging
	ConsoleLoggingEnabled bool `mapstructure:"console_logging_enabled"`

	// EncodeLogsAsJSON makes the log framework log JSON
	EncodeLogsAsJSON bool `mapstructure:"encode_logs_as_json"`

	// FileLoggingEnabled makes the framework log to a file, the fields below can be skipped if this value is false!
	FileLoggingEnabled bool `mapstructure:"file_logging_enabled"`

	// Filename is the name of the logfile which will be placed inside the directory
	Filename string `mapstructure:"filename"`

	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int `mapstructure:"max_size"`

	// MaxBackups the max number of rolled files to keep
	MaxBackups int `mapstructure:"max_backups"`

	// MaxAge the max age in days to keep a logfile
	MaxAge int `mapstructure:"max_age"`

	rollingWrite io.Writer
}

var config = &Config{
	ConsoleLoggingEnabled: true,
	EncodeLogsAsJSON:      false,
	FileLoggingEnabled:    false,
}

// SetConfig set logger config
func SetConfig(cfg Config) {

	config = &cfg

	DefaultLogLevel = Level(cfg.LogLevel)

	if config.FileLoggingEnabled {
		if config.Filename == "" {
			name := filepath.Base(os.Args[0]) + "-fox.log"
			config.Filename = filepath.Join(os.TempDir(), name)
		}

		config.rollingWrite = &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
		}
	}
}
