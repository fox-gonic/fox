# Logger Configuration Example

This example demonstrates various logger configurations in Fox.

## Features

- Console logging
- File logging with rotation
- Combined console and file logging
- JSON formatted logs
- Different log levels
- Structured logging
- Skip paths configuration
- Production-ready configuration

## Running

The example file contains multiple configuration examples. Uncomment the one you want to try:

```bash
go run main.go
```

## Log Levels

Fox supports the following log levels (from most to least verbose):

- **Debug**: Debugging information
- **Info**: General informational messages
- **Warn**: Warning messages
- **Error**: Error messages
- **Fatal**: Fatal errors (application will exit)
- **Panic**: Panic-level errors

## Configuration Options

### Basic Configuration

```go
logger.SetConfig(&logger.Config{
    LogLevel:              logger.InfoLevel,
    ConsoleLoggingEnabled: true,
    EncodeLogsAsJSON:      false,
})
```

### File Logging Configuration

```go
logger.SetConfig(&logger.Config{
    LogLevel:              logger.InfoLevel,
    ConsoleLoggingEnabled: false,
    FileLoggingEnabled:    true,
    Filename:              "./logs/app.log",
    MaxSize:               10,   // MB before rotation
    MaxBackups:            3,    // Number of old log files
    MaxAge:                7,    // Days to keep old logs
    EncodeLogsAsJSON:      false,
})
```

## Example Configurations

### 1. Development Mode

**Goal**: Easy-to-read console logs for debugging

```go
logger.SetConfig(&logger.Config{
    LogLevel:              logger.DebugLevel,
    ConsoleLoggingEnabled: true,
    EncodeLogsAsJSON:      false, // Human-readable
})
```

Output:
```
2025-12-06 20:30:45.123 INF  Processing request path=/api/users method=GET
2025-12-06 20:30:45.124 DBG  Query parameters parsed count=3
```

### 2. Production Mode (Console Only)

**Goal**: JSON logs for log aggregation systems

```go
logger.SetConfig(&logger.Config{
    LogLevel:              logger.InfoLevel, // No debug logs
    ConsoleLoggingEnabled: true,
    EncodeLogsAsJSON:      true, // Machine-readable
})
```

Output:
```json
{"level":"info","time":"2025-12-06T20:30:45.123Z","message":"Processing request","path":"/api/users"}
```

### 3. Production Mode (File + Console)

**Goal**: Logs to both console and rotating log files

```go
logger.SetConfig(&logger.Config{
    LogLevel:              logger.InfoLevel,
    ConsoleLoggingEnabled: true,
    FileLoggingEnabled:    true,
    Filename:              "/var/log/myapp/app.log",
    MaxSize:               100,  // 100MB
    MaxBackups:            30,   // Keep 30 old files
    MaxAge:                90,   // 90 days
    EncodeLogsAsJSON:      true,
})
```

### 4. High-Performance Mode

**Goal**: Minimize logging overhead

```go
logger.SetConfig(&logger.Config{
    LogLevel:              logger.WarnLevel, // Only warnings and errors
    ConsoleLoggingEnabled: false,
    FileLoggingEnabled:    true,
    Filename:              "./logs/errors.log",
    EncodeLogsAsJSON:      true,
})
```

## Middleware Configuration

### Skip Health Check Endpoints

```go
router.Use(fox.Logger(fox.LoggerConfig{
    SkipPaths: []string{
        "/health",
        "/readiness",
        "/metrics",
    },
}))
```

## Usage in Handlers

### Get Logger from Context

```go
router.GET("/example", func(ctx *fox.Context) string {
    log := logger.GetLogger(ctx.Context)
    log.Info("Processing request")
    return "OK"
})
```

### Structured Logging

```go
log.WithFields(map[string]interface{}{
    "user_id":   123,
    "action":    "login",
    "ip":        "192.168.1.1",
}).Info("User logged in")
```

### Log with Error

```go
if err := doSomething(); err != nil {
    log.WithError(err).Error("Operation failed")
    return "", err
}
```

### Log Levels in Code

```go
log.Debug("Variable value:", value)
log.Info("Operation completed")
log.Warn("Deprecated API used")
log.Error("Operation failed")
```

## Log Rotation

Fox uses `lumberjack` for log rotation:

**MaxSize**: When log file reaches this size (in megabytes), it rotates
```go
MaxSize: 100 // Rotate at 100MB
```

**MaxBackups**: Number of old log files to keep
```go
MaxBackups: 10 // Keep 10 old log files
```

**MaxAge**: Days to keep old log files
```go
MaxAge: 30 // Delete logs older than 30 days
```

**Example rotation**:
```
app.log           (current, 95MB)
app.log.1         (yesterday, 100MB)
app.log.2         (2 days ago, 100MB)
app.log.3         (3 days ago, 100MB)
```

When `app.log` reaches 100MB, it becomes `app.log.1`, and a new `app.log` is created.

## TraceID

Fox automatically adds a TraceID to each request:

```go
router.Use(fox.Logger())

router.GET("/trace", func(ctx *fox.Context) string {
    traceID := ctx.TraceID()
    log := logger.GetLogger(ctx.Context)
    log.Info("Request received", "trace_id", traceID)
    return "OK"
})
```

Log output:
```json
{"level":"info","trace_id":"abc123xyz","message":"Request received"}
```

## Integration with Log Aggregation

### ELK Stack (Elasticsearch, Logstash, Kibana)

```go
logger.SetConfig(&logger.Config{
    EncodeLogsAsJSON:      true,
    FileLoggingEnabled:    true,
    Filename:              "/var/log/myapp/app.json.log",
})
```

Configure Logstash to read from `/var/log/myapp/app.json.log`.

### Splunk

```go
logger.SetConfig(&logger.Config{
    EncodeLogsAsJSON:      true,
    ConsoleLoggingEnabled: true, // Splunk reads from stdout
})
```

### CloudWatch (AWS)

Use AWS CloudWatch agent to collect logs from files or stdout.

## Best Practices

1. **Use Appropriate Log Levels**
   - Debug: Development only
   - Info: Important business events
   - Warn: Recoverable issues
   - Error: Errors that need attention

2. **Structured Logging**
   ```go
   // Good
   log.WithFields(map[string]interface{}{
       "user_id": 123,
       "action":  "login",
   }).Info("User action")

   // Avoid
   log.Info("User 123 performed login")
   ```

3. **Include Context**
   - Request ID / Trace ID
   - User ID
   - Session ID
   - Relevant business data

4. **Don't Log Sensitive Data**
   - Passwords
   - API keys
   - Credit card numbers
   - Personal information (PII)

5. **Log Retention**
   - Keep enough logs for debugging
   - Balance storage costs
   - Comply with regulations

6. **Performance Considerations**
   - Higher log levels in production
   - Skip health check endpoints
   - Use JSON format for parsing efficiency

## Example Production Setup

```go
logger.SetConfig(&logger.Config{
    LogLevel:              logger.InfoLevel,
    ConsoleLoggingEnabled: true,  // For container logs
    FileLoggingEnabled:    true,  // For persistent storage
    Filename:              "/var/log/myapp/app.log",
    MaxSize:               100,
    MaxBackups:            30,
    MaxAge:                90,
    EncodeLogsAsJSON:      true,  // For log aggregation
})

router.Use(fox.Logger(fox.LoggerConfig{
    SkipPaths: []string{"/health", "/metrics"},
}))
```

This configuration:
- Logs to both console (for Docker/K8s) and file
- Uses JSON format for machine parsing
- Rotates files at 100MB
- Keeps 30 backup files
- Retains logs for 90 days
- Skips noisy health check endpoints
