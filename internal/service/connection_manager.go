package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/adapter"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/repo"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

// ConnectionManager 连接管理器，负责管理所有主动连接型适配器的生命周期
type ConnectionManager struct {
	logger         zerolog.Logger
	channelRepo    repo.ChannelRepo
	adapterManager *adapter.Manager
	messageHandler port.MessageHandler
	connections    map[int64]ConnectionInfo // channelID -> ConnectionInfo
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	ChannelID     int64
	PlatformType  string
	StreamAdapter port.StreamAdapter
	Config        map[string]interface{}
	IsConnected   bool
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(
	logger zerolog.Logger,
	channelRepo repo.ChannelRepo,
	adapterManager *adapter.Manager,
	messageHandler port.MessageHandler,
) *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConnectionManager{
		logger:         logger,
		channelRepo:    channelRepo,
		adapterManager: adapterManager,
		messageHandler: messageHandler,
		connections:    make(map[int64]ConnectionInfo),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start 启动连接管理器，加载并启动所有活跃的Stream连接
func (cm *ConnectionManager) Start(ctx context.Context) error {
	cm.logger.Info().Msg("starting connection manager")

	// 从数据库加载所有活跃的通道
	channels, err := cm.channelRepo.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to load active channels: %w", err)
	}

	cm.logger.Info().Int("count", len(channels)).Msg("loaded active channels")

	// 为每个通道启动连接
	for _, channel := range channels {
		if err := cm.startChannelConnection(ctx, channel); err != nil {
			cm.logger.Error().
				Err(err).
				Int64("channel_id", channel.ID).
				Str("platform_type", channel.PlatformType).
				Msg("failed to start channel connection")
		}
	}

	return nil
}

// Stop 停止连接管理器，优雅关闭所有连接
func (cm *ConnectionManager) Stop(ctx context.Context) error {
	cm.logger.Info().Msg("stopping connection manager")

	cm.cancel() // 取消内部context

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 停止所有连接
	for channelID, connInfo := range cm.connections {
		if connInfo.IsConnected {
			if err := connInfo.StreamAdapter.Stop(ctx); err != nil {
				cm.logger.Error().
					Err(err).
					Int64("channel_id", channelID).
					Msg("failed to stop stream connection")
			} else {
				cm.logger.Info().
					Int64("channel_id", channelID).
					Msg("stream connection stopped")
			}
		}
	}

	// 清空连接信息
	cm.connections = make(map[int64]ConnectionInfo)

	cm.logger.Info().Msg("connection manager stopped")
	return nil
}

// startChannelConnection 为单个通道启动连接
func (cm *ConnectionManager) startChannelConnection(ctx context.Context, channel *entity.Channel) error {
	// 获取适配器
	platformAdapter, err := cm.adapterManager.GetAdapter(entity.PlatformType(channel.PlatformType))
	if err != nil {
		return fmt.Errorf("failed to get adapter for platform %s: %w", channel.PlatformType, err)
	}

	// 检查适配器是否支持Stream模式
	streamAdapter, ok := platformAdapter.(port.StreamAdapter)
	if !ok {
		cm.logger.Debug().
			Int64("channel_id", channel.ID).
			Str("platform_type", channel.PlatformType).
			Msg("platform does not support stream mode, skipping")
		return nil
	}

	// 创建连接信息
	connInfo := ConnectionInfo{
		ChannelID:     channel.ID,
		PlatformType:  channel.PlatformType,
		StreamAdapter: streamAdapter,
		Config:        channel.Config,
		IsConnected:   false,
	}

	// 启动Stream连接
	if err := streamAdapter.Start(cm.ctx, cm.messageHandler, channel.Config); err != nil {
		return fmt.Errorf("failed to start stream connection: %w", err)
	}

	connInfo.IsConnected = true

	// 保存连接信息
	cm.mu.Lock()
	cm.connections[channel.ID] = connInfo
	cm.mu.Unlock()

	cm.logger.Info().
		Int64("channel_id", channel.ID).
		Str("platform_type", channel.PlatformType).
		Msg("stream connection started")

	return nil
}

// GetConnectionStatus 获取连接状态
func (cm *ConnectionManager) GetConnectionStatus(channelID int64) (bool, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	connInfo, exists := cm.connections[channelID]
	if !exists {
		return false, fmt.Errorf("connection not found for channel %d", channelID)
	}

	return connInfo.IsConnected && connInfo.StreamAdapter.IsConnected(), nil
}

// RestartConnection 重启指定通道的连接
func (cm *ConnectionManager) RestartConnection(ctx context.Context, channelID int64) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	connInfo, exists := cm.connections[channelID]
	if !exists {
		return fmt.Errorf("connection not found for channel %d", channelID)
	}

	// 停止现有连接
	if connInfo.IsConnected {
		if err := connInfo.StreamAdapter.Stop(ctx); err != nil {
			cm.logger.Error().
				Err(err).
				Int64("channel_id", channelID).
				Msg("failed to stop connection during restart")
		}
	}

	// 重新启动连接
	if err := connInfo.StreamAdapter.Start(cm.ctx, cm.messageHandler, connInfo.Config); err != nil {
		connInfo.IsConnected = false
		cm.connections[channelID] = connInfo
		return fmt.Errorf("failed to restart stream connection: %w", err)
	}

	connInfo.IsConnected = true
	cm.connections[channelID] = connInfo

	cm.logger.Info().
		Int64("channel_id", channelID).
		Msg("stream connection restarted")

	return nil
}
