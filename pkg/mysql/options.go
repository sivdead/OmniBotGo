package mysql

import (
	"time"

	"gorm.io/gorm/logger"
)

// Option -.
type Option func(*MySQL)

// MaxConnections -.
func MaxConnections(size int) Option {
	return func(m *MySQL) {
		m.maxConnections = size
	}
}

// MaxIdleConns -.
func MaxIdleConns(size int) Option {
	return func(m *MySQL) {
		m.maxIdleConns = size
	}
}

// ConnMaxLifetime -.
func ConnMaxLifetime(lifetime time.Duration) Option {
	return func(m *MySQL) {
		m.connMaxLifetime = lifetime
	}
}

// ConnMaxIdleTime -.
func ConnMaxIdleTime(idleTime time.Duration) Option {
	return func(m *MySQL) {
		m.connMaxIdleTime = idleTime
	}
}

// ConnAttempts -.
func ConnAttempts(attempts int) Option {
	return func(m *MySQL) {
		m.connAttempts = attempts
	}
}

// ConnTimeout -.
func ConnTimeout(timeout time.Duration) Option {
	return func(m *MySQL) {
		m.connTimeout = timeout
	}
}

// LogLevel 设置GORM日志级别
func LogLevel(level logger.LogLevel) Option {
	return func(m *MySQL) {
		m.logLevel = level
	}
}
