package main

import (
	"os"

	"github.com/fox-gonic/fox"
	"github.com/fox-gonic/fox/logger"
)

func main() {
	// Example 1: Console logging only (default)
	example1ConsoleOnly()

	// Example 2: File logging with rotation
	example2FileLogging()

	// Example 3: Both console and file logging
	example3BothOutputs()

	// Example 4: JSON formatted logs
	example4JSONLogs()

	// Example 5: Different log levels
	example5LogLevels()
}

// Example 1: Console logging only (development mode)
func example1ConsoleOnly() {
	logger.SetConfig(&logger.Config{
		LogLevel:              logger.InfoLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      false, // Human-readable format
	})

	router := fox.New()
	router.Use(fox.Logger())

	router.GET("/console", func(ctx *fox.Context) string {
		log := logger.NewWithContext(ctx.Context)
		log.Info("Processing console logging request")
		log.Debug("This is a debug message (won't show at Info level)")
		return "Check console for logs"
	})

	// Uncomment to run
	// router.Run(":8080")
}

// Example 2: File logging with rotation
func example2FileLogging() {
	logger.SetConfig(&logger.Config{
		LogLevel:              logger.InfoLevel,
		ConsoleLoggingEnabled: false,
		FileLoggingEnabled:    true,
		Filename:              "./logs/app.log",
		MaxSize:               10, // megabytes
		MaxBackups:            3,  // number of backups
		MaxAge:                7,  // days
		EncodeLogsAsJSON:      false,
	})

	router := fox.New()
	router.Use(fox.Logger())

	router.GET("/file", func(ctx *fox.Context) string {
		log := logger.NewWithContext(ctx.Context)
		log.Info("This will be written to file")
		return "Check ./logs/app.log for logs"
	})

	// Uncomment to run
	// router.Run(":8081")
}

// Example 3: Both console and file logging (production mode)
func example3BothOutputs() {
	logger.SetConfig(&logger.Config{
		LogLevel:              logger.InfoLevel,
		ConsoleLoggingEnabled: true,
		FileLoggingEnabled:    true,
		Filename:              "./logs/production.log",
		MaxSize:               50,   // 50MB before rotation
		MaxBackups:            10,   // Keep 10 backup files
		MaxAge:                30,   // Keep logs for 30 days
		EncodeLogsAsJSON:      true, // JSON format for parsing
	})

	router := fox.New()
	router.Use(fox.Logger())

	router.GET("/both", func(ctx *fox.Context) string {
		log := logger.NewWithContext(ctx.Context)
		log.Info("Logged to both console and file")
		return "Check both console and file"
	})

	// Uncomment to run
	// router.Run(":8082")
}

// Example 4: JSON formatted logs (for log aggregation)
func example4JSONLogs() {
	logger.SetConfig(&logger.Config{
		LogLevel:              logger.DebugLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      true, // JSON format
	})

	router := fox.New()
	router.Use(fox.Logger())

	router.GET("/json", func(ctx *fox.Context) map[string]any {
		log := logger.NewWithContext(ctx.Context)

		log.Info("User action logged")
		log.WithFields(map[string]any{
			"user_id": 123,
			"action":  "view_profile",
		}).Info("Structured logging example")

		return map[string]any{
			"message": "Check console for JSON logs",
		}
	})

	// Uncomment to run
	// router.Run(":8083")
}

// Example 5: Different log levels
func example5LogLevels() {
	// Set to Debug to see all messages
	logger.SetConfig(&logger.Config{
		LogLevel:              logger.DebugLevel,
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJSON:      false,
	})

	router := fox.New()

	// Skip logging for health check endpoint
	router.Use(fox.Logger(fox.LoggerConfig{
		SkipPaths: []string{"/health"},
	}))

	router.GET("/levels", func(ctx *fox.Context) string {
		log := logger.NewWithContext(ctx.Context)

		log.Debug("Debug level - debugging info")
		log.Info("Info level - general info")
		log.Warn("Warn level - warning message")
		log.Error("Error level - error occurred")

		return "Check console for different log levels"
	})

	router.GET("/health", func() string {
		return "OK" // This endpoint won't be logged
	})

	// Uncomment to run
	// router.Run(":8084")
}

// Complete production example
//
//nolint:unused
func productionExample() {
	// Production configuration
	logger.SetConfig(&logger.Config{
		LogLevel:              logger.InfoLevel, // Don't log Debug in production
		ConsoleLoggingEnabled: true,
		FileLoggingEnabled:    true,
		Filename:              "/var/log/myapp/app.log",
		MaxSize:               100,  // 100MB
		MaxBackups:            30,   // Keep 30 backups
		MaxAge:                90,   // 90 days retention
		EncodeLogsAsJSON:      true, // JSON for log aggregation (ELK, Splunk, etc.)
	})

	router := fox.New()

	// Global middleware
	router.Use(fox.Logger(fox.LoggerConfig{
		SkipPaths: []string{
			"/health",
			"/readiness",
			"/metrics",
		},
	}))
	router.Use(fox.NewXResponseTimer())

	// Business routes
	router.POST("/api/users", func(ctx *fox.Context) (map[string]any, error) {
		log := logger.NewWithContext(ctx.Context)

		log.Info("Creating new user")

		// Structured logging with fields
		log.WithFields(map[string]any{
			"user_id":   123,
			"user_type": "premium",
		}).Info("User created successfully")

		return map[string]any{
			"id":      123,
			"message": "User created",
		}, nil
	})

	// Error logging example
	router.GET("/api/error", func(ctx *fox.Context) (string, error) {
		log := logger.NewWithContext(ctx.Context)

		err := os.ErrNotExist
		log.WithError(err).Error("File not found")

		return "", err
	})

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
