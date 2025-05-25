// Package database implements database connection factory for different database types.
package database

import (
	"fmt"

	"github.com/evrone/go-clean-template/pkg/mysql"
	"github.com/evrone/go-clean-template/pkg/postgres"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgres"
	PostgresQL DatabaseType = "postgresql" // alias for postgres
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type           string
	DSN            string
	MaxConnections int
}

// NewDatabase creates a database connection based on the specified type
func NewDatabase(config DatabaseConfig) (CommonDB, error) {
	dbType := DatabaseType(config.Type)
	
	switch dbType {
	case MySQL:
		return mysql.New(
			config.DSN,
			mysql.MaxConnections(config.MaxConnections),
		)
		
	case PostgreSQL, PostgresQL:
		return postgres.New(
			config.DSN,
			postgres.MaxPoolSize(config.MaxConnections),
		)
		
	default:
		return nil, fmt.Errorf("unsupported database type: %s. Supported types: mysql, postgres", config.Type)
	}
} 