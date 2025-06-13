package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
)

// DefaultLogTimeFormat default log time format
var DefaultLogTimeFormat = "2006-01-02 15:04:05.000000"

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixNano
}

// Config for logging
type Config struct {
	// log level
	LogLevel Level `json:"log_level" yaml:"log_level" mapstructure:"log_level"`

	// Enable console logging
	ConsoleLoggingEnabled bool `json:"console_logging_enabled" yaml:"console_logging_enabled" mapstructure:"console_logging_enabled"`

	// EncodeLogsAsJSON makes the log framework log JSON
	EncodeLogsAsJSON bool `json:"encode_logs_as_json" yaml:"encode_logs_as_json" mapstructure:"encode_logs_as_json"`

	// FileLoggingEnabled makes the framework log to a file, the fields below can be skipped if this value is false!
	FileLoggingEnabled bool `json:"file_logging_enabled" yaml:"file_logging_enabled" mapstructure:"file_logging_enabled"`

	// Filename is the name of the logfile which will be placed inside the directory
	Filename string `json:"filename" yaml:"filename" mapstructure:"filename"`

	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int `json:"max_size" yaml:"max_size" mapstructure:"max_size"`

	// MaxBackups the max number of rolled files to keep
	MaxBackups int `json:"max_backups" yaml:"max_backups" mapstructure:"max_backups"`

	// MaxAge the max age in days to keep a logfile
	MaxAge int `json:"max_age" yaml:"max_age" mapstructure:"max_age"`

	rollingWrite io.Writer
}

var config = Config{
	ConsoleLoggingEnabled: true,
	EncodeLogsAsJSON:      false,
	FileLoggingEnabled:    false,
}

// SetConfig set logger config
func SetConfig(cfg *Config) {
	config = *cfg

	DefaultLogLevel = cfg.LogLevel

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
