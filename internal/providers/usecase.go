package providers

import (
	"github.com/google/wire"
	"github.com/sivdead/OmniBotGo/internal/adapter"
	"github.com/sivdead/OmniBotGo/internal/repo"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// UseCaseSet 包含所有usecase相关的Provider
var UseCaseSet = wire.NewSet(
	NewChannelUseCase,
	NewBotUseCase,
	NewMessageUseCase,
	NewAdapterManager,
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
