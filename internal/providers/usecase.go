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
	NewSystemConfigUseCase,
	NewPlatformUseCase,
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
	logger logger.Interface,
) usecase.MessageUseCase {
	return usecase.NewMessageUseCase(
		messageRepo,
		channelRepo,
		adapterManager,
		routingUC,
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

// NewSystemConfigUseCase 创建SystemConfigUseCase实例
func NewSystemConfigUseCase(
	systemConfigRepo port.SystemConfigRepository,
	logger logger.Interface,
) usecase.SystemConfigUseCase {
	// TODO: 实现SystemConfigUseCase
	return nil
}

// NewPlatformUseCase 创建PlatformUseCase实例
func NewPlatformUseCase(
	adapterManager port.AdapterManager,
	logger logger.Interface,
) usecase.PlatformUseCase {
	// TODO: 实现PlatformUseCase
	return nil
}

// NewMonitorUseCase 创建MonitorUseCase实例
func NewMonitorUseCase(
	logger logger.Interface,
) usecase.MonitorUseCase {
	// TODO: 实现MonitorUseCase
	return nil
}

// NewLogUseCase 创建LogUseCase实例
func NewLogUseCase(
	logRepo port.ConnectionLogRepository,
	apiCallLogRepo port.APICallLogRepository,
	logger logger.Interface,
) usecase.LogUseCase {
	// TODO: 实现LogUseCase
	return nil
}

// NewQueueUseCase 创建QueueUseCase实例
func NewQueueUseCase(
	queueRepo port.MessageQueueRepository,
	logger logger.Interface,
) usecase.QueueUseCase {
	// TODO: 实现QueueUseCase
	return nil
}

// NewProcessorUseCase 创建ProcessorUseCase实例
func NewProcessorUseCase(
	processorRepo port.MessageProcessorRepository,
	routingRuleRepo port.MessageRoutingRuleRepository,
	logger logger.Interface,
) usecase.ProcessorUseCase {
	// TODO: 实现ProcessorUseCase
	return nil
}
