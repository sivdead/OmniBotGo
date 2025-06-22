package sqlite

import (
	"gorm.io/gorm/logger"
)

// Option -.
type Option func(*SQLite)

// MaxConnections -.
func MaxConnections(maxConnections int) Option {
	return func(s *SQLite) {
		s.maxConnections = maxConnections
	}
}

// LogLevel -.
func LogLevel(logLevel logger.LogLevel) Option {
	return func(s *SQLite) {
		s.logLevel = logLevel
	}
}

// SlowThreshold -.
func SlowThreshold(slowThreshold int) Option {
	return func(s *SQLite) {
		s.slowThreshold = slowThreshold
	}
}
