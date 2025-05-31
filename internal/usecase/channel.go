package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/repo"
)

// channelUseCase 通道管理业务逻辑实现
type channelUseCase struct {
	channelRepo repo.ChannelRepo
	botRepo     repo.BotRepo
	logger      *zerolog.Logger
}

// NewChannelUseCase 创建通道管理业务逻辑实例
func NewChannelUseCase(
	channelRepo repo.ChannelRepo,
	botRepo repo.BotRepo,
	logger *zerolog.Logger,
) ChannelUseCase {
	return &channelUseCase{
		channelRepo: channelRepo,
		botRepo:     botRepo,
		logger:      logger,
	}
}

// CreateChannel 创建通道
func (uc *channelUseCase) CreateChannel(ctx context.Context, req CreateChannelRequest) (*entity.Channel, error) {
	log := uc.logger.With().
		Str("method", "CreateChannel").
		Int64("bot_id", req.BotID).
		Str("platform_type", req.PlatformType).
		Str("channel_name", req.ChannelName).
		Logger()

	log.Info().Msg("开始创建通道")

	// 验证机器人是否存在且活跃
	bot, err := uc.botRepo.GetByID(ctx, req.BotID)
	if err != nil {
		log.Error().Err(err).Msg("获取机器人信息失败")
		return nil, fmt.Errorf("获取机器人信息失败: %w", err)
	}

	if !bot.IsActive() {
		log.Warn().Msg("机器人未激活，无法创建通道")
		return nil, fmt.Errorf("机器人未激活")
	}

	// 检查通道名称是否重复
	existingChannel, err := uc.channelRepo.GetByName(ctx, req.BotID, req.ChannelName)
	if err == nil && existingChannel != nil {
		log.Warn().Msg("通道名称已存在")
		return nil, fmt.Errorf("通道名称已存在")
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
		log.Error().Err(err).Msg("通道数据验证失败")
		return nil, fmt.Errorf("通道数据验证失败: %w", err)
	}

	// 保存通道到数据库
	if err := uc.channelRepo.Create(ctx, channel); err != nil {
		log.Error().Err(err).Msg("保存通道失败")
		return nil, fmt.Errorf("保存通道失败: %w", err)
	}

	log.Info().
		Int64("channel_id", channel.ID).
		Str("webhook_path", channel.WebhookPath).
		Msg("通道创建成功")

	return channel, nil
}

// UpdateChannel 更新通道
func (uc *channelUseCase) UpdateChannel(ctx context.Context, req UpdateChannelRequest) (*entity.Channel, error) {
	log := uc.logger.With().
		Str("method", "UpdateChannel").
		Int64("channel_id", req.ID).
		Logger()

	log.Info().Msg("开始更新通道")

	// 获取现有通道
	channel, err := uc.channelRepo.GetByID(ctx, req.ID)
	if err != nil {
		log.Error().Err(err).Msg("获取通道信息失败")
		return nil, fmt.Errorf("获取通道信息失败: %w", err)
	}

	// 更新通道名称
	if req.ChannelName != nil {
		// 检查新名称是否重复
		if *req.ChannelName != channel.ChannelName {
			existingChannel, err := uc.channelRepo.GetByName(ctx, channel.BotID, *req.ChannelName)
			if err == nil && existingChannel != nil {
				log.Warn().Msg("通道名称已存在")
				return nil, fmt.Errorf("通道名称已存在")
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
		log.Error().Err(err).Msg("通道数据验证失败")
		return nil, fmt.Errorf("通道数据验证失败: %w", err)
	}

	// 保存更新
	if err := uc.channelRepo.Update(ctx, channel); err != nil {
		log.Error().Err(err).Msg("更新通道失败")
		return nil, fmt.Errorf("更新通道失败: %w", err)
	}

	log.Info().Msg("通道更新成功")
	return channel, nil
}

// DeleteChannel 删除通道
func (uc *channelUseCase) DeleteChannel(ctx context.Context, id int64) error {
	log := uc.logger.With().
		Str("method", "DeleteChannel").
		Int64("channel_id", id).
		Logger()

	log.Info().Msg("开始删除通道")

	// 获取通道信息
	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("获取通道信息失败")
		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	// 检查通道是否有未处理的消息
	pendingCount, err := uc.channelRepo.GetPendingMessageCount(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("检查待处理消息失败")
		return fmt.Errorf("检查待处理消息失败: %w", err)
	}

	if pendingCount > 0 {
		log.Warn().Int64("pending_count", pendingCount).Msg("通道存在待处理消息，无法删除")
		return fmt.Errorf("通道存在 %d 条待处理消息，无法删除", pendingCount)
	}

	// 软删除通道（标记为已删除状态）
	channel.Status = entity.StatusDeleted
	channel.UpdatedAt = time.Now()

	if err := uc.channelRepo.Update(ctx, channel); err != nil {
		log.Error().Err(err).Msg("删除通道失败")
		return fmt.Errorf("删除通道失败: %w", err)
	}

	log.Info().Msg("通道删除成功")
	return nil
}

// GetChannel 获取通道信息
func (uc *channelUseCase) GetChannel(ctx context.Context, id int64) (*entity.Channel, error) {
	log := uc.logger.With().
		Str("method", "GetChannel").
		Int64("channel_id", id).
		Logger()

	log.Info().Msg("获取通道信息")

	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("获取通道信息失败")
		return nil, fmt.Errorf("获取通道信息失败: %w", err)
	}

	log.Info().
		Str("channel_name", channel.ChannelName).
		Str("platform_type", channel.PlatformType).
		Str("status", channel.Status.String()).
		Str("connection_status", channel.ConnectionStatus.String()).
		Msg("通道信息获取成功")

	return channel, nil
}

// ListChannels 获取通道列表
func (uc *channelUseCase) ListChannels(ctx context.Context, params ListChannelsParams) (*ChannelListResult, error) {
	log := uc.logger.With().
		Str("method", "ListChannels").
		Logger()

	log.Info().Msg("获取通道列表")

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

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 查询通道列表
	channels, total, err := uc.channelRepo.ListWithPagination(ctx, filters, offset, params.PageSize)
	if err != nil {
		log.Error().Err(err).Msg("查询通道列表失败")
		return nil, fmt.Errorf("查询通道列表失败: %w", err)
	}

	// 计算总页数
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	result := &ChannelListResult{
		Items:      channels,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}

	log.Info().
		Int64("total", total).
		Int("page", params.Page).
		Int("page_size", params.PageSize).
		Msg("通道列表查询完成")

	return result, nil
}

// UpdateChannelStatus 更新通道状态
func (uc *channelUseCase) UpdateChannelStatus(ctx context.Context, id int64, status entity.ConnectionStatus) error {
	log := uc.logger.With().
		Str("method", "UpdateChannelStatus").
		Int64("channel_id", id).
		Str("status", status.String()).
		Logger()

	log.Info().Msg("更新通道连接状态")

	// 获取通道信息
	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("获取通道信息失败")
		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	// 更新连接状态
	channel.UpdateConnectionStatus(status)
	channel.UpdatedAt = time.Now()

	// 保存更新
	if err := uc.channelRepo.Update(ctx, channel); err != nil {
		log.Error().Err(err).Msg("更新通道状态失败")
		return fmt.Errorf("更新通道状态失败: %w", err)
	}

	log.Info().Msg("通道连接状态更新成功")
	return nil
}

// RefreshChannelToken 刷新通道令牌
func (uc *channelUseCase) RefreshChannelToken(ctx context.Context, id int64) error {
	log := uc.logger.With().
		Str("method", "RefreshChannelToken").
		Int64("channel_id", id).
		Logger()

	log.Info().Msg("刷新通道令牌")

	// 获取通道信息
	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("获取通道信息失败")
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
		log.Error().Err(err).Msg("保存令牌更新失败")
		return fmt.Errorf("保存令牌更新失败: %w", err)
	}

	log.Info().
		Str("new_token", "***").
		Time("expires_at", expiresAt).
		Msg("通道令牌刷新成功")

	return nil
}
