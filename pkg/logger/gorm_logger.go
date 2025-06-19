package logger

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm/logger"
)

// GormLogger 自定义的GORM日志器，输出JSON格式的SQL日志
type GormLogger struct {
	logger                    zerolog.Logger
	logLevel                  logger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

// NewGormLogger 创建新的GORM日志器
func NewGormLogger(zlogger zerolog.Logger, level logger.LogLevel) *GormLogger {
	return &GormLogger{
		logger:                    zlogger,
		logLevel:                  level,
		slowThreshold:             200 * time.Millisecond, // 默认慢查询阈值200ms
		ignoreRecordNotFoundError: true,
	}
}

// LogMode 设置日志级别
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info 信息级别日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		l.logger.Info().
			Str("component", "gorm").
			Str("level", "info").
			Msgf(msg, data...)
	}
}

// Warn 警告级别日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		l.logger.Warn().
			Str("component", "gorm").
			Str("level", "warn").
			Msgf(msg, data...)
	}
}

// Error 错误级别日志
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		l.logger.Error().
			Str("component", "gorm").
			Str("level", "error").
			Msgf(msg, data...)
	}
}

// Trace SQL执行日志
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	event := l.logger.Info()
	if err != nil && (!l.ignoreRecordNotFoundError || err.Error() != "record not found") {
		event = l.logger.Error()
	} else if elapsed > l.slowThreshold && l.slowThreshold != 0 {
		event = l.logger.Warn()
	}

	event.
		Str("component", "gorm").
		Str("type", "sql").
		Str("sql", sql).
		Int64("rows", rows).
		Str("duration", elapsed.String()).
		Float64("duration_ms", float64(elapsed.Nanoseconds())/1e6)

	if err != nil {
		event.Err(err)
	}

	if elapsed > l.slowThreshold && l.slowThreshold != 0 {
		event.Bool("slow_query", true)
	}

	event.Msg("SQL execution")
}

// SetSlowThreshold 设置慢查询阈值
func (l *GormLogger) SetSlowThreshold(threshold time.Duration) {
	l.slowThreshold = threshold
}

// SetIgnoreRecordNotFoundError 设置是否忽略记录未找到错误
func (l *GormLogger) SetIgnoreRecordNotFoundError(ignore bool) {
	l.ignoreRecordNotFoundError = ignore
}
