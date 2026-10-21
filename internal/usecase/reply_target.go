package usecase

import (
	"strings"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

const (
	replySenderBotID           = "bot"
	rawKeySessionWebhook       = "sessionWebhook"
	rawKeySessionWebhookAlt    = "session_webhook"
	rawKeySessionWebhookExpiry = "sessionWebhookExpiredTime"
	rawKeyChatID               = "chatId"
	rawKeyChatIDAlt            = "chat_id"
	rawKeyConversationID       = "conversationId"
	rawKeyConversationType     = "conversationType"
	rawKeyOpenConversationID   = "openConversationId"
	dingTalkGroupConvType      = "2"
)

// BuildOutboundReplyFromInbound builds an outbound reply that targets the correct
// DM or group conversation (T04). Group inbound keeps group as receiver; DM
// inbound replies to the original sender. Proactive SendMessage stays separate.
func BuildOutboundReplyFromInbound(inbound *entity.Message, content, messageType string) *entity.Message {
	if messageType == "" {
		messageType = entity.MessageTypeText
	}

	parentID := inbound.ID
	reply := &entity.Message{
		ChannelID:         inbound.ChannelID,
		Direction:         entity.MessageDirectionOutbound,
		MessageType:       messageType,
		Content:           content,
		ConversationID:    inbound.ConversationID,
		ParentMessageID:   &parentID,
		PlatformTimestamp: time.Now(),
		MessageStatus:     entity.MessageStatusPending,
		RawContent:        copyReplyRawContent(inbound.RawContent),
		SenderID:          replySenderBotID,
		SenderName:        inbound.ReceiverName,
		SenderType:        entity.SenderTypeBot,
	}

	if isGroupConversation(inbound) {
		applyGroupReplyTarget(reply, inbound)

		return reply
	}

	applyDMReplyTarget(reply, inbound)

	return reply
}

func applyGroupReplyTarget(reply, inbound *entity.Message) {
	groupID := inbound.ConversationID
	if groupID == "" || strings.EqualFold(inbound.ReceiverType, entity.ReceiverTypeGroup) {
		groupID = firstNonEmpty(inbound.ReceiverID, inbound.ConversationID)
	}

	if groupID == "" {
		groupID = inbound.ReceiverID
	}

	reply.ReceiverID = groupID
	reply.ReceiverName = inbound.ReceiverName
	reply.ReceiverType = entity.ReceiverTypeGroup
}

func applyDMReplyTarget(reply, inbound *entity.Message) {
	reply.ReceiverID = inbound.SenderID
	reply.ReceiverName = inbound.SenderName
	reply.ReceiverType = entity.ReceiverTypeUser

	if inbound.SenderType != "" && inbound.SenderType != entity.SenderTypeGroup {
		reply.ReceiverType = inbound.SenderType
		if reply.ReceiverType == entity.SenderTypeUser {
			reply.ReceiverType = entity.ReceiverTypeUser
		}
	}

	if inbound.ReceiverID != "" {
		reply.SenderID = inbound.ReceiverID
	}
}

func isGroupConversation(msg *entity.Message) bool {
	if strings.EqualFold(msg.ReceiverType, entity.ReceiverTypeGroup) {
		return true
	}

	if strings.EqualFold(msg.SenderType, entity.SenderTypeGroup) {
		return true
	}

	return rawConversationIsGroup(msg.RawContent)
}

func rawConversationIsGroup(raw entity.JSONField) bool {
	if raw == nil {
		return false
	}

	v, ok := raw[rawKeyConversationType]
	if !ok {
		return false
	}

	switch typed := v.(type) {
	case string:
		return typed == dingTalkGroupConvType || strings.EqualFold(typed, entity.ReceiverTypeGroup)
	case float64:
		return int(typed) == 2 //nolint:mnd // DingTalk conversationType 2 = group
	case int:
		return typed == 2 //nolint:mnd // DingTalk conversationType 2 = group
	default:
		return false
	}
}

func copyReplyRawContent(src entity.JSONField) entity.JSONField {
	if src == nil {
		return nil
	}

	keys := []string{
		rawKeySessionWebhook,
		rawKeySessionWebhookExpiry,
		rawKeySessionWebhookAlt,
		rawKeyChatID,
		rawKeyChatIDAlt,
		rawKeyConversationID,
		rawKeyConversationType,
		rawKeyOpenConversationID,
	}

	out := make(entity.JSONField)

	for _, key := range keys {
		if v, ok := src[key]; ok {
			out[key] = v
		}
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}

	return ""
}
