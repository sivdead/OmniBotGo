// Package database defines common interfaces for database operations.
package database

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"gorm.io/gorm"
)

// CommonDB defines a common interface for database operations
// Both MySQL and PostgreSQL implementations should satisfy this interface
type CommonDB interface {
	// GetGORM returns GORM database instance for ORM operations
	GetGORM() *gorm.DB
	
	// GetSquirrel returns Squirrel query builder for complex queries
	GetSquirrel() squirrel.StatementBuilderType
	
	// GetSqlDB returns raw SQL database connection
	GetSqlDB() *sql.DB
	
	// Close closes the database connection
	Close() error
} 