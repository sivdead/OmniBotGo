package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// channelUseCase 通道管理业务逻辑实现
type channelUseCase struct {
	channelRepo    port.ChannelRepository
	botRepo        port.BotRepository
	adapterManager port.AdapterManager
	logger         logger.Interface
}

// NewChannelUseCase 创建通道管理业务逻辑实例
func NewChannelUseCase(
	channelRepo port.ChannelRepository,
	botRepo port.BotRepository,
	adapterManager port.AdapterManager,
	logger logger.Interface,
) ChannelUseCase {
	return &channelUseCase{
		channelRepo:    channelRepo,
		botRepo:        botRepo,
		adapterManager: adapterManager,
		logger:         logger,
	}
}

// CreateChannel 创建通道
func (uc *channelUseCase) CreateChannel(ctx context.Context, req CreateChannelRequest) (*entity.Channel, error) {
	uc.logger.Info("开始创建通道", "method", "CreateChannel", "bot_id", req.BotID, "platform_type", req.PlatformType, "channel_name", req.ChannelName)

	// 验证机器人是否存在且活跃
	bot, err := uc.botRepo.GetByID(ctx, req.BotID)
	if err != nil {
		uc.logger.Error("获取机器人信息失败", "error", err)
		return nil, fmt.Errorf("获取机器人信息失败: %w", err)
	}

	if !bot.IsActive() {
		uc.logger.Warn("机器人未激活，无法创建通道")
		return nil, fmt.Errorf("机器人未激活")
	}

	// 检查通道名称是否重复
	existingChannels, err := uc.channelRepo.GetByBotID(ctx, req.BotID)
	if err != nil {
		uc.logger.Error("获取机器人通道列表失败", "error", err)
		return nil, fmt.Errorf("获取机器人通道列表失败: %w", err)
	}

	for _, ch := range existingChannels {
		if ch.ChannelName == req.ChannelName && ch.Status != entity.StatusDeleted {
			uc.logger.Warn("通道名称已存在")
			return nil, fmt.Errorf("通道名称已存在")
		}
	}

	// 创建通道实体
	channel := &entity.Channel{
		BotID:            req.BotID,
		PlatformType:     req.PlatformType,
		ChannelName:      req.ChannelName,
		ConnectionStatus: entity.ConnectionStatusDisconnected,
		Status:           entity.StatusActive,
	}

	// 设置基础信息
	now := time.Now()
	channel.CreatedAt = now
	channel.UpdatedAt = now

	// 设置配置信息
	if req.Config != nil {
		channel.Config = entity.JSONField(req.Config)
	}

	// 生成Webhook路径
	channel.WebhookPath = fmt.Sprintf("/webhook/%s/%d/%d", req.PlatformType, req.BotID, time.Now().Unix())

	// 验证通道数据
	if err := channel.Validate(); err != nil {
		uc.logger.Error("通道数据验证失败", "error", err)
		return nil, fmt.Errorf("通道数据验证失败: %w", err)
	}

	// 保存通道到数据库
	if err := uc.channelRepo.Create(ctx, channel); err != nil {
		uc.logger.Error("保存通道失败", "error", err)
		return nil, fmt.Errorf("保存通道失败: %w", err)
	}

	uc.logger.Info("通道创建成功", "channel_id", channel.ID, "webhook_path", channel.WebhookPath)

	return channel, nil
}

// UpdateChannel 更新通道
func (uc *channelUseCase) UpdateChannel(ctx context.Context, req UpdateChannelRequest) (*entity.Channel, error) {
	uc.logger.Info("开始更新通道", "method", "UpdateChannel", "channel_id", req.ID)

	// 获取现有通道
	channel, err := uc.channelRepo.GetByID(ctx, req.ID)
	if err != nil {
		uc.logger.Error("获取通道信息失败", "error", err)
		return nil, fmt.Errorf("获取通道信息失败: %w", err)
	}

	// 更新通道名称
	if req.ChannelName != nil {
		// 检查新名称是否重复
		if *req.ChannelName != channel.ChannelName {
			existingChannels, err := uc.channelRepo.GetByBotID(ctx, channel.BotID)
			if err != nil {
				uc.logger.Error("获取机器人通道列表失败", "error", err)
				return nil, fmt.Errorf("获取机器人通道列表失败: %w", err)
			}

			for _, ch := range existingChannels {
				if ch.ChannelName == *req.ChannelName && ch.ID != channel.ID && ch.Status != entity.StatusDeleted {
					uc.logger.Warn("通道名称已存在")
					return nil, fmt.Errorf("通道名称已存在")
				}
			}
		}
		channel.ChannelName = *req.ChannelName
	}

	// 更新配置信息
	if req.Config != nil {
		if channel.Config == nil {
			channel.Config = make(entity.JSONField)
		}
		for key, value := range req.Config {
			channel.Config.Set(key, value)
		}
	}

	// 更新状态
	if req.Status != nil {
		channel.Status = *req.Status
	}

	// 更新时间戳
	channel.UpdatedAt = time.Now()

	// 验证通道数据
	if err := channel.Validate(); err != nil {
		uc.logger.Error("通道数据验证失败", "error", err)
		return nil, fmt.Errorf("通道数据验证失败: %w", err)
	}

	// 保存更新
	if err := uc.channelRepo.Update(ctx, channel); err != nil {
		uc.logger.Error("更新通道失败", "error", err)
		return nil, fmt.Errorf("更新通道失败: %w", err)
	}

	uc.logger.Info("通道更新成功")
	return channel, nil
}

// DeleteChannel 删除通道
func (uc *channelUseCase) DeleteChannel(ctx context.Context, id int64) error {
	uc.logger.Info("开始删除通道", "method", "DeleteChannel", "channel_id", id)

	// 获取通道信息
	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取通道信息失败", "error", err)
		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	// 检查通道是否正在连接中
	if channel.IsConnected() {
		uc.logger.Warn("通道正在连接中，无法删除")
		return fmt.Errorf("通道正在连接中，请先断开连接")
	}

	// 软删除通道（标记为已删除状态）
	channel.Status = entity.StatusDeleted
	channel.UpdatedAt = time.Now()

	if err := uc.channelRepo.Update(ctx, channel); err != nil {
		uc.logger.Error("删除通道失败", "error", err)
		return fmt.Errorf("删除通道失败: %w", err)
	}

	uc.logger.Info("通道删除成功")
	return nil
}

// GetChannel 获取通道信息
func (uc *channelUseCase) GetChannel(ctx context.Context, id int64) (*entity.Channel, error) {
	uc.logger.Info("获取通道信息", "method", "GetChannel", "channel_id", id)

	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取通道信息失败", "error", err)
		return nil, fmt.Errorf("获取通道信息失败: %w", err)
	}

	uc.logger.Info("通道信息获取成功", "channel_name", channel.ChannelName, "platform_type", channel.PlatformType, "status", channel.Status.String(), "connection_status", channel.ConnectionStatus.String())

	return channel, nil
}

// ListChannels 获取通道列表
func (uc *channelUseCase) ListChannels(ctx context.Context, params ListChannelsParams) (*ChannelListResult, error) {
	uc.logger.Info("获取通道列表", "method", "ListChannels")

	// 构建查询过滤器
	filters := make(map[string]interface{})

	if params.BotID != nil {
		filters["bot_id"] = *params.BotID
	}
	if params.PlatformType != nil {
		filters["platform_type"] = *params.PlatformType
	}
	if params.Status != nil {
		filters["status"] = *params.Status
	}

	// 查询通道列表
	listParams := port.ListParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		Filters:  filters,
	}

	result, err := uc.channelRepo.List(ctx, listParams)
	if err != nil {
		uc.logger.Error("查询通道列表失败", "error", err)
		return nil, fmt.Errorf("查询通道列表失败: %w", err)
	}

	// 转换指针切片为值切片
	channelValues := make([]entity.Channel, len(result.Items))
	for i, ch := range result.Items {
		channelValues[i] = *ch
	}

	channelResult := &ChannelListResult{
		Items:      channelValues,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}

	uc.logger.Info("通道列表查询完成", "total", result.Total, "page", result.Page, "page_size", result.PageSize)

	return channelResult, nil
}

// UpdateChannelStatus 更新通道状态
func (uc *channelUseCase) UpdateChannelStatus(ctx context.Context, id int64, status entity.ConnectionStatus) error {
	uc.logger.Info("更新通道连接状态", "method", "UpdateChannelStatus", "channel_id", id, "status", status.String())

	// 获取通道信息
	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取通道信息失败", "error", err)
		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	// 更新连接状态
	channel.UpdateConnectionStatus(status)
	channel.UpdatedAt = time.Now()

	// 保存更新
	if err := uc.channelRepo.Update(ctx, channel); err != nil {
		uc.logger.Error("更新通道状态失败", "error", err)
		return fmt.Errorf("更新通道状态失败: %w", err)
	}

	uc.logger.Info("通道连接状态更新成功")
	return nil
}

// RefreshChannelToken 刷新通道令牌
func (uc *channelUseCase) RefreshChannelToken(ctx context.Context, id int64) error {
	uc.logger.Info("刷新通道令牌", "method", "RefreshChannelToken", "channel_id", id)

	// 获取通道信息
	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取通道信息失败", "error", err)
		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	// TODO: 根据平台类型调用相应的令牌刷新逻辑
	// 这里应该根据 channel.PlatformType 选择对应的平台API来刷新令牌

	// 模拟令牌刷新
	newToken := fmt.Sprintf("token_%d_%d", id, time.Now().Unix())
	expiresAt := time.Now().Add(2 * time.Hour)

	channel.AccessToken = newToken
	channel.AccessTokenExpiresAt = &expiresAt
	channel.UpdatedAt = time.Now()

	// 保存更新
	if err := uc.channelRepo.Update(ctx, channel); err != nil {
		uc.logger.Error("保存令牌更新失败", "error", err)
		return fmt.Errorf("保存令牌更新失败: %w", err)
	}

	uc.logger.Info("通道令牌刷新成功", "new_token", "***", "expires_at", expiresAt)

	return nil
}

// GetActiveChannels 获取所有活跃的通道
func (uc *channelUseCase) GetActiveChannels(ctx context.Context) ([]*entity.Channel, error) {
	uc.logger.Info("获取活跃通道列表", "method", "GetActiveChannels")

	// 查询所有状态为激活的通道
	filters := map[string]interface{}{
		"status": entity.StatusActive,
	}

	listParams := port.ListParams{
		Filters:  filters,
		Page:     1,
		PageSize: 1000, // 获取所有活跃通道，设置一个较大的值
	}

	result, err := uc.channelRepo.List(ctx, listParams)
	if err != nil {
		uc.logger.Error("获取活跃通道列表失败", "error", err)
		return nil, fmt.Errorf("获取活跃通道列表失败: %w", err)
	}

	return result.Items, nil
}

// IsChannelConnected 检查通道是否已连接
func (uc *channelUseCase) IsChannelConnected(ctx context.Context, id int64) bool {
	uc.logger.Debug("检查通道连接状态", "method", "IsChannelConnected", "channel_id", id)

	// 获取通道信息
	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取通道信息失败", "error", err, "channel_id", id)
		return false
	}

	// 检查通道是否激活且已连接
	return channel.IsActive() && channel.IsConnected()
}
