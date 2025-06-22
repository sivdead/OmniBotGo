// Package service 提供业务逻辑服务
package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

// MessageService 消息业务逻辑服务
type MessageService struct{}

// NewMessageService 创建消息服务实例
func NewMessageService() *MessageService {
	return &MessageService{}
}

// IsInbound 检查是否为入站消息
func (s *MessageService) IsInbound(m *entity.Message) bool {
	return m.Direction == entity.MessageDirectionInbound
}

// IsOutbound 检查是否为出站消息
func (s *MessageService) IsOutbound(m *entity.Message) bool {
	return m.Direction == entity.MessageDirectionOutbound
}

// IsProcessed 检查消息是否已处理
func (s *MessageService) IsProcessed(m *entity.Message) bool {
	return m.MessageStatus == entity.MessageStatusProcessed ||
		m.MessageStatus == entity.MessageStatusSent
}

// CanRetry 检查消息是否可以重试
func (s *MessageService) CanRetry(m *entity.Message) bool {
	return m.MessageStatus == entity.MessageStatusFailed && m.RetryCount < 3
}

// MarkAsProcessing 标记消息为处理中
func (s *MessageService) MarkAsProcessing(m *entity.Message) {
	m.MessageStatus = entity.MessageStatusProcessing
}

// MarkAsProcessed 标记消息为已处理
func (s *MessageService) MarkAsProcessed(m *entity.Message) {
	m.MessageStatus = entity.MessageStatusProcessed
	now := time.Now()
	m.ProcessedAt = &now
}

// MarkAsSent 标记消息为已发送
func (s *MessageService) MarkAsSent(m *entity.Message) {
	m.MessageStatus = entity.MessageStatusSent
	now := time.Now()
	m.SentAt = &now
}

// MarkAsFailed 标记消息为失败并记录错误
func (s *MessageService) MarkAsFailed(m *entity.Message, errorMsg string) {
	m.MessageStatus = entity.MessageStatusFailed
	m.ErrorMessage = errorMsg
	m.RetryCount++
}

// HasMedia 检查消息是否包含媒体文件
func (s *MessageService) HasMedia(m *entity.Message) bool {
	return m.MediaURL != ""
}

// IsReply 检查是否为回复消息
func (s *MessageService) IsReply(m *entity.Message) bool {
	return m.ParentMessageID != nil
}

// GetRawContentValue 获取原始内容中的特定值
func (s *MessageService) GetRawContentValue(m *entity.Message, key string) interface{} {
	if m.RawContent == nil {
		return nil
	}
	return m.RawContent.Get(key)
}

// GetUnifiedContentValue 获取统一内容中的特定值
func (s *MessageService) GetUnifiedContentValue(m *entity.Message, key string) interface{} {
	if m.UnifiedContent == nil {
		return nil
	}
	return m.UnifiedContent.Get(key)
}

// ValidateMessage 验证Message实体数据
func (s *MessageService) ValidateMessage(m *entity.Message) error {
	if m.MessageID == "" {
		return entity.NewValidationError("message_id", "消息ID不能为空")
	}
	if m.ChannelID <= 0 {
		return entity.NewValidationError("channel_id", "通道ID必须大于0")
	}
	if m.MessageType == "" {
		return entity.NewValidationError("message_type", "消息类型不能为空")
	}

	// 验证特定消息类型的必填字段
	switch m.MessageType {
	case entity.MessageTypeMarkdown:
		return s.validateMarkdownMessage(m)
	case entity.MessageTypeCard:
		return s.validateCardMessage(m)
	case entity.MessageTypeNews:
		return s.validateNewsMessage(m)
	case entity.MessageTypeFile:
		return s.validateFileMessage(m)
	case entity.MessageTypeLocation:
		return s.validateLocationMessage(m)
	case entity.MessageTypeEvent:
		return s.validateEventMessage(m)
	case entity.MessageTypeTemplate:
		return s.validateTemplateMessage(m)
	}

	return nil
}

// 验证Markdown消息
func (s *MessageService) validateMarkdownMessage(m *entity.Message) error {
	content := s.GetUnifiedContentValue(m, "markdown")
	if content == nil {
		return entity.NewValidationError("content", "Markdown消息内容不能为空")
	}
	return nil
}

// 验证卡片消息
func (s *MessageService) validateCardMessage(m *entity.Message) error {
	content := s.GetUnifiedContentValue(m, "card")
	if content == nil {
		return entity.NewValidationError("content", "卡片消息内容不能为空")
	}
	return nil
}

// 验证图文消息
func (s *MessageService) validateNewsMessage(m *entity.Message) error {
	content := s.GetUnifiedContentValue(m, "news")
	if content == nil {
		return entity.NewValidationError("content", "图文消息内容不能为空")
	}
	return nil
}

// 验证文件消息
func (s *MessageService) validateFileMessage(m *entity.Message) error {
	if m.MediaURL == "" {
		return entity.NewValidationError("media_url", "文件URL不能为空")
	}
	return nil
}

// 验证位置消息
func (s *MessageService) validateLocationMessage(m *entity.Message) error {
	content := s.GetUnifiedContentValue(m, "location")
	if content == nil {
		return entity.NewValidationError("content", "位置消息内容不能为空")
	}
	return nil
}

// 验证事件消息
func (s *MessageService) validateEventMessage(m *entity.Message) error {
	content := s.GetUnifiedContentValue(m, "event")
	if content == nil {
		return entity.NewValidationError("content", "事件消息内容不能为空")
	}
	return nil
}

// 验证模板消息
func (s *MessageService) validateTemplateMessage(m *entity.Message) error {
	content := s.GetUnifiedContentValue(m, "template")
	if content == nil {
		return entity.NewValidationError("content", "模板消息内容不能为空")
	}
	return nil
}

// ParseMarkdownContent 解析Markdown消息内容
func (s *MessageService) ParseMarkdownContent(content interface{}) (*entity.MarkdownMessage, error) {
	var markdown entity.MarkdownMessage

	// 尝试从map或JSON字符串解析
	switch v := content.(type) {
	case map[string]interface{}:
		if title, ok := v["title"].(string); ok {
			markdown.Title = title
		}
		if content, ok := v["content"].(string); ok {
			markdown.Content = content
		} else {
			return nil, fmt.Errorf("Markdown内容缺少content字段")
		}
	case string:
		// 尝试从JSON字符串解析
		if err := json.Unmarshal([]byte(v), &markdown); err != nil {
			// 如果解析失败，将整个字符串作为内容
			markdown.Content = v
		}
	default:
		return nil, fmt.Errorf("无效的Markdown内容格式")
	}

	return &markdown, nil
}

// ParseCardContent 解析卡片消息内容
func (s *MessageService) ParseCardContent(content interface{}) (*entity.CardMessage, error) {
	var card entity.CardMessage

	// 尝试从map或JSON字符串解析
	switch v := content.(type) {
	case map[string]interface{}:
		if cardType, ok := v["card_type"].(string); ok {
			card.CardType = cardType
		}
		if title, ok := v["title"].(string); ok {
			card.Title = title
		}
		if content, ok := v["content"].(string); ok {
			card.Content = content
		}
		// 解析按钮
		if buttons, ok := v["buttons"].([]interface{}); ok {
			for _, btn := range buttons {
				if btnMap, ok := btn.(map[string]interface{}); ok {
					button := entity.CardButton{}
					if title, ok := btnMap["title"].(string); ok {
						button.Title = title
					}
					if actionURL, ok := btnMap["action_url"].(string); ok {
						button.ActionURL = actionURL
					}
					if actionType, ok := btnMap["action_type"].(string); ok {
						button.ActionType = actionType
					}
					card.Buttons = append(card.Buttons, button)
				}
			}
		}
		// 解析扩展字段
		if extra, ok := v["extra"].(map[string]interface{}); ok {
			card.Extra = extra
		}
	case string:
		// 尝试从JSON字符串解析
		if err := json.Unmarshal([]byte(v), &card); err != nil {
			return nil, fmt.Errorf("解析卡片内容失败: %w", err)
		}
	default:
		return nil, fmt.Errorf("无效的卡片内容格式")
	}

	return &card, nil
}

// ParseEventContent 解析事件消息内容
func (s *MessageService) ParseEventContent(content interface{}) (*entity.EventMessage, error) {
	var event entity.EventMessage

	// 尝试从map或JSON字符串解析
	switch v := content.(type) {
	case map[string]interface{}:
		if eventType, ok := v["event_type"].(string); ok {
			event.EventType = eventType
		} else {
			return nil, fmt.Errorf("事件消息缺少event_type字段")
		}
		if eventKey, ok := v["event_key"].(string); ok {
			event.EventKey = eventKey
		}
		if eventData, ok := v["event_data"].(map[string]interface{}); ok {
			event.EventData = eventData
		}
	case string:
		// 尝试从JSON字符串解析
		if err := json.Unmarshal([]byte(v), &event); err != nil {
			return nil, fmt.Errorf("解析事件内容失败: %w", err)
		}
	default:
		return nil, fmt.Errorf("无效的事件内容格式")
	}

	return &event, nil
}

// IsRichMessageType 判断是否为富文本消息类型
func (s *MessageService) IsRichMessageType(messageType string) bool {
	return entity.IsRichMessageType(messageType)
}

// IsEventMessageType 判断是否为事件消息类型
func (s *MessageService) IsEventMessageType(messageType string) bool {
	return entity.IsEventMessageType(messageType)
}

// IsMediaMessageType 判断是否为媒体消息类型
func (s *MessageService) IsMediaMessageType(messageType string) bool {
	return entity.IsMediaMessageType(messageType)
}
