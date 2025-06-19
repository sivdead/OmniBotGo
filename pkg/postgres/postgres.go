// Package postgres implements postgres connection.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	customLogger "github.com/sivdead/OmniBotGo/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	_defaultMaxPoolSize   = 1
	_defaultConnAttempts  = 10
	_defaultConnTimeout   = time.Second
	_defaultLogLevel      = logger.Silent
	_defaultSlowThreshold = 200 // 默认200毫秒
)

// Postgres represents PostgreSQL database connection with both GORM and Squirrel support.
type Postgres struct {
	maxPoolSize   int
	connAttempts  int
	connTimeout   time.Duration
	logLevel      logger.LogLevel
	slowThreshold int // 慢查询阈值(毫秒)

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
		maxPoolSize:   _defaultMaxPoolSize,
		connAttempts:  _defaultConnAttempts,
		connTimeout:   _defaultConnTimeout,
		logLevel:      _defaultLogLevel,
		slowThreshold: _defaultSlowThreshold,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	var err error

	// Configure GORM logger with JSON format
	zlogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	gormLogger := customLogger.NewGormLogger(zlogger, pg.logLevel)
	gormLogger.SetSlowThreshold(time.Duration(pg.slowThreshold) * time.Millisecond)

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

// Close closes all PostgreSQL connections.
func (p *Postgres) Close() error {
	if p.Pool != nil {
		p.Pool.Close()
	}
	if p.SqlDB != nil {
		return p.SqlDB.Close()
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

// GetPool returns native pgx pool for advanced PostgreSQL operations
func (p *Postgres) GetPool() *pgxpool.Pool {
	return p.Pool
}
