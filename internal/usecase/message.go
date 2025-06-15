package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/sivdead/OmniBotGo/internal/adapter"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/repo"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// messageUseCase 消息处理业务逻辑实现
type messageUseCase struct {
	messageRepo    repo.MessageRepo
	channelRepo    repo.ChannelRepo
	adapterManager *adapter.Manager
	logger         logger.Interface
}

// NewMessageUseCase 创建消息处理业务逻辑实例
func NewMessageUseCase(
	messageRepo repo.MessageRepo,
	channelRepo repo.ChannelRepo,
	adapterManager *adapter.Manager,
	logger logger.Interface,
) MessageUseCase {
	return &messageUseCase{
		messageRepo:    messageRepo,
		channelRepo:    channelRepo,
		adapterManager: adapterManager,
		logger:         logger,
	}
}

// ProcessInboundMessage 处理入站消息
func (uc *messageUseCase) ProcessInboundMessage(ctx context.Context, msg *entity.Message) error {
	uc.logger.Info("开始处理入站消息", "method", "ProcessInboundMessage", "message_id", msg.MessageID, "channel_id", msg.ChannelID)

	// 验证消息数据
	if err := msg.Validate(); err != nil {
		uc.logger.Error("消息验证失败", "error", err)
		return fmt.Errorf("消息验证失败: %w", err)
	}

	// 检查通道是否存在且活跃
	channel, err := uc.channelRepo.GetByID(ctx, msg.ChannelID)
	if err != nil {
		uc.logger.Error("获取通道信息失败", "error", err)
		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	if !channel.IsActive() {
		uc.logger.Warn("通道未激活，拒绝处理消息")
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
		uc.logger.Error("保存消息失败", "error", err)
		return fmt.Errorf("保存消息失败: %w", err)
	}

	// 标记消息为处理中
	msg.MarkAsProcessing()
	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		uc.logger.Error("更新消息状态失败", "error", err)
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	// TODO: 将消息发送到消息队列进行异步处理
	// 这里可以集成 RabbitMQ 或其他消息队列

	uc.logger.Info("入站消息处理完成")
	return nil
}

// SendMessage 发送消息
func (uc *messageUseCase) SendMessage(ctx context.Context, msg *entity.Message) error {
	uc.logger.Info("开始发送消息", "method", "SendMessage", "message_id", msg.MessageID, "channel_id", msg.ChannelID)

	// 验证消息数据
	if err := msg.Validate(); err != nil {
		uc.logger.Error("消息验证失败", "error", err)
		return fmt.Errorf("消息验证失败: %w", err)
	}

	// 检查通道是否存在且活跃
	channel, err := uc.channelRepo.GetByID(ctx, msg.ChannelID)
	if err != nil {
		uc.logger.Error("获取通道信息失败", "error", err)
		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	if !channel.IsActive() || !channel.IsConnected() {
		uc.logger.Warn("通道未就绪，无法发送消息")
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
		uc.logger.Error("保存消息失败", "error", err)
		return fmt.Errorf("保存消息失败: %w", err)
	}

	// 标记消息为处理中
	msg.MarkAsProcessing()
	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		uc.logger.Error("更新消息状态失败", "error", err)
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	// 调用平台适配器发送消息
	if err := uc.sendMessageToPlatform(ctx, msg, channel); err != nil {
		// 发送失败，标记消息为失败状态
		msg.MarkAsFailed(err.Error())
		if updateErr := uc.messageRepo.Update(ctx, msg); updateErr != nil {
			uc.logger.Error("更新消息失败状态失败", "error", updateErr)
		}
		uc.logger.Error("发送消息到平台失败", "error", err)
		return fmt.Errorf("发送消息失败: %w", err)
	}

	// 发送成功，标记消息为已发送
	msg.MarkAsSent()
	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		uc.logger.Error("更新消息发送状态失败", "error", err)
		return fmt.Errorf("更新消息发送状态失败: %w", err)
	}

	uc.logger.Info("消息发送完成")
	return nil
}

// sendMessageToPlatform 发送消息到具体平台
func (uc *messageUseCase) sendMessageToPlatform(ctx context.Context, msg *entity.Message, channel *entity.Channel) error {
	// 检查访问令牌是否存在和有效
	if channel.AccessToken == "" {
		return fmt.Errorf("通道访问令牌为空")
	}

	// 检查访问令牌是否过期
	if channel.AccessTokenExpiresAt != nil && time.Now().After(*channel.AccessTokenExpiresAt) {
		// 尝试刷新令牌
		if err := uc.refreshChannelToken(ctx, channel); err != nil {
			return fmt.Errorf("刷新访问令牌失败: %w", err)
		}
	}

	// 转换为统一消息格式
	unifiedMessage := &entity.UnifiedMessage{
		MessageID:         msg.MessageID,
		MessageType:       msg.MessageType,
		SenderID:          msg.SenderID,
		SenderName:        msg.SenderName,
		SenderType:        msg.SenderType,
		ReceiverID:        msg.ReceiverID,
		ReceiverName:      msg.ReceiverName,
		ReceiverType:      msg.ReceiverType,
		Content:           msg.Content,
		RawContent:        msg.RawContent,
		MediaURL:          msg.MediaURL,
		MediaType:         msg.MediaType,
		FileSize:          msg.FileSize,
		ConversationID:    msg.ConversationID,
		PlatformMessageID: msg.PlatformMessageID,
		PlatformTimestamp: msg.PlatformTimestamp,
	}

	// 发送消息到平台
	platformType := entity.PlatformType(channel.PlatformType)
	return uc.adapterManager.SendMessage(ctx, platformType, unifiedMessage, channel.Config, channel.AccessToken)
}

// refreshChannelToken 刷新通道访问令牌
func (uc *messageUseCase) refreshChannelToken(ctx context.Context, channel *entity.Channel) error {
	uc.logger.Info("刷新通道访问令牌", "channel_id", channel.ID)

	platformType := entity.PlatformType(channel.PlatformType)
	tokenResponse, err := uc.adapterManager.RefreshAccessToken(ctx, platformType, channel.Config, channel.AccessToken)
	if err != nil {
		return fmt.Errorf("刷新访问令牌失败: %w", err)
	}

	// 更新通道访问令牌
	if err := uc.channelRepo.UpdateAccessToken(ctx, channel.ID, tokenResponse.AccessToken, tokenResponse.ExpiresAt); err != nil {
		return fmt.Errorf("更新通道访问令牌失败: %w", err)
	}

	// 更新内存中的令牌
	channel.AccessToken = tokenResponse.AccessToken
	channel.AccessTokenExpiresAt = tokenResponse.ExpiresAt

	uc.logger.Info("通道访问令牌刷新成功", "channel_id", channel.ID)
	return nil
}

// GetMessageHistory 获取消息历史
func (uc *messageUseCase) GetMessageHistory(ctx context.Context, params GetMessageHistoryParams) (*MessageHistoryResult, error) {
	uc.logger.Info("获取消息历史", "method", "GetMessageHistory")

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

	// 查询消息列表
	listParams := repo.ListParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		Filters:  filters,
	}

	result, err := uc.messageRepo.List(ctx, listParams)
	if err != nil {
		uc.logger.Error("查询消息历史失败", "error", err)
		return nil, fmt.Errorf("查询消息历史失败: %w", err)
	}

	// 转换指针切片为值切片
	messageValues := make([]entity.Message, len(result.Items))
	for i, msg := range result.Items {
		messageValues[i] = *msg
	}

	historyResult := &MessageHistoryResult{
		Items:      messageValues,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}

	uc.logger.Info("消息历史查询完成", "total", result.Total, "page", result.Page, "page_size", result.PageSize)

	return historyResult, nil
}

// GetMessage 根据ID获取消息
func (uc *messageUseCase) GetMessage(ctx context.Context, id int64) (*entity.Message, error) {
	uc.logger.Info("获取消息详情", "method", "GetMessage", "message_id", id)

	message, err := uc.messageRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取消息失败", "error", err)
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	uc.logger.Info("消息获取成功", "message_id", message.MessageID, "message_type", message.MessageType, "message_status", message.MessageStatus)

	return message, nil
}

// RetryFailedMessage 重试失败的消息
func (uc *messageUseCase) RetryFailedMessage(ctx context.Context, messageID int64) error {
	uc.logger.Info("重试失败消息", "method", "RetryFailedMessage", "message_id", messageID)

	// 获取消息
	message, err := uc.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		uc.logger.Error("获取消息失败", "error", err)
		return fmt.Errorf("获取消息失败: %w", err)
	}

	// 检查消息状态
	if message.MessageStatus != entity.MessageStatusFailed {
		uc.logger.Warn("消息状态不是失败状态，无法重试", "current_status", message.MessageStatus)
		return fmt.Errorf("消息状态不是失败状态")
	}

	// 重新发送消息
	return uc.SendMessage(ctx, message)
}

// generateMessageID 生成消息ID
func (uc *messageUseCase) generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
