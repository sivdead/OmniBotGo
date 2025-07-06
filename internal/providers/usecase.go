package providers

import (
	"github.com/google/wire"
	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/service"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// UseCaseSet 包含所有usecase相关的Provider
var UseCaseSet = wire.NewSet(
	NewChannelUseCase,
	NewBotUseCase,
	NewMessageUseCase,
	ProvideSystemConfigUseCase,
	ProvidePlatformUseCase,
	NewMonitorUseCase,
	NewLogUseCase,
	NewQueueUseCase,
	NewProcessorUseCase,
	NewConnectionManager,
	NewRoutingUseCase,
)

// NewChannelUseCase 创建ChannelUseCase实例
func NewChannelUseCase(
	channelRepo port.ChannelRepository,
	botRepo port.BotRepository,
	adapterManager port.AdapterManager,
	logger logger.Interface,
) usecase.ChannelUseCase {
	return usecase.NewChannelUseCase(channelRepo, botRepo, adapterManager, logger)
}

// NewBotUseCase 创建BotUseCase实例
func NewBotUseCase(
	botRepo port.BotRepository,
	channelRepo port.ChannelRepository,
	logger logger.Interface,
) usecase.BotUseCase {
	return usecase.NewBotUseCase(botRepo, channelRepo, logger)
}

// NewMessageUseCase 创建MessageUseCase实例
func NewMessageUseCase(
	messageRepo port.MessageRepository,
	channelRepo port.ChannelRepository,
	adapterManager port.AdapterManager,
	routingUC usecase.RoutingUseCase,
	queueRepo port.MessageQueueRepository,
	logger logger.Interface,
) usecase.MessageUseCase {
	return usecase.NewMessageUseCase(
		messageRepo,
		channelRepo,
		adapterManager,
		routingUC,
		queueRepo,
		logger,
	)
}

// NewConnectionManager 创建连接管理器实例
func NewConnectionManager(
	logger zerolog.Logger,
	channelRepo port.ChannelRepository,
	adapterManager port.AdapterManager,
	messageUseCase usecase.MessageUseCase,
) *service.ConnectionManager {
	return service.NewConnectionManager(logger, channelRepo, adapterManager, messageUseCase)
}

// NewRoutingUseCase 创建路由UseCase
func NewRoutingUseCase(
	routingRuleRepo port.MessageRoutingRuleRepository,
	messageProcessorRepo port.MessageProcessorRepository,
	channelRepo port.ChannelRepository,
	logger logger.Interface,
) usecase.RoutingUseCase {
	return usecase.NewRoutingUseCase(
		routingRuleRepo,
		messageProcessorRepo,
		channelRepo,
		logger,
	)
}

// ProvideSystemConfigUseCase 提供系统配置用例
func ProvideSystemConfigUseCase(
	repo port.SystemConfigRepository,
	logger logger.Interface,
) usecase.SystemConfigUseCase {
	return usecase.NewSystemConfigUseCase(repo, logger)
}

// ProvidePlatformUseCase 提供平台管理用例
func ProvidePlatformUseCase(
	adapterManager port.AdapterManager,
	channelRepo port.ChannelRepository,
	logger logger.Interface,
) usecase.PlatformUseCase {
	return usecase.NewPlatformUseCase(adapterManager, channelRepo, logger)
}

// NewMonitorUseCase 创建MonitorUseCase实例
func NewMonitorUseCase(
	botRepo port.BotRepository,
	channelRepo port.ChannelRepository,
	messageRepo port.MessageRepository,
	queueRepo port.MessageQueueRepository,
	logger logger.Interface,
) usecase.MonitorUseCase {
	return usecase.NewMonitorUseCase(botRepo, channelRepo, messageRepo, queueRepo, logger)
}

// NewLogUseCase 创建LogUseCase实例
func NewLogUseCase(
	logRepo port.ConnectionLogRepository,
	apiCallLogRepo port.APICallLogRepository,
	logger logger.Interface,
) usecase.LogUseCase {
	return usecase.NewLogUseCase(logRepo, apiCallLogRepo, logger)
}

// NewQueueUseCase 创建QueueUseCase实例
func NewQueueUseCase(
	queueRepo port.MessageQueueRepository,
	logger logger.Interface,
) usecase.QueueUseCase {
	return usecase.NewQueueUseCase(queueRepo, logger)
}

// NewProcessorUseCase 创建ProcessorUseCase实例
func NewProcessorUseCase(
	processorRepo port.MessageProcessorRepository,
	routingRuleRepo port.MessageRoutingRuleRepository,
	logger logger.Interface,
) usecase.ProcessorUseCase {
	return usecase.NewProcessorUseCase(processorRepo, routingRuleRepo, logger)
}
