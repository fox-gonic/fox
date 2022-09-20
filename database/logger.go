package database

import (
	"context"
	"time"

	"gorm.io/gorm/logger"

	"github.com/fox-gonic/fox"
)

// Log implement gorm logger
type Log struct {
	fox.Logger
	slowThreshold time.Duration
}

var defaultSlowThreshold = 50 * time.Millisecond

func toLoggerLevel(lvl logger.LogLevel) fox.Level {
	switch lvl {
	case logger.Error:
		return fox.ErrorLevel
	case logger.Info:
		return fox.InfoLevel
	case logger.Silent:
		return fox.NoLevel
	case logger.Warn:
		return fox.WarnLevel
	default:
		return fox.TraceLevel
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
		l.Logger.WithFields(field).Errorf("%v", err)
	case elapsed > l.slowThreshold:
		field["slow_query"] = true
		l.Logger.WithFields(field).Warnf("Elapsed %s exceeded, Max %s", elapsed.String(), l.slowThreshold.String())
	default:
		l.Logger.WithFields(field).Info()
	}
}

// newLog
func newLog(slow int64, xReqID string) *Log {

	log := fox.NewLogger(xReqID).Caller(6).WithFields(map[string]interface{}{"type": "DATABASE"})

	threshold := defaultSlowThreshold
	if slow > 0 {
		threshold = time.Duration(slow) * time.Millisecond
	}
	return &Log{log, threshold}
}
