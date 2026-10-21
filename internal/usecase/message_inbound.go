package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

// Inbound message path (T03)

// CreateStreamMessageHandler 创建用于Stream适配器的消息处理回调函数.
func (uc *messageUseCase) CreateStreamMessageHandler() port.MessageHandler {
	return func(ctx context.Context, unifiedMessage *dto.UnifiedMessage) error {
		uc.logger.Info("Stream消息处理回调", "message_id", unifiedMessage.MessageID, "sender_id", unifiedMessage.SenderID)

		channelID := ""

		if unifiedMessage.RawContent != nil {
			if v, ok := unifiedMessage.RawContent["channel_id"].(string); ok {
				channelID = v
			}
		}

		// T02: shared Unified → entity conversion (same as Webhook path)
		return uc.ProcessInboundMessage(ctx, unifiedMessage.ToInboundEntity(channelID))
	}
}

// ProcessInboundMessage 处理入站消息.
func (uc *messageUseCase) ProcessInboundMessage(ctx context.Context, msg *entity.Message) error {
	uc.logger.Info("开始处理入站消息", "method", "ProcessInboundMessage", "message_id", msg.MessageID, "channel_id", msg.ChannelID)

	if err := uc.checkChannelStatus(ctx, msg.ChannelID); err != nil {
		return err
	}

	// Prepare (incl. MessageID) before ValidateMessage so empty ID is not rejected.
	if err := uc.prepareInboundMessage(msg); err != nil {
		return err
	}

	if err := uc.validateInboundMessage(ctx, msg); err != nil {
		return err
	}

	if skip, err := uc.checkDuplicateMessage(ctx, msg); err != nil {
		return err
	} else if skip {
		return nil // 已成功处理过的重复消息，静默跳过
	}

	if err := uc.persistAndProcessMessage(ctx, msg); err != nil {
		return err
	}

	uc.logger.Info("入站消息处理完成")

	return nil
}

// validateInboundMessage 验证入站消息.
func (uc *messageUseCase) validateInboundMessage(_ context.Context, msg *entity.Message) error {
	if err := uc.msgService.ValidateMessage(msg); err != nil {
		uc.logger.Error("消息验证失败", "error", err)

		return fmt.Errorf("消息验证失败: %w", err)
	}

	return nil
}

// checkChannelStatus 检查通道状态.
func (uc *messageUseCase) checkChannelStatus(ctx context.Context, channelID string) error {
	channel, err := uc.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		uc.logger.Error("获取通道信息失败", "error", err)

		return fmt.Errorf("获取通道信息失败: %w", err)
	}

	if !channel.IsActive() {
		uc.logger.Warn("通道未激活，拒绝处理消息")

		return errChannelInactive
	}

	return nil
}

// prepareInboundMessage 准备入站消息.
func (uc *messageUseCase) prepareInboundMessage(msg *entity.Message) error {
	msg.Direction = entity.MessageDirectionInbound

	if msg.MessageStatus == 0 {
		msg.MessageStatus = entity.MessageStatusPending
	}

	if msg.ReceivedAt.IsZero() {
		msg.ReceivedAt = time.Now()
	}

	if msg.PlatformTimestamp.IsZero() {
		msg.PlatformTimestamp = time.Now()
	}

	if msg.MessageID == "" {
		msg.MessageID = uc.generateMessageID()
	}

	return nil
}

// checkDuplicateMessage 检查重复消息。
// 仅对最终成功（processed/sent）去重跳过；pending/failed/processing 恢复同一行继续处理。
func (uc *messageUseCase) checkDuplicateMessage(ctx context.Context, msg *entity.Message) (bool, error) {
	if msg.PlatformMessageID == "" {
		return false, nil
	}

	existingMsg, err := uc.findExistingByPlatformMessageID(ctx, msg)
	if err != nil {
		return false, err
	}

	if existingMsg == nil {
		return false, nil
	}

	if existingMsg.MessageStatus.IsProcessed() {
		uc.logger.Warn("检测到已成功处理的重复消息，跳过",
			"message_id", msg.MessageID,
			"platform_message_id", msg.PlatformMessageID,
			"existing_message_id", existingMsg.MessageID,
			"status", existingMsg.MessageStatus.String())

		return true, nil
	}

	// Resume unfinished row instead of dropping redelivery.
	uc.logger.Info("恢复未完成消息处理",
		"existing_id", existingMsg.ID,
		"status", existingMsg.MessageStatus.String(),
		"platform_message_id", msg.PlatformMessageID)

	msg.ID = existingMsg.ID
	msg.MessageID = existingMsg.MessageID
	msg.RetryCount = existingMsg.RetryCount

	if existingMsg.UnifiedContent != nil {
		if msg.UnifiedContent == nil {
			msg.UnifiedContent = make(entity.JSONField)
		}

		if ids, ok := existingMsg.UnifiedContent[successfulProcessorsKey]; ok {
			msg.UnifiedContent[successfulProcessorsKey] = ids
		}
	}

	return false, nil
}

// findExistingByPlatformMessageID looks up by current PlatformMessageID and,
// for chat_id:message_id form, also legacy message_id-only rows in the same conversation.
func (uc *messageUseCase) findExistingByPlatformMessageID(ctx context.Context, msg *entity.Message) (*entity.Message, error) {
	existingMsg, err := uc.messageRepo.GetByPlatformMessageID(ctx, msg.ChannelID, msg.PlatformMessageID)
	if err != nil {
		return nil, fmt.Errorf("查询平台消息失败: %w", err)
	}

	if existingMsg != nil {
		return existingMsg, nil
	}

	// Legacy Telegram-style IDs were message_id only; new form is chat_id:message_id.
	if idx := strings.LastIndex(msg.PlatformMessageID, ":"); idx > 0 {
		legacyID := msg.PlatformMessageID[idx+1:]

		legacy, err := uc.messageRepo.GetByPlatformMessageID(ctx, msg.ChannelID, legacyID)
		if err != nil {
			return nil, fmt.Errorf("查询遗留平台消息失败: %w", err)
		}

		if legacy != nil && (msg.ConversationID == "" || legacy.ConversationID == "" || legacy.ConversationID == msg.ConversationID) {
			return legacy, nil
		}
	}

	return nil, nil
}

// persistAndProcessMessage 持久化并处理消息.
func (uc *messageUseCase) persistAndProcessMessage(ctx context.Context, msg *entity.Message) error {
	if msg.ID == "" {
		if err := uc.messageRepo.Create(ctx, msg); err != nil {
			uc.logger.Error("保存消息失败", "error", err)

			return fmt.Errorf("保存消息失败: %w", err)
		}
	}

	uc.msgService.MarkAsProcessing(msg)

	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		uc.logger.Error("更新消息状态失败", "error", err)

		return fmt.Errorf("更新消息状态失败: %w", err)
	}

	if err := uc.routeAndProcessMessage(ctx, msg); err != nil {
		uc.logger.Error("消息路由处理失败", "error", err)
		uc.handleMessageProcessingError(ctx, msg, err)

		return err // propagate: do not ACK upstream as success
	}

	return nil
}

// handleMessageProcessingError 处理消息处理错误.
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
