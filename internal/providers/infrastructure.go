package providers

import (
	"os"

	"github.com/google/wire"
	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/pkg/database"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// InfrastructureSet 包含所有基础设施相关的Provider
var InfrastructureSet = wire.NewSet(
	NewLogger,
	NewZerologLogger,
	NewDatabase,
)

// NewLogger 创建logger实例
func NewLogger(cfg *config.Config) logger.Interface {
	return logger.New(cfg.Log.Level)
}

// NewZerologLogger 创建zerolog.Logger实例
func NewZerologLogger(cfg *config.Config) zerolog.Logger {
	// 根据配置设置日志级别
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	return zerolog.New(os.Stdout).
		Level(level).
		With().
		Timestamp().
		Logger()
}

// NewDatabase 创建数据库连接
func NewDatabase(cfg *config.Config) (database.CommonDB, error) {
	dbConfig := database.DatabaseConfig{
		Type:           cfg.DB.Type,
		DSN:            cfg.DB.DSN,
		MaxConnections: cfg.DB.MaxConnections,
	}
	return database.NewDatabase(dbConfig)
}
