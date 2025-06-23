package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/internal/usecase/service"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// messageUseCase 消息处理业务逻辑实现
type messageUseCase struct {
	messageRepo    port.MessageRepository
	channelRepo    port.ChannelRepository
	adapterManager port.AdapterManager
	routingUC      RoutingUseCase
	msgService     *service.MessageService
	queueRepo      port.MessageQueueRepository
	logger         logger.Interface
}

// NewMessageUseCase 创建消息处理业务逻辑实例
func NewMessageUseCase(
	messageRepo port.MessageRepository,
	channelRepo port.ChannelRepository,
	adapterManager port.AdapterManager,
	routingUC RoutingUseCase,
	queueRepo port.MessageQueueRepository,
	logger logger.Interface,
) MessageUseCase {
	return &messageUseCase{
		messageRepo:    messageRepo,
		channelRepo:    channelRepo,
		adapterManager: adapterManager,
		routingUC:      routingUC,
		msgService:     service.NewMessageService(),
		queueRepo:      queueRepo,
		logger:         logger,
	}
}

// CreateStreamMessageHandler 创建用于Stream适配器的消息处理回调函数
func (uc *messageUseCase) CreateStreamMessageHandler() port.MessageHandler {
	return func(ctx context.Context, unifiedMessage *dto.UnifiedMessage) error {
		uc.logger.Info("Stream消息处理回调", "message_id", unifiedMessage.MessageID, "sender_id", unifiedMessage.SenderID)

		// 创建消息实体
		message := &entity.Message{
			MessageID:         unifiedMessage.MessageID,
			PlatformMessageID: unifiedMessage.PlatformMessageID,
			Direction:         entity.MessageDirectionInbound,
			MessageType:       unifiedMessage.MessageType,
			ContentType:       unifiedMessage.MediaType,
			SenderID:          unifiedMessage.SenderID,
			SenderName:        unifiedMessage.SenderName,
			SenderType:        unifiedMessage.SenderType,
			ReceiverID:        unifiedMessage.ReceiverID,
			ReceiverName:      unifiedMessage.ReceiverName,
			ReceiverType:      unifiedMessage.ReceiverType,
			Content:           unifiedMessage.Content,
			RawContent:        entity.JSONField(unifiedMessage.RawContent),
			MediaURL:          unifiedMessage.MediaURL,
			MediaType:         unifiedMessage.MediaType,
			FileSize:          unifiedMessage.FileSize,
			MessageStatus:     entity.MessageStatusPending,
			ConversationID:    unifiedMessage.ConversationID,
			PlatformTimestamp: unifiedMessage.PlatformTimestamp,
			ReceivedAt:        time.Now(),
		}

		// 这里需要找到对应的ChannelID，可以通过平台类型和配置信息查找
		// 为了简化，我们可以在RawContent中传递channelID
		if channelIDValue, ok := unifiedMessage.RawContent["channel_id"]; ok {
			if channelID, ok := channelIDValue.(float64); ok {
				message.ChannelID = int64(channelID)
			}
		}

		// 处理入站消息
		return uc.ProcessInboundMessage(ctx, message)
	}
}

// ProcessInboundMessage 处理入站消息
func (uc *messageUseCase) ProcessInboundMessage(ctx context.Context, msg *entity.Message) error {
	uc.logger.Info("开始处理入站消息", "method", "ProcessInboundMessage", "message_id", msg.MessageID, "channel_id", msg.ChannelID)

	if err := uc.validateInboundMessage(ctx, msg); err != nil {
		return err
	}

	if err := uc.checkChannelStatus(ctx, msg.ChannelID); err != nil {
		return err
	}

	if err := uc.prepareInboundMessage(msg); err != nil {
		return err
	}

	if skip, err := uc.checkDuplicateMessage(ctx, msg); err != nil {
		return err
	} else if skip {
		return nil // 重复消息，静默跳过
	}

	if err := uc.persistAndProcessMessage(ctx, msg); err != nil {
		return err
	}

	uc.logger.Info("入站消息处理完成")
	return nil
}

// validateInboundMessage 验证入站消息
func (uc *messageUseCase) validateInboundMessage(ctx context.Context, msg *entity.Message) error {
	if err := uc.msgService.ValidateMessage(msg); err != nil {
		uc.logger.Error("消息验证失败", "error", err)
		return fmt.Errorf("消息验证失败: %w", err)
	}
	return nil
}

// checkChannelStatus 检查通道状态
func (uc *messageUseCase) checkChannelStatus(ctx context.Context, channelID int64) error {
	channel, err := uc.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		uc.logger.Error("获取通道信息失败", "error", err)
		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	if !channel.IsActive() {
		uc.logger.Warn("通道未激活，拒绝处理消息")
		return fmt.Errorf("通道未激活")
	}
	return nil
}

// prepareInboundMessage 准备入站消息
func (uc *messageUseCase) prepareInboundMessage(msg *entity.Message) error {
	msg.Direction = entity.MessageDirectionInbound
	msg.MessageStatus = entity.MessageStatusPending
	msg.ReceivedAt = time.Now()

	if msg.MessageID == "" {
		msg.MessageID = uc.generateMessageID()
	}
	return nil
}

// checkDuplicateMessage 检查重复消息
func (uc *messageUseCase) checkDuplicateMessage(ctx context.Context, msg *entity.Message) (bool, error) {
	if msg.PlatformMessageID == "" {
		return false, nil
	}

	existingMsg, err := uc.messageRepo.GetByPlatformMessageID(ctx, msg.ChannelID, msg.PlatformMessageID)
	if err == nil && existingMsg != nil {
		uc.logger.Warn("检测到重复消息，跳过处理",
			"message_id", msg.MessageID,
			"platform_message_id", msg.PlatformMessageID,
			"existing_message_id", existingMsg.MessageID)
		return true, nil // 返回true表示跳过处理
	}
	return false, nil
}

// persistAndProcessMessage 持久化并处理消息
func (uc *messageUseCase) persistAndProcessMessage(ctx context.Context, msg *entity.Message) error {
	// 保存消息到数据库
	if err := uc.messageRepo.Create(ctx, msg); err != nil {
		uc.logger.Error("保存消息失败", "error", err)
		return fmt.Errorf("保存消息失败: %w", err)
	}

	// 标记消息为处理中
	uc.msgService.MarkAsProcessing(msg)
	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		uc.logger.Error("更新消息状态失败", "error", err)
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	// 进行消息路由处理
	if err := uc.routeAndProcessMessage(ctx, msg); err != nil {
		uc.logger.Error("消息路由处理失败", "error", err)
		uc.handleMessageProcessingError(ctx, msg, err)
	}

	return nil
}

// handleMessageProcessingError 处理消息处理错误
func (uc *messageUseCase) handleMessageProcessingError(ctx context.Context, msg *entity.Message, err error) {
	uc.msgService.MarkAsFailed(msg, err.Error())
	if updateErr := uc.messageRepo.Update(ctx, msg); updateErr != nil {
		uc.logger.Error("更新消息状态失败", "error", updateErr)
	}

	// 将失败的消息放入重试队列
	if enqueueErr := uc.enqueueFailedMessage(ctx, msg); enqueueErr != nil {
		uc.logger.Error("将失败消息加入队列失败", "error", enqueueErr)
	}
}

// SendMessage 发送消息
func (uc *messageUseCase) SendMessage(ctx context.Context, msg *entity.Message) error {
	uc.logger.Info("开始发送消息", "method", "SendMessage", "message_id", msg.MessageID, "channel_id", msg.ChannelID)

	// 使用service验证消息数据
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

// sendMessageToPlatform 发送消息到具体平台
func (uc *messageUseCase) sendMessageToPlatform(ctx context.Context, msg *entity.Message, channel *entity.Channel) error {
	platformType := entity.PlatformType(channel.PlatformType)

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

// routeAndProcessMessage 路由并处理消息
func (uc *messageUseCase) routeAndProcessMessage(ctx context.Context, msg *entity.Message) error {
	// 使用路由规则处理消息
	processors, err := uc.routingUC.RouteMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("消息路由失败: %w", err)
	}

	if len(processors) == 0 {
		uc.logger.Warn("没有找到匹配的处理器", "message_id", msg.MessageID)
		return nil // 没有处理器不算错误，可能是正常情况
	}

	// 依次调用处理器处理消息
	for _, processor := range processors {
		if err := uc.processMessageWithProcessor(ctx, msg, processor); err != nil {
			uc.logger.Error("处理器处理消息失败", "processor_id", processor.ID, "processor_name", processor.ProcessorName, "error", err)
			// 继续处理其他处理器，不中断流程
		}
	}

	// 使用service标记消息为已处理
	uc.msgService.MarkAsProcessed(msg)
	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	return nil
}

// processMessageWithProcessor 使用特定处理器处理消息
func (uc *messageUseCase) processMessageWithProcessor(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor) error {
	uc.logger.Info("开始处理消息", "processor_id", processor.ID, "processor_type", processor.ProcessorType, "message_id", msg.MessageID)

	switch processor.ProcessorType {
	case "webhook_forwarder":
		return uc.processWebhookForwarder(ctx, msg, processor)
	case "auto_reply":
		return uc.processAutoReply(ctx, msg, processor)
	case "ai_chat":
		return uc.processAIChat(ctx, msg, processor)
	default:
		uc.logger.Warn("未知的处理器类型", "processor_type", processor.ProcessorType)
		return fmt.Errorf("未知的处理器类型: %s", processor.ProcessorType)
	}
}

// processWebhookForwarder 处理Webhook转发
func (uc *messageUseCase) processWebhookForwarder(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor) error {
	// 获取Webhook URL配置
	webhookURL := processor.GetConfigValue("webhook_url")
	if webhookURL == nil {
		return fmt.Errorf("Webhook处理器缺少webhook_url配置")
	}

	url, ok := webhookURL.(string)
	if !ok || url == "" {
		return fmt.Errorf("无效的webhook_url配置")
	}

	// 检查是否需要将消息暂存到队列（例如：后端服务不可用时）
	if shouldQueueMessage(processor) {
		return uc.enqueueWebhookMessage(ctx, msg, processor, url)
	}

	// 实现Webhook转发逻辑
	// 构建要发送的数据
	payload := map[string]interface{}{
		"message_id":   msg.MessageID,
		"channel_id":   msg.ChannelID,
		"sender_id":    msg.SenderID,
		"sender_name":  msg.SenderName,
		"content":      msg.Content,
		"message_type": msg.MessageType,
		"timestamp":    msg.ReceivedAt,
		"raw_content":  msg.RawContent,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("构建Webhook数据失败: %w", err)
	}

	// 发送HTTP POST请求到目标URL
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建Webhook请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "OmniBotGo/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送Webhook请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 处理响应并记录日志
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		uc.logger.Info("Webhook转发成功", "url", url, "message_id", msg.MessageID, "status_code", resp.StatusCode)
	} else {
		body, _ := io.ReadAll(resp.Body)
		uc.logger.Error("Webhook转发失败", "url", url, "message_id", msg.MessageID, "status_code", resp.StatusCode, "response", string(body))
		return fmt.Errorf("Webhook响应错误: %d", resp.StatusCode)
	}

	uc.logger.Info("Webhook转发处理完成", "url", url, "message_id", msg.MessageID)
	return nil
}

// enqueueWebhookMessage 将Webhook消息加入队列
func (uc *messageUseCase) enqueueWebhookMessage(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor, webhookURL string) error {
	queue := &entity.MessageQueue{
		QueueName:   "webhook_forward",
		MessageID:   msg.MessageID,
		Priority:    50, // Webhook转发优先级较高
		MaxRetries:  5,
		Status:      entity.QueueStatusPending,
		ScheduledAt: time.Now(),
		Payload: entity.JSONField{
			"message_id":   msg.ID,
			"processor_id": processor.ID,
			"webhook_url":  webhookURL,
			"message_data": msg,
		},
	}

	if err := uc.queueRepo.Create(ctx, queue); err != nil {
		return fmt.Errorf("将消息加入Webhook队列失败: %w", err)
	}

	uc.logger.Info("消息已加入Webhook转发队列", "message_id", msg.MessageID, "queue_id", queue.ID)
	return nil
}

// shouldQueueMessage 判断是否应该将消息放入队列
func shouldQueueMessage(processor *entity.MessageProcessor) bool {
	// 可以根据处理器配置判断
	// 例如：检查后端服务健康状态、是否启用队列模式等
	queueMode := processor.GetConfigValue("queue_mode")
	if queueMode != nil {
		if enabled, ok := queueMode.(bool); ok && enabled {
			return true
		}
	}
	return false
}

// processAutoReply 处理自动回复
func (uc *messageUseCase) processAutoReply(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor) error {
	// 获取自动回复内容配置
	replyContent := processor.GetConfigValue("reply_content")
	if replyContent == nil {
		return fmt.Errorf("自动回复处理器缺少reply_content配置")
	}

	content, ok := replyContent.(string)
	if !ok || content == "" {
		return fmt.Errorf("无效的reply_content配置")
	}

	// 构建回复消息
	replyMsg := &entity.Message{
		MessageID:       uc.generateMessageID(),
		ChannelID:       msg.ChannelID,
		Direction:       entity.MessageDirectionOutbound,
		MessageType:     "text",
		SenderID:        msg.ReceiverID, // 回复者变成发送者
		SenderName:      msg.ReceiverName,
		SenderType:      msg.ReceiverType,
		ReceiverID:      msg.SenderID, // 原发送者变成接收者
		ReceiverName:    msg.SenderName,
		ReceiverType:    msg.SenderType,
		Content:         content,
		ConversationID:  msg.ConversationID,
		ParentMessageID: &msg.ID, // 设置父消息ID
	}

	// 发送回复消息
	if err := uc.SendMessage(ctx, replyMsg); err != nil {
		return fmt.Errorf("发送自动回复失败: %w", err)
	}

	uc.logger.Info("自动回复处理完成", "reply_message_id", replyMsg.MessageID, "original_message_id", msg.MessageID)
	return nil
}

// processAIChat 处理AI聊天
func (uc *messageUseCase) processAIChat(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor) error {
	// 获取AI配置
	apiURL := processor.GetConfigValue("api_url")
	apiKey := processor.GetConfigValue("api_key")
	model := processor.GetConfigValue("model")
	
	if apiURL == nil || apiKey == nil {
		return fmt.Errorf("AI聊天处理器缺少必要配置：api_url 或 api_key")
	}

	url, ok := apiURL.(string)
	if !ok || url == "" {
		return fmt.Errorf("无效的api_url配置")
	}

	key, ok := apiKey.(string)
	if !ok || key == "" {
		return fmt.Errorf("无效的api_key配置")
	}

	modelName := "gpt-3.5-turbo" // 默认模型
	if model != nil {
		if m, ok := model.(string); ok && m != "" {
			modelName = m
		}
	}

	// 构建AI请求
	aiRequest := map[string]interface{}{
		"model": modelName,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": msg.Content,
			},
		},
		"max_tokens":  1000,
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(aiRequest)
	if err != nil {
		return fmt.Errorf("构建AI请求数据失败: %w", err)
	}

	// 发送请求到AI API
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建AI请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("User-Agent", "OmniBotGo/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送AI请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		uc.logger.Error("AI API请求失败", "status_code", resp.StatusCode, "response", string(body))
		return fmt.Errorf("AI API响应错误: %d", resp.StatusCode)
	}

	// 解析AI响应
	var aiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&aiResponse); err != nil {
		return fmt.Errorf("解析AI响应失败: %w", err)
	}

	if aiResponse.Error.Message != "" {
		return fmt.Errorf("AI API错误: %s", aiResponse.Error.Message)
	}

	if len(aiResponse.Choices) == 0 {
		return fmt.Errorf("AI响应为空")
	}

	aiReply := aiResponse.Choices[0].Message.Content
	if aiReply == "" {
		return fmt.Errorf("AI回复内容为空")
	}

	// 构建回复消息
	replyMsg := &entity.Message{
		MessageID:       uc.generateMessageID(),
		ChannelID:       msg.ChannelID,
		Direction:       entity.MessageDirectionOutbound,
		MessageType:     entity.MessageTypeText,
		SenderID:        msg.ReceiverID, // AI回复者变成发送者
		SenderName:      msg.ReceiverName,
		SenderType:      entity.SenderTypeBot,
		ReceiverID:      msg.SenderID, // 原发送者变成接收者
		ReceiverName:    msg.SenderName,
		ReceiverType:    msg.SenderType,
		Content:         aiReply,
		ConversationID:  msg.ConversationID,
		ParentMessageID: &msg.ID, // 设置父消息ID
	}

	// 发送AI回复消息
	if err := uc.SendMessage(ctx, replyMsg); err != nil {
		return fmt.Errorf("发送AI回复失败: %w", err)
	}

	uc.logger.Info("AI聊天处理完成", "reply_message_id", replyMsg.MessageID, "original_message_id", msg.MessageID, "model", modelName)
	return nil
}

// enqueueFailedMessage 将失败的消息加入重试队列
func (uc *messageUseCase) enqueueFailedMessage(ctx context.Context, msg *entity.Message) error {
	queue := &entity.MessageQueue{
		QueueName:   "message_retry",
		MessageID:   msg.MessageID,
		Priority:    100,
		MaxRetries:  3,
		Status:      entity.QueueStatusPending,
		ScheduledAt: time.Now().Add(5 * time.Minute), // 5分钟后重试
		Payload: entity.JSONField{
			"message_id":   msg.ID,
			"channel_id":   msg.ChannelID,
			"retry_reason": "processing_failed",
		},
	}

	return uc.queueRepo.Create(ctx, queue)
}
