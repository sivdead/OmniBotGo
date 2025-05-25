// Package postgres implements postgres connection.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres represents PostgreSQL database connection with both GORM and Squirrel support.
type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	// GORM instance for ORM operations (similar to MySQL implementation)
	DB *gorm.DB
	
	// Squirrel query builder for complex queries
	Builder squirrel.StatementBuilderType
	
	// Native pgx connection pool
	Pool *pgxpool.Pool
	
	// Raw SQL database connection (for Squirrel)
	SqlDB *sql.DB
}

// New creates a new PostgreSQL connection with both GORM and Squirrel support.
func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	var err error

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Silent)

	// Initialize GORM connection
	for pg.connAttempts > 0 {
		pg.DB, err = gorm.Open(postgres.Open(url), &gorm.Config{
			Logger: gormLogger,
		})
		if err == nil {
			break
		}

		log.Printf("PostgreSQL GORM is trying to connect, attempts left: %d", pg.connAttempts)
		time.Sleep(pg.connTimeout)
		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - New - GORM connection failed: %w", err)
	}

	// Get underlying sql.DB for both connection pool configuration and Squirrel
	pg.SqlDB, err = pg.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("postgres - New - get sql.DB: %w", err)
	}

	// Configure connection pool
	pg.SqlDB.SetMaxOpenConns(pg.maxPoolSize)
	pg.SqlDB.SetMaxIdleConns(pg.maxPoolSize / 2)
	pg.SqlDB.SetConnMaxLifetime(time.Hour)
	pg.SqlDB.SetConnMaxIdleTime(30 * time.Minute)

	// Initialize Squirrel builder with PostgreSQL syntax
	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Initialize native pgx pool for advanced PostgreSQL features (optional)
	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres - New - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Printf("Warning: PostgreSQL native pool connection failed: %v", err)
		// Don't fail completely if pgx pool fails, GORM is enough for most cases
	}

	// Test GORM connection
	if err = pg.SqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("postgres - New - ping failed: %w", err)
	}

	return pg, nil
}

// Close closes both GORM and pgx connections.
func (p *Postgres) Close() error {
	var errs []error

	// Close GORM connection
	if p.SqlDB != nil {
		if err := p.SqlDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close GORM connection: %w", err))
		}
	}

	// Close pgx pool
	if p.Pool != nil {
		p.Pool.Close()
	}

	if len(errs) > 0 {
		return fmt.Errorf("postgres close errors: %v", errs)
	}

	return nil
}

// GetGORM returns GORM database instance for ORM operations
func (p *Postgres) GetGORM() *gorm.DB {
	return p.DB
}

// GetSquirrel returns Squirrel query builder for complex queries
func (p *Postgres) GetSquirrel() squirrel.StatementBuilderType {
	return p.Builder
}

// GetSqlDB returns raw SQL database connection
func (p *Postgres) GetSqlDB() *sql.DB {
	return p.SqlDB
}
