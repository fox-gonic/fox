package database

import (
	"context"
	"time"

	"gorm.io/gorm/logger"

	log "great-dev.com/common/logger"
)

// Log implement gorm logger
type Log struct {
	log.Logger
	slowThreshold time.Duration
}

var defaultSlowThreshold = 50 * time.Millisecond

func toLoggerLevel(lvl logger.LogLevel) log.Level {
	switch lvl {
	case logger.Error:
		return log.ErrorLevel
	case logger.Info:
		return log.InfoLevel
	case logger.Silent:
		return ""
	case logger.Warn:
		return log.WarnLevel
	default:
		return log.TraceLevel
	}
}

// LogMode implement gorm logger
func (l *Log) LogMode(lvl logger.LogLevel) logger.Interface {
	l.Logger = l.Logger.SetLevel(toLoggerLevel(lvl))
	return l
}

// Info implement gorm logger
func (l *Log) Info(_ context.Context, s string, vals ...interface{}) {
	l.Infof(s, vals...)
}

// Warn implement gorm logger
func (l *Log) Warn(_ context.Context, s string, vals ...interface{}) {
	l.Warnf(s, vals...)
}

// Error implement gorm logger
func (l *Log) Error(_ context.Context, s string, vals ...interface{}) {
	l.Errorf(s, vals...)
}

// Trace implement gorm logger
func (l *Log) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	field := map[string]interface{}{
		"latency":       elapsed.String(),
		"sql":           sql,
		"rows_affected": rows,
	}

	switch {
	case err != nil:
		l.Logger.WithField(field).Errorf("%v", err)
	case elapsed > l.slowThreshold:
		field["slow_query"] = true
		l.Logger.WithField(field).Warnf("Elapsed %s exceeded, Max %s", elapsed.String(), l.slowThreshold.String())
	default:
		l.Logger.WithField(field).Info()
	}
}

// newLog
func newLog(slow int64, xReqID string) *Log {
	log := log.NewWithoutCaller(xReqID).Caller(6).WithField(map[string]interface{}{"type": "DATABASE"})

	threshold := defaultSlowThreshold
	if slow > 0 {
		threshold = time.Duration(slow) * time.Millisecond
	}
	return &Log{log, threshold}
}
