package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
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
func (uc *messageUseCase) checkChannelStatus(ctx context.Context, channelID string) error {
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
func (uc *messageUseCase) GetMessage(ctx context.Context, id string) (*entity.Message, error) {
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
func (uc *messageUseCase) RetryFailedMessage(ctx context.Context, messageID string) error {
	uc.logger.Info("重试失败消息", "method", "RetryFailedMessage", "message_id", messageID)

	// 获取消息
	message, err := uc.messageRepo.GetByMessageID(ctx, messageID)
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
	// 1. 构建要发送的数据
	webhookData := map[string]interface{}{
		"message_id":          msg.MessageID,
		"channel_id":          msg.ChannelID,
		"message_type":        msg.MessageType,
		"sender_id":           msg.SenderID,
		"sender_name":         msg.SenderName,
		"sender_type":         msg.SenderType,
		"receiver_id":         msg.ReceiverID,
		"receiver_name":       msg.ReceiverName,
		"receiver_type":       msg.ReceiverType,
		"content":             msg.Content,
		"raw_content":         msg.RawContent,
		"media_url":           msg.MediaURL,
		"media_type":          msg.MediaType,
		"file_size":           msg.FileSize,
		"conversation_id":     msg.ConversationID,
		"platform_message_id": msg.PlatformMessageID,
		"received_at":         msg.ReceivedAt,
		"unified_content":     msg.UnifiedContent,
	}

	// 2. 发送HTTP POST请求到目标URL
	jsonData, err := json.Marshal(webhookData)
	if err != nil {
		return fmt.Errorf("序列化Webhook数据失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "OmniBotGo/1.0")

	// 检查是否配置了认证信息
	if authHeader := processor.GetConfigValue("auth_header"); authHeader != nil {
		if header, ok := authHeader.(string); ok && header != "" {
			req.Header.Set("Authorization", header)
		}
	}

	// 设置超时
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 发送请求
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	// 记录API调用日志
	apiLog := &entity.APICallLog{
		ChannelID:   &msg.ChannelID,
		ProcessorID: &processor.ID,
		RequestID:   msg.MessageID,
		Method:      "POST",
		URL:         url,
		RequestBody: string(jsonData),
		DurationMs:  int(duration.Milliseconds()),
	}

	if err != nil {
		apiLog.ResponseStatus = 0
		apiLog.ErrorMessage = err.Error()
		uc.logger.Error("发送Webhook请求失败", "error", err, "url", url, "message_id", msg.MessageID)
		// 如果请求失败，考虑加入重试队列
		if shouldRetryOnError(err) {
			return uc.enqueueWebhookMessage(ctx, msg, processor, url)
		}
		return fmt.Errorf("发送Webhook请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		uc.logger.Error("读取Webhook响应失败", "error", err)
		responseBody = []byte("failed to read response body")
	}

	// 更新API日志
	apiLog.ResponseStatus = resp.StatusCode
	apiLog.ResponseBody = string(responseBody)

	// 3. 处理响应并记录日志
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		uc.logger.Info("Webhook转发成功",
			"url", url,
			"message_id", msg.MessageID,
			"status_code", resp.StatusCode,
			"duration_ms", duration.Milliseconds())
	} else {
		uc.logger.Warn("Webhook转发返回非成功状态",
			"url", url,
			"message_id", msg.MessageID,
			"status_code", resp.StatusCode,
			"response", string(responseBody))

		// 对于某些状态码，可能需要重试
		if shouldRetryOnStatusCode(resp.StatusCode) {
			return uc.enqueueWebhookMessage(ctx, msg, processor, url)
		}
	}

	uc.logger.Info("Webhook转发处理完成", "url", url, "message_id", msg.MessageID)
	return nil
}

// shouldRetryOnError 判断错误是否应该重试
func shouldRetryOnError(err error) bool {
	// 对于网络错误、超时等临时错误，应该重试
	if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
		return true
	}
	// 检查是否是超时错误
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return false
}

// shouldRetryOnStatusCode 判断HTTP状态码是否应该重试
func shouldRetryOnStatusCode(statusCode int) bool {
	// 5xx 错误通常是服务器临时问题，应该重试
	if statusCode >= 500 && statusCode < 600 {
		return true
	}
	// 429 Too Many Requests 也应该重试
	if statusCode == 429 {
		return true
	}
	// 408 Request Timeout
	if statusCode == 408 {
		return true
	}
	return false
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

// processAIChat 处理AI聊天功能
func (uc *messageUseCase) processAIChat(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor) error {
	// 解析AI配置
	var aiConfig struct {
		Provider     string                   `json:"provider"`           // AI供应商：openai, claude, gemini
		Model        string                   `json:"model"`              // 模型名称
		SystemPrompt string                   `json:"system_prompt"`      // 系统提示词
		Temperature  float64                  `json:"temperature"`        // 温度参数
		MaxTokens    int                      `json:"max_tokens"`         // 最大token数
		APIKey       string                   `json:"api_key"`            // API密钥
		BaseURL      string                   `json:"base_url,omitempty"` // 自定义API地址
		EnableTools  bool                     `json:"enable_tools"`       // 是否启用工具调用
		Tools        []map[string]interface{} `json:"tools,omitempty"`    // 可用工具列表
		StreamMode   bool                     `json:"stream_mode"`        // 是否启用流式响应
	}

	// 将JSONField转换为字节数组
	configBytes, err := json.Marshal(processor.Config)
	if err != nil {
		uc.logger.Error("Failed to marshal processor config", "error", err)
		return fmt.Errorf("invalid processor config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &aiConfig); err != nil {
		uc.logger.Error("Failed to parse AI config", "error", err)
		return fmt.Errorf("invalid AI config: %w", err)
	}

	// 验证必要参数
	if aiConfig.Provider == "" {
		return fmt.Errorf("AI provider is required")
	}
	if aiConfig.Model == "" {
		return fmt.Errorf("AI model is required")
	}
	if aiConfig.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	// 设置默认值
	if aiConfig.Temperature == 0 {
		aiConfig.Temperature = 0.7
	}
	if aiConfig.MaxTokens == 0 {
		aiConfig.MaxTokens = 1024
	}
	if aiConfig.SystemPrompt == "" {
		aiConfig.SystemPrompt = "You are a helpful assistant."
	}

	// 获取会话历史
	history, err := uc.getConversationHistory(ctx, msg.ConversationID, 10)
	if err != nil {
		uc.logger.Error("Failed to get conversation history", "error", err)
		return fmt.Errorf("failed to get conversation history: %w", err)
	}

	// 构建消息列表
	messages := []*schema.Message{
		schema.SystemMessage(aiConfig.SystemPrompt),
	}

	// 添加历史消息
	for _, historyMsg := range history {
		if historyMsg.ID == msg.ID {
			continue // 跳过当前消息
		}
		if historyMsg.Direction == entity.MessageDirectionInbound {
			messages = append(messages, schema.UserMessage(historyMsg.Content))
		} else {
			// 检查是否是工具调用响应
			if toolCallData := uc.extractToolCallFromMessage(historyMsg); toolCallData != nil {
				messages = append(messages, schema.AssistantMessage(historyMsg.Content, toolCallData))
			} else {
				messages = append(messages, schema.AssistantMessage(historyMsg.Content, []schema.ToolCall{}))
			}
		}
	}

	// 添加当前用户消息
	messages = append(messages, schema.UserMessage(msg.Content))

	// 通过AdapterManager获取AI模型
	aiModelManager := uc.adapterManager.GetAIModelManager()

	// 将配置转换为map格式
	configMap := map[string]interface{}{
		"provider":      aiConfig.Provider,
		"model":         aiConfig.Model,
		"system_prompt": aiConfig.SystemPrompt,
		"temperature":   aiConfig.Temperature,
		"max_tokens":    aiConfig.MaxTokens,
		"api_key":       aiConfig.APIKey,
		"base_url":      aiConfig.BaseURL,
		"enable_tools":  aiConfig.EnableTools,
		"tools":         aiConfig.Tools,
		"stream_mode":   aiConfig.StreamMode,
	}

	chatModelInterface, err := aiModelManager.CreateChatModel(ctx, aiConfig.Provider, configMap)
	if err != nil {
		uc.logger.Error("Failed to create chat model", "error", err)
		return fmt.Errorf("failed to create chat model: %w", err)
	}

	// 类型断言为ToolCallingChatModel
	chatModel, ok := chatModelInterface.(model.ToolCallingChatModel)
	if !ok {
		return fmt.Errorf("created model does not implement ToolCallingChatModel interface")
	}

	// 根据配置选择处理方式
	if aiConfig.StreamMode {
		return uc.processAIChatStream(ctx, msg, processor, chatModel, messages, &aiConfig)
	} else {
		return uc.processAIChatSync(ctx, msg, processor, chatModel, messages, &aiConfig)
	}
}

// processAIChatSync 同步处理AI聊天
func (uc *messageUseCase) processAIChatSync(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor, chatModel model.ToolCallingChatModel, messages []*schema.Message, aiConfig *struct {
	Provider     string                   `json:"provider"`
	Model        string                   `json:"model"`
	SystemPrompt string                   `json:"system_prompt"`
	Temperature  float64                  `json:"temperature"`
	MaxTokens    int                      `json:"max_tokens"`
	APIKey       string                   `json:"api_key"`
	BaseURL      string                   `json:"base_url,omitempty"`
	EnableTools  bool                     `json:"enable_tools"`
	Tools        []map[string]interface{} `json:"tools,omitempty"`
	StreamMode   bool                     `json:"stream_mode"`
}) error {
	// 调用AI模型
	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		uc.logger.Error("Failed to call AI model", "error", err)
		return fmt.Errorf("failed to call AI model: %w", err)
	}

	// 处理工具调用
	if aiConfig.EnableTools && len(response.ToolCalls) > 0 {
		return uc.handleToolCalls(ctx, msg, processor, chatModel, messages, response, aiConfig)
	}

	// 获取AI响应内容
	aiResponse := response.Content
	if aiResponse == "" {
		return fmt.Errorf("empty response from AI model")
	}

	// 创建AI回复消息
	replyMsg := uc.createAIReplyMessage(msg, aiResponse, aiConfig, processor, nil)

	// 保存AI回复消息
	if err := uc.messageRepo.Create(ctx, replyMsg); err != nil {
		uc.logger.Error("Failed to save AI reply message", "error", err)
		return fmt.Errorf("failed to save AI reply message: %w", err)
	}

	uc.logger.Info("AI chat processed successfully",
		"provider", aiConfig.Provider,
		"model", aiConfig.Model,
		"message_id", msg.ID,
		"reply_id", replyMsg.ID,
		"response_length", len(aiResponse))

	return nil
}

// processAIChatStream 流式处理AI聊天
func (uc *messageUseCase) processAIChatStream(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor, chatModel model.ToolCallingChatModel, messages []*schema.Message, aiConfig *struct {
	Provider     string                   `json:"provider"`
	Model        string                   `json:"model"`
	SystemPrompt string                   `json:"system_prompt"`
	Temperature  float64                  `json:"temperature"`
	MaxTokens    int                      `json:"max_tokens"`
	APIKey       string                   `json:"api_key"`
	BaseURL      string                   `json:"base_url,omitempty"`
	EnableTools  bool                     `json:"enable_tools"`
	Tools        []map[string]interface{} `json:"tools,omitempty"`
	StreamMode   bool                     `json:"stream_mode"`
}) error {
	// 调用流式AI模型
	streamReader, err := chatModel.Stream(ctx, messages)
	if err != nil {
		uc.logger.Error("Failed to call AI model stream", "error", err)
		return fmt.Errorf("failed to call AI model stream: %w", err)
	}
	defer streamReader.Close()

	var fullResponse strings.Builder
	var toolCalls []schema.ToolCall

	// 读取流式响应
	for {
		chunk, err := streamReader.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			uc.logger.Error("Failed to read stream chunk", "error", err)
			return fmt.Errorf("failed to read stream chunk: %w", err)
		}

		// 累积响应内容
		if chunk.Content != "" {
			fullResponse.WriteString(chunk.Content)
		}

		// 收集工具调用
		if len(chunk.ToolCalls) > 0 {
			toolCalls = append(toolCalls, chunk.ToolCalls...)
		}

		// 这里可以实现实时推送到客户端的逻辑
		// 例如通过WebSocket或Server-Sent Events
		uc.logger.Debug("Received stream chunk",
			"content", chunk.Content,
			"tool_calls", len(chunk.ToolCalls))
	}

	finalResponse := fullResponse.String()
	if finalResponse == "" && len(toolCalls) == 0 {
		return fmt.Errorf("empty response from AI model stream")
	}

	// 处理工具调用
	if aiConfig.EnableTools && len(toolCalls) > 0 {
		// 创建带工具调用的响应
		response := &schema.Message{
			Content:   finalResponse,
			ToolCalls: toolCalls,
		}
		return uc.handleToolCalls(ctx, msg, processor, chatModel, messages, response, aiConfig)
	}

	// 创建AI回复消息
	replyMsg := uc.createAIReplyMessage(msg, finalResponse, aiConfig, processor, nil)

	// 保存AI回复消息
	if err := uc.messageRepo.Create(ctx, replyMsg); err != nil {
		uc.logger.Error("Failed to save AI reply message", "error", err)
		return fmt.Errorf("failed to save AI reply message: %w", err)
	}

	uc.logger.Info("AI chat stream processed successfully",
		"provider", aiConfig.Provider,
		"model", aiConfig.Model,
		"message_id", msg.ID,
		"reply_id", replyMsg.ID,
		"response_length", len(finalResponse))

	return nil
}

// handleToolCalls 处理工具调用
func (uc *messageUseCase) handleToolCalls(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor, chatModel model.ToolCallingChatModel, messages []*schema.Message, response *schema.Message, aiConfig *struct {
	Provider     string                   `json:"provider"`
	Model        string                   `json:"model"`
	SystemPrompt string                   `json:"system_prompt"`
	Temperature  float64                  `json:"temperature"`
	MaxTokens    int                      `json:"max_tokens"`
	APIKey       string                   `json:"api_key"`
	BaseURL      string                   `json:"base_url,omitempty"`
	EnableTools  bool                     `json:"enable_tools"`
	Tools        []map[string]interface{} `json:"tools,omitempty"`
	StreamMode   bool                     `json:"stream_mode"`
}) error {
	uc.logger.Info("Processing tool calls", "count", len(response.ToolCalls))

	// 添加助手的工具调用消息
	messages = append(messages, response)

	// 执行工具调用
	for _, toolCall := range response.ToolCalls {
		toolResult, err := uc.executeToolCall(ctx, &toolCall)
		if err != nil {
			uc.logger.Error("Failed to execute tool call", "tool", toolCall.Function.Name, "error", err)
			toolResult = fmt.Sprintf("Error executing tool %s: %v", toolCall.Function.Name, err)
		}

		// 添加工具结果消息
		toolMessage := schema.ToolMessage(toolResult, toolCall.ID)
		messages = append(messages, toolMessage)
		messages = append(messages, toolMessage)
	}

	// 再次调用AI模型获取最终响应
	finalResponse, err := chatModel.Generate(ctx, messages)
	if err != nil {
		uc.logger.Error("Failed to get final response after tool calls", "error", err)
		return fmt.Errorf("failed to get final response after tool calls: %w", err)
	}

	// 创建包含工具调用信息的回复消息
	toolCallsData := make([]map[string]interface{}, len(response.ToolCalls))
	for i, toolCall := range response.ToolCalls {
		toolCallsData[i] = map[string]interface{}{
			"id":       toolCall.ID,
			"function": toolCall.Function.Name,
			"args":     toolCall.Function.Arguments,
		}
	}

	replyMsg := uc.createAIReplyMessage(msg, finalResponse.Content, aiConfig, processor, toolCallsData)

	// 保存AI回复消息
	if err := uc.messageRepo.Create(ctx, replyMsg); err != nil {
		uc.logger.Error("Failed to save AI reply message with tool calls", "error", err)
		return fmt.Errorf("failed to save AI reply message with tool calls: %w", err)
	}

	uc.logger.Info("AI chat with tool calls processed successfully",
		"provider", aiConfig.Provider,
		"model", aiConfig.Model,
		"message_id", msg.ID,
		"reply_id", replyMsg.ID,
		"tool_calls", len(response.ToolCalls))

	return nil
}

// executeToolCall 执行工具调用
func (uc *messageUseCase) executeToolCall(ctx context.Context, toolCall *schema.ToolCall) (string, error) {
	uc.logger.Info("Executing tool call", "function", toolCall.Function.Name, "args", toolCall.Function.Arguments)

	// 根据工具名称执行相应的功能
	switch toolCall.Function.Name {
	case "get_current_time":
		return uc.toolGetCurrentTime(ctx, toolCall.Function.Arguments)
	case "get_weather":
		return uc.toolGetWeather(ctx, toolCall.Function.Arguments)
	case "search_web":
		return uc.toolSearchWeb(ctx, toolCall.Function.Arguments)
	case "send_message":
		return uc.toolSendMessage(ctx, toolCall.Function.Arguments)
	default:
		return "", fmt.Errorf("unknown tool function: %s", toolCall.Function.Name)
	}
}

// toolGetCurrentTime 获取当前时间工具
func (uc *messageUseCase) toolGetCurrentTime(ctx context.Context, args string) (string, error) {
	now := time.Now()
	return fmt.Sprintf("Current time: %s", now.Format("2006-01-02 15:04:05 MST")), nil
}

// toolGetWeather 获取天气工具
func (uc *messageUseCase) toolGetWeather(ctx context.Context, args string) (string, error) {
	// 解析参数
	var params struct {
		Location string `json:"location"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid weather tool arguments: %w", err)
	}

	// 这里应该调用实际的天气API
	// 暂时返回模拟数据
	return fmt.Sprintf("Weather in %s: Sunny, 25°C", params.Location), nil
}

// toolSearchWeb 网络搜索工具
func (uc *messageUseCase) toolSearchWeb(ctx context.Context, args string) (string, error) {
	// 解析参数
	var params struct {
		Query string `json:"query"`
		Count int    `json:"count,omitempty"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid search tool arguments: %w", err)
	}

	if params.Count == 0 {
		params.Count = 3
	}

	// 这里应该调用实际的搜索API
	// 暂时返回模拟数据
	return fmt.Sprintf("Search results for '%s': Found %d relevant results", params.Query, params.Count), nil
}

// toolSendMessage 发送消息工具
func (uc *messageUseCase) toolSendMessage(ctx context.Context, args string) (string, error) {
	// 解析参数
	var params struct {
		ChannelID string `json:"channel_id"`
		Content   string `json:"content"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid send message tool arguments: %w", err)
	}

	// 创建新消息
	newMsg := &entity.Message{
		ChannelID:         params.ChannelID,
		PlatformMessageID: fmt.Sprintf("tool_msg_%d", time.Now().Unix()),
		Direction:         entity.MessageDirectionOutbound,
		MessageType:       "text",
		ContentType:       "text",
		SenderID:          "ai_tool",
		SenderName:        "AI Tool",
		SenderType:        "bot",
		Content:           params.Content,
		MessageStatus:     entity.MessageStatusProcessed,
		PlatformTimestamp: time.Now(),
	}

	// 发送消息
	if err := uc.SendMessage(ctx, newMsg); err != nil {
		return "", fmt.Errorf("failed to send message via tool: %w", err)
	}

	return fmt.Sprintf("Message sent successfully to channel %s", params.ChannelID), nil
}

// createAIReplyMessage 创建AI回复消息
func (uc *messageUseCase) createAIReplyMessage(msg *entity.Message, content string, aiConfig *struct {
	Provider     string                   `json:"provider"`
	Model        string                   `json:"model"`
	SystemPrompt string                   `json:"system_prompt"`
	Temperature  float64                  `json:"temperature"`
	MaxTokens    int                      `json:"max_tokens"`
	APIKey       string                   `json:"api_key"`
	BaseURL      string                   `json:"base_url,omitempty"`
	EnableTools  bool                     `json:"enable_tools"`
	Tools        []map[string]interface{} `json:"tools,omitempty"`
	StreamMode   bool                     `json:"stream_mode"`
}, processor *entity.MessageProcessor, toolCalls []map[string]interface{}) *entity.Message {
	unifiedContent := entity.JSONField{
		"ai_provider":  aiConfig.Provider,
		"ai_model":     aiConfig.Model,
		"processor_id": processor.ID,
		"stream_mode":  aiConfig.StreamMode,
	}

	if toolCalls != nil {
		unifiedContent["tool_calls"] = toolCalls
	}

	return &entity.Message{
		ChannelID:         msg.ChannelID,
		PlatformMessageID: fmt.Sprintf("ai_reply_%d_%d", msg.ID, time.Now().Unix()),
		Direction:         entity.MessageDirectionOutbound,
		MessageType:       msg.MessageType,
		ContentType:       "text",
		SenderID:          "ai_bot",
		SenderName:        fmt.Sprintf("AI Bot (%s)", aiConfig.Model),
		SenderType:        "bot",
		ReceiverID:        msg.SenderID,
		ReceiverName:      msg.SenderName,
		ReceiverType:      msg.SenderType,
		Content:           content,
		ConversationID:    msg.ConversationID,
		ParentMessageID:   &msg.ID,
		UnifiedContent:    unifiedContent,
		MessageStatus:     entity.MessageStatusProcessed,
		PlatformTimestamp: time.Now(),
		ProcessedAt:       &[]time.Time{time.Now()}[0],
	}
}

// extractToolCallFromMessage 从消息中提取工具调用信息
func (uc *messageUseCase) extractToolCallFromMessage(msg *entity.Message) []schema.ToolCall {
	if msg.UnifiedContent == nil {
		return nil
	}

	toolCallsData, exists := msg.UnifiedContent["tool_calls"]
	if !exists {
		return nil
	}

	// 尝试转换为工具调用列表
	toolCallsSlice, ok := toolCallsData.([]interface{})
	if !ok {
		return nil
	}

	var toolCalls []schema.ToolCall
	for _, item := range toolCallsSlice {
		toolCallMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := toolCallMap["id"].(string)
		functionName, _ := toolCallMap["function"].(string)
		args, _ := toolCallMap["args"].(string)

		toolCall := schema.ToolCall{
			ID: id,
			Function: schema.FunctionCall{
				Name:      functionName,
				Arguments: args,
			},
		}
		toolCalls = append(toolCalls, toolCall)
	}

	return toolCalls
}

// getConversationHistory 获取会话历史消息
func (uc *messageUseCase) getConversationHistory(ctx context.Context, conversationID string, limit int) ([]*entity.Message, error) {
	if conversationID == "" {
		return []*entity.Message{}, nil
	}

	// 使用GetByConversationID方法查询
	params := port.ListParams{
		Page:     1,
		PageSize: limit,
		OrderBy:  "created_at",
		Filters: map[string]interface{}{
			"conversation_id": conversationID,
		},
	}

	result, err := uc.messageRepo.GetByConversationID(ctx, conversationID, params)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
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
