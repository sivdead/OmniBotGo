package postgres

import (
	"time"

	"gorm.io/gorm/logger"
)

// Option -.
type Option func(*Postgres)

// MaxPoolSize -.
func MaxPoolSize(size int) Option {
	return func(p *Postgres) {
		p.maxPoolSize = size
	}
}

// ConnAttempts -.
func ConnAttempts(attempts int) Option {
	return func(p *Postgres) {
		p.connAttempts = attempts
	}
}

// ConnTimeout -.
func ConnTimeout(timeout time.Duration) Option {
	return func(p *Postgres) {
		p.connTimeout = timeout
	}
}

// LogLevel 设置GORM日志级别
func LogLevel(level logger.LogLevel) Option {
	return func(p *Postgres) {
		p.logLevel = level
	}
}
