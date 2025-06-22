// Package dto 定义数据传输对象
package dto

import (
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

// UnifiedMessage 统一消息格式，用于平台适配器和业务逻辑之间的数据传输
type UnifiedMessage struct {
	// 基本信息
	MessageID   string `json:"message_id"`
	MessageType string `json:"message_type"` // text, image, audio, video, file, event, markdown, card, etc.

	// 发送者信息
	SenderID   string `json:"sender_id"`
	SenderName string `json:"sender_name"`
	SenderType string `json:"sender_type"` // user, group, system, bot

	// 接收者信息
	ReceiverID   string `json:"receiver_id"`
	ReceiverName string `json:"receiver_name"`
	ReceiverType string `json:"receiver_type"` // user, group, all

	// 消息内容
	Content    string                 `json:"content"`
	RawContent map[string]interface{} `json:"raw_content"`

	// 媒体信息
	MediaURL  string `json:"media_url,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	FileSize  int64  `json:"file_size,omitempty"`

	// 会话信息
	ConversationID string `json:"conversation_id"`

	// 回复信息
	ParentMessageID *string `json:"parent_message_id,omitempty"`

	// 平台信息
	PlatformMessageID string    `json:"platform_message_id"`
	PlatformTimestamp time.Time `json:"platform_timestamp"`

	// 富文本消息内容
	MarkdownContent *entity.MarkdownMessage `json:"markdown_content,omitempty"`
	CardContent     *entity.CardMessage     `json:"card_content,omitempty"`
	NewsContent     *entity.NewsMessage     `json:"news_content,omitempty"`
	FileContent     *entity.FileMessage     `json:"file_content,omitempty"`
	LocationContent *entity.LocationMessage `json:"location_content,omitempty"`
	EventContent    *entity.EventMessage    `json:"event_content,omitempty"`
	TemplateContent *entity.TemplateMessage `json:"template_content,omitempty"`
}

// IsRichMessage 判断是否为富文本消息
func (m *UnifiedMessage) IsRichMessage() bool {
	return entity.IsRichMessageType(m.MessageType)
}

// IsEventMessage 判断是否为事件消息
func (m *UnifiedMessage) IsEventMessage() bool {
	return entity.IsEventMessageType(m.MessageType)
}

// IsMediaMessage 判断是否为媒体消息
func (m *UnifiedMessage) IsMediaMessage() bool {
	return entity.IsMediaMessageType(m.MessageType)
}

// GetEventType 获取事件类型（仅对事件消息有效）
func (m *UnifiedMessage) GetEventType() string {
	if m.EventContent != nil {
		return m.EventContent.EventType
	}
	return ""
}

// GetCardType 获取卡片类型（仅对卡片消息有效）
func (m *UnifiedMessage) GetCardType() string {
	if m.CardContent != nil {
		return m.CardContent.CardType
	}
	return ""
}
