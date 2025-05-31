package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/repo"
)

// messageUseCase 消息处理业务逻辑实现
type messageUseCase struct {
	messageRepo repo.MessageRepo
	channelRepo repo.ChannelRepo
	logger      *zerolog.Logger
}

// NewMessageUseCase 创建消息处理业务逻辑实例
func NewMessageUseCase(
	messageRepo repo.MessageRepo,
	channelRepo repo.ChannelRepo,
	logger *zerolog.Logger,
) MessageUseCase {
	return &messageUseCase{
		messageRepo: messageRepo,
		channelRepo: channelRepo,
		logger:      logger,
	}
}

// ProcessInboundMessage 处理入站消息
func (uc *messageUseCase) ProcessInboundMessage(ctx context.Context, msg *entity.Message) error {
	log := uc.logger.With().
		Str("method", "ProcessInboundMessage").
		Str("message_id", msg.MessageID).
		Int64("channel_id", msg.ChannelID).
		Logger()

	log.Info().Msg("开始处理入站消息")

	// 验证消息数据
	if err := msg.Validate(); err != nil {
		log.Error().Err(err).Msg("消息验证失败")
		return fmt.Errorf("消息验证失败: %w", err)
	}

	// 检查通道是否存在且活跃
	channel, err := uc.channelRepo.GetByID(ctx, msg.ChannelID)
	if err != nil {
		log.Error().Err(err).Msg("获取通道信息失败")
		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	if !channel.IsActive() {
		log.Warn().Msg("通道未激活，拒绝处理消息")
		return fmt.Errorf("通道未激活")
	}

	// 设置消息为入站方向
	msg.Direction = entity.MessageDirectionInbound
	msg.MessageStatus = entity.MessageStatusPending
	msg.ReceivedAt = time.Now()

	// 生成消息ID（如果为空）
	if msg.MessageID == "" {
		msg.MessageID = uc.generateMessageID()
	}

	// 保存消息到数据库
	if err := uc.messageRepo.Create(ctx, msg); err != nil {
		log.Error().Err(err).Msg("保存消息失败")
		return fmt.Errorf("保存消息失败: %w", err)
	}

	// 标记消息为处理中
	msg.MarkAsProcessing()
	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		log.Error().Err(err).Msg("更新消息状态失败")
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	// TODO: 将消息发送到消息队列进行异步处理
	// 这里可以集成 RabbitMQ 或其他消息队列

	log.Info().Msg("入站消息处理完成")
	return nil
}

// SendMessage 发送消息
func (uc *messageUseCase) SendMessage(ctx context.Context, msg *entity.Message) error {
	log := uc.logger.With().
		Str("method", "SendMessage").
		Str("message_id", msg.MessageID).
		Int64("channel_id", msg.ChannelID).
		Logger()

	log.Info().Msg("开始发送消息")

	// 验证消息数据
	if err := msg.Validate(); err != nil {
		log.Error().Err(err).Msg("消息验证失败")
		return fmt.Errorf("消息验证失败: %w", err)
	}

	// 检查通道是否存在且活跃
	channel, err := uc.channelRepo.GetByID(ctx, msg.ChannelID)
	if err != nil {
		log.Error().Err(err).Msg("获取通道信息失败")
		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	if !channel.IsActive() || !channel.IsConnected() {
		log.Warn().Msg("通道未就绪，无法发送消息")
		return fmt.Errorf("通道未就绪")
	}

	// 设置消息为出站方向
	msg.Direction = entity.MessageDirectionOutbound
	msg.MessageStatus = entity.MessageStatusPending

	// 生成消息ID（如果为空）
	if msg.MessageID == "" {
		msg.MessageID = uc.generateMessageID()
	}

	// 保存消息到数据库
	if err := uc.messageRepo.Create(ctx, msg); err != nil {
		log.Error().Err(err).Msg("保存消息失败")
		return fmt.Errorf("保存消息失败: %w", err)
	}

	// 标记消息为处理中
	msg.MarkAsProcessing()
	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		log.Error().Err(err).Msg("更新消息状态失败")
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	// TODO: 调用平台适配器发送消息
	// 这里应该根据 channel.PlatformType 选择对应的适配器

	// 模拟发送成功
	msg.MarkAsSent()
	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		log.Error().Err(err).Msg("更新消息发送状态失败")
		return fmt.Errorf("更新消息发送状态失败: %w", err)
	}

	log.Info().Msg("消息发送完成")
	return nil
}

// GetMessageHistory 获取消息历史
func (uc *messageUseCase) GetMessageHistory(ctx context.Context, params GetMessageHistoryParams) (*MessageHistoryResult, error) {
	log := uc.logger.With().
		Str("method", "GetMessageHistory").
		Logger()

	log.Info().Msg("获取消息历史")

	// 构建查询过滤器
	filters := make(map[string]interface{})

	if params.ChannelID != nil {
		filters["channel_id"] = *params.ChannelID
	}
	if params.SenderID != nil {
		filters["sender_id"] = *params.SenderID
	}
	if params.ReceiverID != nil {
		filters["receiver_id"] = *params.ReceiverID
	}
	if params.MessageType != nil {
		filters["message_type"] = *params.MessageType
	}
	if params.MessageStatus != nil {
		filters["message_status"] = *params.MessageStatus
	}
	if params.Direction != nil {
		filters["direction"] = *params.Direction
	}
	if params.StartTime != nil {
		filters["start_time"] = *params.StartTime
	}
	if params.EndTime != nil {
		filters["end_time"] = *params.EndTime
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 查询消息列表
	messages, total, err := uc.messageRepo.ListWithPagination(ctx, filters, offset, params.PageSize)
	if err != nil {
		log.Error().Err(err).Msg("查询消息历史失败")
		return nil, fmt.Errorf("查询消息历史失败: %w", err)
	}

	// 计算总页数
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	result := &MessageHistoryResult{
		Items:      messages,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}

	log.Info().
		Int64("total", total).
		Int("page", params.Page).
		Int("page_size", params.PageSize).
		Msg("消息历史查询完成")

	return result, nil
}

// GetMessage 根据ID获取消息
func (uc *messageUseCase) GetMessage(ctx context.Context, id int64) (*entity.Message, error) {
	log := uc.logger.With().
		Str("method", "GetMessage").
		Int64("id", id).
		Logger()

	log.Info().Msg("获取消息详情")

	message, err := uc.messageRepo.GetByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("获取消息失败")
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	log.Info().
		Str("message_id", message.MessageID).
		Str("status", message.MessageStatus.String()).
		Msg("消息详情获取完成")

	return message, nil
}

// RetryFailedMessage 重试失败的消息
func (uc *messageUseCase) RetryFailedMessage(ctx context.Context, messageID int64) error {
	log := uc.logger.With().
		Str("method", "RetryFailedMessage").
		Int64("message_id", messageID).
		Logger()

	log.Info().Msg("重试失败消息")

	// 获取消息
	message, err := uc.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		log.Error().Err(err).Msg("获取消息失败")
		return fmt.Errorf("获取消息失败: %w", err)
	}

	// 检查消息状态是否可以重试
	if !message.CanRetry() {
		log.Warn().
			Str("status", message.MessageStatus.String()).
			Msg("消息状态不支持重试")
		return fmt.Errorf("消息状态不支持重试")
	}

	// 重置消息状态
	message.MessageStatus = entity.MessageStatusPending
	message.ErrorMessage = ""

	// 更新消息状态
	if err := uc.messageRepo.Update(ctx, message); err != nil {
		log.Error().Err(err).Msg("更新消息状态失败")
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	// 根据消息方向选择处理方式
	if message.IsOutbound() {
		// 出站消息重新发送
		return uc.SendMessage(ctx, message)
	} else {
		// 入站消息重新处理
		return uc.ProcessInboundMessage(ctx, message)
	}
}

// generateMessageID 生成消息ID
func (uc *messageUseCase) generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
