// Package database implements database connection factory for different database types.
package database

import (
	"fmt"
	"strings"

	"github.com/sivdead/OmniBotGo/pkg/mysql"
	"github.com/sivdead/OmniBotGo/pkg/postgres"
	"github.com/sivdead/OmniBotGo/pkg/sqlite"
	"gorm.io/gorm/logger"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgres"
	PostgresQL DatabaseType = "postgresql" // alias for postgres
	SQLite     DatabaseType = "sqlite"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type           string
	DSN            string
	MaxConnections int
	LogLevel       string // 新增日志级别配置
	SlowThreshold  int    // 慢查询阈值(毫秒)
}

// parseLogLevel 将字符串转换为GORM日志级别
func parseLogLevel(level string) logger.LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return logger.Info // GORM中Info级别会显示SQL
	case "info":
		return logger.Warn
	case "warn", "warning":
		return logger.Error
	case "error":
		return logger.Error
	case "silent":
		return logger.Silent
	default:
		return logger.Silent
	}
}

// NewDatabase creates a database connection based on the specified type
func NewDatabase(config DatabaseConfig) (CommonDB, error) {
	dbType := DatabaseType(config.Type)
	logLevel := parseLogLevel(config.LogLevel)

	switch dbType {
	case MySQL:
		return mysql.New(
			config.DSN,
			mysql.MaxConnections(config.MaxConnections),
			mysql.LogLevel(logLevel),
			mysql.SlowThreshold(config.SlowThreshold),
		)

	case PostgreSQL, PostgresQL:
		return postgres.New(
			config.DSN,
			postgres.MaxPoolSize(config.MaxConnections),
			postgres.LogLevel(logLevel),
			postgres.SlowThreshold(config.SlowThreshold),
		)

	case SQLite:
		return sqlite.New(
			config.DSN,
			sqlite.MaxConnections(config.MaxConnections),
			sqlite.LogLevel(logLevel),
			sqlite.SlowThreshold(config.SlowThreshold),
		)

	default:
		return nil, fmt.Errorf("unsupported database type: %s. Supported types: mysql, postgres, sqlite", config.Type)
	}
}
