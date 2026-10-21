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
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

// Routing and processors (T03)

// routeAndProcessMessage 路由并处理消息。
// B02: 任一处理器失败则不 MarkAsProcessed；广播模式下已成功的目标记录后重试时跳过。
func (uc *messageUseCase) routeAndProcessMessage(ctx context.Context, msg *entity.Message) error {
	processors, err := uc.routingUC.RouteMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("消息路由失败: %w", err)
	}

	if len(processors) == 0 {
		uc.logger.Warn("没有找到匹配的处理器", "message_id", msg.MessageID)
		uc.msgService.MarkAsProcessed(msg)

		if err := uc.messageRepo.Update(ctx, msg); err != nil {
			return fmt.Errorf("更新消息状态失败: %w", err)
		}

		return nil
	}

	done := successfulProcessorSet(msg)

	var failures []error

	for _, processor := range processors {
		if done[processor.ID] {
			uc.logger.Info("跳过已成功的广播目标", "processor_id", processor.ID, "message_id", msg.MessageID)

			continue
		}

		if err := uc.processMessageWithProcessor(ctx, msg, processor); err != nil {
			uc.logger.Error("处理器处理消息失败", "processor_id", processor.ID, "processor_name", processor.ProcessorName, "error", err)
			failures = append(failures, fmt.Errorf("%s: %w", processor.ID, err))

			continue
		}

		markProcessorSuccessful(msg, processor.ID)

		if err := uc.messageRepo.Update(ctx, msg); err != nil {
			uc.logger.Error("持久化成功处理器记录失败", "error", err)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("processor failures (%d): %w", len(failures), errors.Join(failures...))
	}

	uc.msgService.MarkAsProcessed(msg)

	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	return nil
}

func successfulProcessorSet(msg *entity.Message) map[string]bool {
	out := make(map[string]bool)
	if msg.UnifiedContent == nil {
		return out
	}

	raw, ok := msg.UnifiedContent[successfulProcessorsKey]
	if !ok {
		return out
	}

	switch v := raw.(type) {
	case []string:
		for _, id := range v {
			out[id] = true
		}
	case []interface{}:
		for _, item := range v {
			if id, ok := item.(string); ok {
				out[id] = true
			}
		}
	}

	return out
}

func markProcessorSuccessful(msg *entity.Message, processorID string) {
	if msg.UnifiedContent == nil {
		msg.UnifiedContent = make(entity.JSONField)
	}

	set := successfulProcessorSet(msg)

	if set[processorID] {
		return
	}

	ids := make([]string, 0, len(set)+1)

	for id := range set {
		ids = append(ids, id)
	}

	ids = append(ids, processorID)
	msg.UnifiedContent[successfulProcessorsKey] = ids
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
	webhookURL := processor.GetConfigValue(configKeyWebhookURL)
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
		ChannelID:   msg.ChannelID,
		ProcessorID: processor.ID,
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

	// B05: any non-2xx is a delivery failure; retryability must not be treated as success.
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		uc.logger.Info("Webhook转发成功",
			"url", url,
			"message_id", msg.MessageID,
			"status_code", resp.StatusCode,
			"duration_ms", duration.Milliseconds())

		return nil
	}

	uc.logger.Warn("Webhook转发返回非成功状态",
		"url", url,
		"message_id", msg.MessageID,
		"status_code", resp.StatusCode,
		"response", string(responseBody))

	if shouldRetryOnStatusCode(resp.StatusCode) {
		if enqErr := uc.enqueueWebhookMessage(ctx, msg, processor, url); enqErr != nil {
			uc.logger.Error("非2xx后入队重试失败", "error", enqErr)
		}
	}

	return fmt.Errorf("%w: HTTP %d", errWebhookDeliveryFailed, resp.StatusCode)
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
			"message_id":          msg.ID,
			payloadKeyProcessorID: processor.ID,
			configKeyWebhookURL:   webhookURL,
			"message_data":        msg,
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

	// T04: shared DM/group targeting (no blind sender↔receiver swap).
	replyMsg := BuildOutboundReplyFromInbound(msg, content, entity.MessageTypeText)

	// 发送回复消息
	if err := uc.SendMessage(ctx, replyMsg); err != nil {
		return fmt.Errorf("发送自动回复失败: %w", err)
	}

	uc.logger.Info("自动回复处理完成", "reply_message_id", replyMsg.MessageID, "original_message_id", msg.MessageID)
	return nil
}
