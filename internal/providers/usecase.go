package providers

import (
	"github.com/google/wire"
	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/adapter"
	"github.com/sivdead/OmniBotGo/internal/repo"
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
	NewAdapterManager,
	NewConnectionManager,
	NewStreamMessageHandler,
)

// NewChannelUseCase 创建ChannelUseCase实例
func NewChannelUseCase(
	channelRepo repo.ChannelRepo,
	messageRepo repo.MessageRepo,
	botRepo repo.BotRepo,
	l logger.Interface,
) usecase.ChannelUseCase {
	return usecase.NewChannelUseCase(channelRepo, messageRepo, botRepo, l)
}

// NewBotUseCase 创建BotUseCase实例
func NewBotUseCase(
	botRepo repo.BotRepo,
	channelRepo repo.ChannelRepo,
	l logger.Interface,
) usecase.BotUseCase {
	return usecase.NewBotUseCase(botRepo, channelRepo, l)
}

// NewMessageUseCase 创建MessageUseCase实例
func NewMessageUseCase(
	messageRepo repo.MessageRepo,
	channelRepo repo.ChannelRepo,
	adapterManager *adapter.Manager,
	l logger.Interface,
) usecase.MessageUseCase {
	return usecase.NewMessageUseCase(messageRepo, channelRepo, adapterManager, l)
}

// NewAdapterManager 创建适配器管理器实例
func NewAdapterManager() *adapter.Manager {
	return adapter.NewManager()
}

// NewConnectionManager 创建连接管理器实例
func NewConnectionManager(
	logger zerolog.Logger,
	channelRepo repo.ChannelRepo,
	adapterManager *adapter.Manager,
	messageHandler port.MessageHandler,
) *service.ConnectionManager {
	return service.NewConnectionManager(logger, channelRepo, adapterManager, messageHandler)
}

// NewStreamMessageHandler 创建Stream消息处理器
func NewStreamMessageHandler(
	messageUseCase usecase.MessageUseCase,
) port.MessageHandler {
	return messageUseCase.CreateStreamMessageHandler()
}
