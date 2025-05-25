// Package mysql implements MySQL connection with GORM and Squirrel integration.
package mysql

import (
	"database/sql"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	_defaultMaxConnections    = 10
	_defaultMaxIdleConns      = 5
	_defaultConnMaxLifetime   = time.Hour
	_defaultConnMaxIdleTime   = time.Minute * 30
	_defaultConnAttempts      = 10
	_defaultConnTimeout       = time.Second
)

// MySQL represents MySQL database connection with both GORM and Squirrel support.
type MySQL struct {
	maxConnections    int
	maxIdleConns      int
	connMaxLifetime   time.Duration
	connMaxIdleTime   time.Duration
	connAttempts      int
	connTimeout       time.Duration

	// GORM instance for ORM operations
	DB *gorm.DB
	
	// Squirrel query builder for complex queries
	Builder squirrel.StatementBuilderType
	
	// Raw SQL database connection (for Squirrel)
	SqlDB *sql.DB
}

// New creates a new MySQL connection with both GORM and Squirrel support.
func New(dsn string, opts ...Option) (*MySQL, error) {
	m := &MySQL{
		maxConnections:  _defaultMaxConnections,
		maxIdleConns:    _defaultMaxIdleConns,
		connMaxLifetime: _defaultConnMaxLifetime,
		connMaxIdleTime: _defaultConnMaxIdleTime,
		connAttempts:    _defaultConnAttempts,
		connTimeout:     _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(m)
	}

	var err error
	
	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Silent)

	for m.connAttempts > 0 {
		m.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
		if err == nil {
			break
		}

		log.Printf("MySQL is trying to connect, attempts left: %d", m.connAttempts)
		time.Sleep(m.connTimeout)
		m.connAttempts--
	}

	if err != nil {
		return nil, err
	}

	// Get underlying sql.DB for both connection pool configuration and Squirrel
	m.SqlDB, err = m.DB.DB()
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	m.SqlDB.SetMaxOpenConns(m.maxConnections)
	m.SqlDB.SetMaxIdleConns(m.maxIdleConns)
	m.SqlDB.SetConnMaxLifetime(m.connMaxLifetime)
	m.SqlDB.SetConnMaxIdleTime(m.connMaxIdleTime)

	// Initialize Squirrel builder with MySQL syntax
	m.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	// Test connection
	if err = m.SqlDB.Ping(); err != nil {
		return nil, err
	}

	return m, nil
}

// Close closes the MySQL connection.
func (m *MySQL) Close() error {
	if m.SqlDB != nil {
		return m.SqlDB.Close()
	}
	return nil
}

// GetGORM returns GORM database instance for ORM operations
func (m *MySQL) GetGORM() *gorm.DB {
	return m.DB
}

// GetSquirrel returns Squirrel query builder for complex queries
func (m *MySQL) GetSquirrel() squirrel.StatementBuilderType {
	return m.Builder
}

// GetSqlDB returns raw SQL database connection
func (m *MySQL) GetSqlDB() *sql.DB {
	return m.SqlDB
} 