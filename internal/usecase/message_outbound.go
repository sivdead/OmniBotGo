package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

// Outbound send / reply / retry / query / troubleshoot (T03/T11)

// SendMessage 发送消息
func (uc *messageUseCase) SendMessage(ctx context.Context, msg *entity.Message) error {
	uc.logger.Info("开始发送消息", "method", "SendMessage", "message_id", msg.MessageID, "channel_id", msg.ChannelID)

	// B01: generate MessageID before ValidateMessage so HTTP callers need not supply one.
	if msg.MessageID == "" {
		msg.MessageID = uc.generateMessageID()
	}
	// E01: strict MySQL rejects zero datetime for platform_timestamp.
	if msg.PlatformTimestamp.IsZero() {
		msg.PlatformTimestamp = time.Now()
	}

	if err := uc.msgService.ValidateMessage(msg); err != nil {
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
		uc.logger.Warn("通道未激活，无法发送消息")

		return errChannelNotReady
	}
	// Outbound HTTP send (e.g. Telegram sendMessage) does not require a live stream.
	// connection_status used to lag behind long-poll Start → false 「通道未就绪」.
	// Warn only; ConnectionManager now also persists connected on Start.
	if !channel.IsConnected() {
		uc.logger.Warn("通道 connection_status 非已连接，仍尝试发送", "channel_id", channel.ID, "connection_status", channel.ConnectionStatus.String())
	}

	// 设置消息为出站方向
	msg.Direction = entity.MessageDirectionOutbound
	msg.MessageStatus = entity.MessageStatusPending

	// 保存消息到数据库
	if err := uc.messageRepo.Create(ctx, msg); err != nil {
		uc.logger.Error("保存消息失败", "error", err)
		return fmt.Errorf("保存消息失败: %w", err)
	}

	// 使用service标记消息为处理中
	uc.msgService.MarkAsProcessing(msg)
	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		uc.logger.Error("更新消息状态失败", "error", err)
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	// 调用平台适配器发送消息
	if err := uc.sendMessageToPlatform(ctx, msg, channel); err != nil {
		// 发送失败，使用service标记消息为失败状态
		uc.msgService.MarkAsFailed(msg, err.Error())
		if updateErr := uc.messageRepo.Update(ctx, msg); updateErr != nil {
			uc.logger.Error("更新消息失败状态失败", "error", updateErr)
		}
		uc.logger.Error("发送消息到平台失败", "error", err)
		return fmt.Errorf("发送消息失败: %w", err)
	}

	// 发送成功，使用service标记消息为已发送
	uc.msgService.MarkAsSent(msg)
	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		uc.logger.Error("更新消息发送状态失败", "error", err)
		return fmt.Errorf("更新消息发送状态失败: %w", err)
	}

	uc.logger.Info("消息发送完成")
	return nil
}

// ReplyToOriginal resolves an inbound message by business message_id (fallback DB PK)
// and sends a reply into the correct DM or group conversation (T04).
func (uc *messageUseCase) ReplyToOriginal(ctx context.Context, inboundRef string, req ReplyToOriginalRequest) (*entity.Message, error) {
	inbound, err := uc.resolveInboundMessage(ctx, inboundRef)
	if err != nil {
		return nil, err
	}

	if inbound.Direction != entity.MessageDirectionInbound {
		return nil, ErrReplyRequiresInbound
	}

	reply := BuildOutboundReplyFromInbound(inbound, req.Content, req.MessageType)
	if err := uc.SendMessage(ctx, reply); err != nil {
		return nil, fmt.Errorf("reply to original: %w", err)
	}

	return reply, nil
}

func (uc *messageUseCase) resolveInboundMessage(ctx context.Context, ref string) (*entity.Message, error) {
	if ref == "" {
		return nil, ErrInboundMessageNotFound
	}

	msg, err := uc.messageRepo.GetByMessageID(ctx, ref)
	if err == nil {
		return msg, nil
	}

	msg, err = uc.messageRepo.GetByID(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInboundMessageNotFound, ref)
	}

	return msg, nil
}

// sendMessageToPlatform 发送消息到具体平台
func (uc *messageUseCase) sendMessageToPlatform(ctx context.Context, msg *entity.Message, channel *entity.Channel) error {
	platformType := entity.PlatformType(channel.PlatformType)

	// T08: reject undeclared message types before calling the sender
	if platformAdapter, err := uc.adapterManager.GetAdapter(platformType); err == nil {
		if caps, ok := platformAdapter.(port.MessageTypeSupport); ok {
			supported := false
			for _, t := range caps.SupportedMessageTypes() {
				if t == msg.MessageType {
					supported = true
					break
				}
			}
			if !supported {
				return fmt.Errorf("%w: platform %s cannot send %q", port.ErrUnsupportedMessageType, platformType, msg.MessageType)
			}
		}
	}

	// 使用新的MessageSender接口
	messageSender, err := uc.adapterManager.GetMessageSender(platformType)
	if err != nil {
		return fmt.Errorf("获取消息发送器失败: %w", err)
	}

	// 检查访问令牌是否存在和有效（对于需要Token的平台）
	if channel.AccessToken == "" {
		// 对于Stream模式（如钉钉Stream），可能不需要传统的AccessToken
		uc.logger.Debug("通道访问令牌为空，可能为Stream模式")
	}

	// 检查访问令牌是否过期
	if channel.AccessTokenExpiresAt != nil && time.Now().After(*channel.AccessTokenExpiresAt) {
		// 尝试刷新令牌
		if err := uc.refreshChannelToken(ctx, channel); err != nil {
			uc.logger.Warn("刷新访问令牌失败", "error", err)
			// 对于Stream模式，Token刷新失败不是致命错误
		}
	}

	// 转换为统一消息格式
	unifiedMessage := &dto.UnifiedMessage{
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
	return messageSender.SendMessage(ctx, unifiedMessage, channel.Config, channel.AccessToken)
}

// refreshChannelToken 刷新通道访问令牌
func (uc *messageUseCase) refreshChannelToken(ctx context.Context, channel *entity.Channel) error {
	uc.logger.Info("刷新通道访问令牌", "channel_id", channel.ID)

	platformType := entity.PlatformType(channel.PlatformType)

	// 使用新的TokenManager接口
	tokenManager, err := uc.adapterManager.GetTokenManager(platformType)
	if err != nil {
		return fmt.Errorf("获取Token管理器失败: %w", err)
	}

	tokenResponse, err := tokenManager.RefreshAccessToken(ctx, channel.Config, channel.AccessToken)
	if err != nil {
		return fmt.Errorf("刷新访问令牌失败: %w", err)
	}

	// 更新通道的访问令牌
	channel.AccessToken = tokenResponse.AccessToken
	channel.AccessTokenExpiresAt = tokenResponse.ExpiresAt

	if err := uc.channelRepo.Update(ctx, channel); err != nil {
		return fmt.Errorf("更新通道访问令牌失败: %w", err)
	}

	uc.logger.Info("通道访问令牌刷新成功")
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
	listParams := port.ListParams{
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

// GetMessage 根据业务 message_id 或 DB PK 获取消息
func (uc *messageUseCase) GetMessage(ctx context.Context, id string) (*entity.Message, error) {
	uc.logger.Info("获取消息详情", "method", "GetMessage", "message_id", id)

	message, err := uc.resolveInboundMessage(ctx, id)
	if err != nil {
		uc.logger.Error("获取消息失败", "error", err)
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	uc.logger.Info("消息获取成功", "message_id", message.MessageID, "message_type", message.MessageType, "message_status", message.MessageStatus)

	return message, nil
}

// TroubleshootMessage returns receive/process/deliver stage + recoverability for a single message (T11).
func (uc *messageUseCase) TroubleshootMessage(ctx context.Context, id string) (*MessageTroubleshootResult, error) {
	msg, err := uc.resolveInboundMessage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	stage := deriveMessagePipelineStage(msg)
	canRetry := msg.MessageStatus.CanRetry()
	recoverable := canRetry

	action := "none"
	switch {
	case !canRetry && msg.MessageStatus == entity.MessageStatusExpired:
		action = "expired_terminal"
	case canRetry && msg.Direction == entity.MessageDirectionInbound:
		action = "retry_inbound_process"
	case canRetry && msg.Direction == entity.MessageDirectionOutbound:
		action = "retry_outbound_deliver"
	case msg.MessageStatus == entity.MessageStatusPending:
		action = "await_or_inspect_receive"
	case msg.MessageStatus == entity.MessageStatusProcessing:
		action = "await_process"
	case msg.MessageStatus.IsProcessed() || msg.MessageStatus == entity.MessageStatusSent:
		action = "healthy"
	}

	result := &MessageTroubleshootResult{
		ID:              msg.ID,
		MessageID:       msg.MessageID,
		ChannelID:       msg.ChannelID,
		Direction:       msg.Direction.String(),
		Status:          msg.MessageStatus.String(),
		Stage:           stage,
		Recoverable:     recoverable,
		CanRetry:        canRetry,
		RetryCount:      msg.RetryCount,
		ErrorMessage:    msg.ErrorMessage,
		ConversationID:  msg.ConversationID,
		SuggestedAction: action,
	}

	if !msg.ReceivedAt.IsZero() {
		s := msg.ReceivedAt.UTC().Format(time.RFC3339)
		result.ReceivedAt = &s
	}
	if msg.ProcessedAt != nil {
		s := msg.ProcessedAt.UTC().Format(time.RFC3339)
		result.ProcessedAt = &s
	}
	if msg.SentAt != nil {
		s := msg.SentAt.UTC().Format(time.RFC3339)
		result.SentAt = &s
	}

	return result, nil
}

// deriveMessagePipelineStage maps status/direction/timestamps to receive|process|deliver (T10/T11).
func deriveMessagePipelineStage(msg *entity.Message) MessagePipelineStage {
	if msg == nil {
		return MessageStageUnknown
	}

	switch msg.MessageStatus {
	case entity.MessageStatusPending:
		return MessageStageReceive
	case entity.MessageStatusProcessing, entity.MessageStatusProcessed:
		return MessageStageProcess
	case entity.MessageStatusSent:
		return MessageStageDeliver
	case entity.MessageStatusFailed, entity.MessageStatusExpired:
		if msg.Direction == entity.MessageDirectionOutbound {
			return MessageStageDeliver
		}
		if msg.ProcessedAt != nil || msg.MessageStatus == entity.MessageStatusExpired {
			return MessageStageProcess
		}
		// Inbound failed before/during process
		return MessageStageProcess
	default:
		return MessageStageUnknown
	}
}

// RetryFailedMessage 重试失败的消息（更新已有行并重新投递，禁止再 Create）.
func (uc *messageUseCase) RetryFailedMessage(ctx context.Context, messageID string) error {
	uc.logger.Info("重试失败消息", "method", "RetryFailedMessage", "message_id", messageID)

	message, err := uc.resolveInboundMessage(ctx, messageID)
	if err != nil {
		uc.logger.Error("获取消息失败", "error", err)

		return fmt.Errorf("获取消息失败: %w", err)
	}

	if message.MessageStatus != entity.MessageStatusFailed {
		uc.logger.Warn("消息状态不是失败状态，无法重试", "current_status", message.MessageStatus)

		return fmt.Errorf("消息状态不是失败状态")
	}

	channel, err := uc.channelRepo.GetByID(ctx, message.ChannelID)
	if err != nil {
		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	if !channel.IsActive() {
		return errChannelNotReady
	}

	if message.PlatformTimestamp.IsZero() {
		message.PlatformTimestamp = time.Now()
	}

	message.ErrorMessage = ""
	uc.msgService.MarkAsProcessing(message)

	if err := uc.messageRepo.Update(ctx, message); err != nil {
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	switch message.Direction {
	case entity.MessageDirectionInbound:
		if err := uc.routeAndProcessMessage(ctx, message); err != nil {
			uc.handleMessageProcessingError(ctx, message, err)

			return err
		}

		return nil
	case entity.MessageDirectionOutbound:
		if err := uc.sendMessageToPlatform(ctx, message, channel); err != nil {
			uc.msgService.MarkAsFailed(message, err.Error())

			if updateErr := uc.messageRepo.Update(ctx, message); updateErr != nil {
				uc.logger.Error("更新消息失败状态失败", "error", updateErr)
			}

			return fmt.Errorf("发送消息失败: %w", err)
		}

		uc.msgService.MarkAsSent(message)

		if err := uc.messageRepo.Update(ctx, message); err != nil {
			return fmt.Errorf("更新消息发送状态失败: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("%w: %v", errUnsupportedMsgDirection, message.Direction)
	}
}
