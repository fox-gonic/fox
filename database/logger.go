package database

import (
	"context"
	"time"

	gormlogger "gorm.io/gorm/logger"

	log "github.com/fox-gonic/fox/logger"
)

// logger implement gorm logger.Interface
type logger struct {
	log.Logger
	SlowThreshold time.Duration
}

var defaultSlowThreshold = 50 * time.Millisecond

func toLoggerLevel(level gormlogger.LogLevel) log.Level {
	switch level {
	case gormlogger.Error:
		return log.ErrorLevel
	case gormlogger.Info:
		return log.InfoLevel
	case gormlogger.Silent:
		return log.NoLevel
	case gormlogger.Warn:
		return log.WarnLevel
	default:
		return log.TraceLevel
	}
}

// FromContext from context logger
func (l *logger) FromContext(ctx context.Context) log.Logger {
	if requestID, ok := ctx.Value(log.TraceIDKey).(string); ok {
		return l.WithField(log.TraceIDKey, requestID)
	}
	return l.Logger
}

// LogMode implement gorm logger
func (l *logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	l.Logger = l.Logger.SetLevel(toLoggerLevel(level))
	return l
}

// Info implement gorm logger
func (l *logger) Info(ctx context.Context, s string, vals ...interface{}) {
	l.FromContext(ctx).Infof(s, vals...)
}

// Warn implement gorm logger
func (l *logger) Warn(ctx context.Context, s string, vals ...interface{}) {
	l.FromContext(ctx).Warnf(s, vals...)
}

// Error implement gorm logger
func (l *logger) Error(ctx context.Context, s string, vals ...interface{}) {
	l.FromContext(ctx).Errorf(s, vals...)
}

// Trace implement gorm logger
func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	var (
		elapsed   = time.Since(begin)
		sql, rows = fc()
		fields    = map[string]interface{}{
			"latency":       elapsed.String(),
			"sql":           sql,
			"rows_affected": rows,
		}
		logger = l.FromContext(ctx)
	)

	switch {
	case err != nil:
		logger.WithFields(fields).Errorf("%v", err)
	case elapsed > l.SlowThreshold:
		fields["slow_query"] = true
		logger.WithFields(fields).Warnf("Elapsed %s exceeded, Max %s", elapsed.String(), l.SlowThreshold.String())
	default:
		logger.WithFields(fields).Info()
	}
}

// NewLogger return custom logger
func NewLogger(slowThreshold int, requestID ...string) gormlogger.Interface {

	fields := map[string]interface{}{"type": "DATABASE"}

	if len(requestID) > 0 {
		fields[log.TraceIDKey] = requestID[0]
	}

	l := log.New().Caller(6).WithFields(fields)

	threshold := defaultSlowThreshold
	if slowThreshold > 0 {
		threshold = time.Duration(slowThreshold) * time.Millisecond
	}

	return &logger{Logger: l, SlowThreshold: threshold}
}
