package providers

import (
	"github.com/google/wire"
	"github.com/sivdead/OmniBotGo/internal/repo"
	"github.com/sivdead/OmniBotGo/internal/repo/persistent"
	"github.com/sivdead/OmniBotGo/pkg/database"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// RepositorySet 包含所有repository相关的Provider
var RepositorySet = wire.NewSet(
	NewBotRepo,
	NewChannelRepo,
	NewMessageRepo,
	NewMessageProcessorRepo,
	NewMessageRoutingRuleRepo,
	NewSystemConfigRepo,
	NewMessageQueueRepo,
	NewConnectionLogRepo,
	NewAPICallLogRepo,
	NewDataSeeder,

	// 绑定接口到实现
	wire.Bind(new(repo.BotRepo), new(*persistent.BotRepo)),
	wire.Bind(new(repo.ChannelRepo), new(*persistent.ChannelRepo)),
	wire.Bind(new(repo.MessageRepo), new(*persistent.MessageRepo)),
	wire.Bind(new(repo.MessageProcessorRepo), new(*persistent.MessageProcessorRepo)),
	wire.Bind(new(repo.MessageRoutingRuleRepo), new(*persistent.MessageRoutingRuleRepo)),
	wire.Bind(new(repo.SystemConfigRepo), new(*persistent.SystemConfigRepo)),
	wire.Bind(new(repo.MessageQueueRepo), new(*persistent.MessageQueueRepo)),
	wire.Bind(new(repo.ConnectionLogRepo), new(*persistent.ConnectionLogRepo)),
	wire.Bind(new(repo.APICallLogRepo), new(*persistent.APICallLogRepo)),
)

// NewBotRepo 创建BotRepo实例
func NewBotRepo(db database.CommonDB) *persistent.BotRepo {
	return persistent.NewBotRepo(db).(*persistent.BotRepo)
}

// NewChannelRepo 创建ChannelRepo实例
func NewChannelRepo(db database.CommonDB) *persistent.ChannelRepo {
	return persistent.NewChannelRepo(db).(*persistent.ChannelRepo)
}

// NewMessageRepo 创建MessageRepo实例
func NewMessageRepo(db database.CommonDB) *persistent.MessageRepo {
	return persistent.NewMessageRepo(db).(*persistent.MessageRepo)
}

// NewMessageProcessorRepo 创建MessageProcessorRepo实例
func NewMessageProcessorRepo(db database.CommonDB) *persistent.MessageProcessorRepo {
	return persistent.NewMessageProcessorRepo(db).(*persistent.MessageProcessorRepo)
}

// NewMessageRoutingRuleRepo 创建MessageRoutingRuleRepo实例
func NewMessageRoutingRuleRepo(db database.CommonDB) *persistent.MessageRoutingRuleRepo {
	return persistent.NewMessageRoutingRuleRepo(db).(*persistent.MessageRoutingRuleRepo)
}

// NewSystemConfigRepo 创建SystemConfigRepo实例
func NewSystemConfigRepo(db database.CommonDB) *persistent.SystemConfigRepo {
	return persistent.NewSystemConfigRepo(db).(*persistent.SystemConfigRepo)
}

// NewMessageQueueRepo 创建MessageQueueRepo实例
func NewMessageQueueRepo(db database.CommonDB) *persistent.MessageQueueRepo {
	return persistent.NewMessageQueueRepo(db).(*persistent.MessageQueueRepo)
}

// NewConnectionLogRepo 创建ConnectionLogRepo实例
func NewConnectionLogRepo(db database.CommonDB) *persistent.ConnectionLogRepo {
	return persistent.NewConnectionLogRepo(db).(*persistent.ConnectionLogRepo)
}

// NewAPICallLogRepo 创建APICallLogRepo实例
func NewAPICallLogRepo(db database.CommonDB) *persistent.APICallLogRepo {
	return persistent.NewAPICallLogRepo(db).(*persistent.APICallLogRepo)
}

// NewDataSeeder 创建DataSeeder实例
func NewDataSeeder(db database.CommonDB, logger logger.Interface) *persistent.DataSeeder {
	return persistent.NewDataSeeder(db, logger)
}
