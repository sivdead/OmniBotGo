package providers

import (
	"github.com/google/wire"
	"github.com/sivdead/OmniBotGo/internal/repo/persistent"
	"github.com/sivdead/OmniBotGo/pkg/database"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// RepositorySet 包含所有repository相关的Provider
var RepositorySet = wire.NewSet(
	// 直接提供Repository实现
	persistent.NewBotRepository,
	persistent.NewChannelRepository,
	persistent.NewMessageRepository,
	persistent.NewMessageProcessorRepository,
	persistent.NewMessageRoutingRuleRepository,
	persistent.NewSystemConfigRepository,
	persistent.NewMessageQueueRepository,
	persistent.NewConnectionLogRepository,
	persistent.NewAPICallLogRepository,
	NewDataSeeder,
)

// NewDataSeeder 创建DataSeeder实例
func NewDataSeeder(db database.CommonDB, logger logger.Interface) *persistent.DataSeeder {
	return persistent.NewDataSeeder(db, logger)
}
