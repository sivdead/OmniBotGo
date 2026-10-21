package sqlite

import (
	"database/sql"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	// 使用纯Go SQLite驱动，无需CGO
	_ "modernc.org/sqlite"
)

const (
	_defaultMaxConnections = 10
	_defaultSlowThreshold  = 200 // 200ms
)

// SQLite -.
type SQLite struct {
	maxConnections int
	logLevel       gormlogger.LogLevel
	slowThreshold  int
	gormDB         *gorm.DB
	sqlDB          *sql.DB
	builder        squirrel.StatementBuilderType
}

// New -.
func New(dsn string, opts ...Option) (*SQLite, error) {
	s := &SQLite{
		maxConnections: _defaultMaxConnections,
		logLevel:       gormlogger.Silent,
		slowThreshold:  _defaultSlowThreshold,
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// 创建zerolog实例
	zlogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	gormLogger := logger.NewGormLogger(zlogger, s.logLevel)
	gormLogger.SetSlowThreshold(time.Duration(s.slowThreshold) * time.Millisecond)

	// 创建GORM配置
	gormConfig := &gorm.Config{
		Logger: gormLogger,
	}

	// E08: enable foreign_keys on every physical connection via modernc _pragma DSN,
	// not only the first pooled connection from a one-shot PRAGMA Exec.
	dsn = ensureForeignKeysPragma(dsn)

	// 使用modernc.org/sqlite作为驱动
	dialector := sqlite.Dialector{
		DSN:        dsn,
		DriverName: "sqlite",
	}
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, err
	}

	// SQLite特殊配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SQLite连接池配置
	sqlDB.SetMaxOpenConns(s.maxConnections)
	sqlDB.SetMaxIdleConns(s.maxConnections)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Best-effort on the current connection; DSN pragma covers subsequent pool conns.
	db.Exec("PRAGMA foreign_keys = ON")
	db.Exec("PRAGMA journal_mode = WAL")

	s.gormDB = db
	s.sqlDB = sqlDB

	// 初始化Squirrel构建器，SQLite使用question placeholder
	s.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	return s, nil
}

// GetDB -.
func (s *SQLite) GetDB() *gorm.DB {
	return s.gormDB
}

// GetGORM 返回GORM数据库实例用于ORM操作
func (s *SQLite) GetGORM() *gorm.DB {
	return s.gormDB
}

// GetSquirrel 返回Squirrel查询构建器用于复杂查询
func (s *SQLite) GetSquirrel() squirrel.StatementBuilderType {
	return s.builder
}

// GetSqlDB 返回原始SQL数据库连接
func (s *SQLite) GetSqlDB() *sql.DB {
	return s.sqlDB
}

// Close -.
func (s *SQLite) Close() error {
	if s.sqlDB != nil {
		return s.sqlDB.Close()
	}

	return nil
}

// ensureForeignKeysPragma appends modernc.org/sqlite _pragma=foreign_keys(1) when absent.
func ensureForeignKeysPragma(dsn string) string {
	if dsn == "" {
		return dsn
	}

	lower := strings.ToLower(dsn)
	if strings.Contains(lower, "_pragma=foreign_keys") || strings.Contains(lower, "foreign_keys(1)") {
		return dsn
	}

	if dsn == ":memory:" {
		return "file::memory:?cache=shared&_pragma=foreign_keys(1)"
	}

	if strings.Contains(dsn, "?") {
		return dsn + "&_pragma=foreign_keys(1)"
	}

	return dsn + "?_pragma=foreign_keys(1)"
}
