package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

// FeishuAdapter 飞书平台适配器
type FeishuAdapter struct {
	client *lark.Client
}

// NewFeishuAdapter 创建飞书适配器实例
func NewFeishuAdapter() *FeishuAdapter {
	return &FeishuAdapter{}
}

// GetPlatformType 获取平台类型
func (f *FeishuAdapter) GetPlatformType() entity.PlatformType {
	return entity.PlatformTypeFeishu
}

// ValidateConfig 验证平台配置
func (f *FeishuAdapter) ValidateConfig(cfg map[string]interface{}) error {
	return config.ValidatePlatformConfig(entity.PlatformTypeFeishu, cfg)
}

// SendMessage 发送消息
func (f *FeishuAdapter) SendMessage(ctx context.Context, message *dto.UnifiedMessage, cfg map[string]interface{}, accessToken string) error {
	client, err := f.getClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to get feishu client: %w", err)
	}

	// 构建飞书消息
	content, err := f.buildMessageContent(message)
	if err != nil {
		return fmt.Errorf("failed to build message content: %w", err)
	}

	// 发送消息
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeUserId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(message.ReceiverID).
			MsgType(f.mapToFeishuMessageType(message.MessageType)).
			Content(content).
			Build()).
		Build()

	resp, err := client.Im.Message.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	if !resp.Success() {
		return fmt.Errorf("feishu API error: %d - %s", resp.Code, resp.Msg)
	}

	return nil
}

// buildMessageContent 构建消息内容
func (f *FeishuAdapter) buildMessageContent(message *dto.UnifiedMessage) (string, error) {
	switch message.MessageType {
	case "text":
		return fmt.Sprintf(`{"text":"%s"}`, message.Content), nil
	case "markdown", "md":
		return fmt.Sprintf(`{"content":"%s"}`, message.Content), nil
	case "image":
		return fmt.Sprintf(`{"image_key":"%s"}`, message.MediaURL), nil
	default:
		// 默认作为文本消息处理
		return fmt.Sprintf(`{"text":"%s"}`, message.Content), nil
	}
}

// mapToFeishuMessageType 映射消息类型到飞书格式
func (f *FeishuAdapter) mapToFeishuMessageType(msgType string) string {
	switch msgType {
	case "text":
		return "text"
	case "markdown", "md":
		return "interactive"
	case "image":
		return "image"
	case "audio":
		return "audio"
	case "media":
		return "media"
	case "file":
		return "file"
	case "sticker":
		return "sticker"
	default:
		return "text"
	}
}

// getClient 获取飞书客户端
func (f *FeishuAdapter) getClient(cfg map[string]interface{}) (*lark.Client, error) {
	if f.client != nil {
		return f.client, nil
	}

	appID, ok := cfg["app_id"].(string)
	if !ok || appID == "" {
		return nil, fmt.Errorf("app_id is required in feishu config")
	}
	appSecret, ok := cfg["app_secret"].(string)
	if !ok || appSecret == "" {
		return nil, fmt.Errorf("app_secret is required in feishu config")
	}

	// 创建飞书客户端
	client := lark.NewClient(appID, appSecret, lark.WithLogLevel(larkcore.LogLevelError))

	f.client = client
	return client, nil
}

// VerifyWebhook 实现WebhookProcessor接口
// 飞书使用 challenge 机制进行验证，此处验证 verification_token
func (f *FeishuAdapter) VerifyWebhook(ctx context.Context, signature string, timestamp string, nonce string, body []byte, cfg map[string]interface{}) error {
	// 飞书的 URL 验证通过解析 challenge 完成，签名验证由 SDK 内部处理
	// 这里仅做基础的 token 校验（如配置了 verification_token）
	verificationToken, _ := cfg["verification_token"].(string)
	if verificationToken == "" {
		// 未配置 token 时跳过验证
		return nil
	}

	// 解析请求体检查 token
	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to parse feishu webhook body: %w", err)
	}

	// 飞书回调中 token 字段在顶层
	if token, ok := event["token"].(string); ok && token != verificationToken {
		return fmt.Errorf("feishu webhook token mismatch")
	}

	return nil
}

// ParseInboundMessage 实现WebhookProcessor接口
func (f *FeishuAdapter) ParseInboundMessage(ctx context.Context, body []byte, cfg map[string]interface{}) (*dto.UnifiedMessage, error) {
	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to parse feishu webhook body: %w", err)
	}

	// 处理飞书 URL 验证挑战
	if challenge, ok := event["challenge"].(string); ok {
		return &dto.UnifiedMessage{
			MessageID:   fmt.Sprintf("feishu_challenge_%d", time.Now().UnixNano()),
			MessageType: entity.MessageTypeEvent,
			Content:     challenge,
			RawContent:  map[string]interface{}{"challenge": challenge, "type": "url_verification"},
		}, nil
	}

	// 解析飞书事件消息
	header, _ := event["header"].(map[string]interface{})
	eventData, _ := event["event"].(map[string]interface{})

	eventType := ""
	if header != nil {
		eventType, _ = header["event_type"].(string)
	}

	msg := &dto.UnifiedMessage{
		MessageType:       entity.MessageTypeEvent,
		RawContent:        event,
		PlatformTimestamp: time.Now(),
	}

	if eventID, ok := header["event_id"].(string); ok {
		msg.MessageID = fmt.Sprintf("feishu_%s", eventID)
		msg.PlatformMessageID = eventID
	}

	switch eventType {
	case "im.message.receive_v1":
		if eventData != nil {
			msg.MessageType = entity.MessageTypeText
			if sender, ok := eventData["sender"].(map[string]interface{}); ok {
				msg.SenderID, _ = sender["sender_id"].(string)
				msg.SenderType = entity.SenderTypeUser
			}
			if message, ok := eventData["message"].(map[string]interface{}); ok {
				msg.Content, _ = message["content"].(string)
				if msgType, ok := message["msg_type"].(string); ok {
					msg.MessageType = msgType
				}
				if msgID, ok := message["message_id"].(string); ok {
					msg.PlatformMessageID = msgID
					msg.MessageID = fmt.Sprintf("feishu_%s", msgID)
				}
				if chatID, ok := message["chat_id"].(string); ok {
					msg.ConversationID = chatID
					msg.ReceiverID = chatID
				}
			}
		}
	}

	return msg, nil
}

// BuildWebhookPath 实现WebhookProcessor接口
func (f *FeishuAdapter) BuildWebhookPath(channelID int64) string {
	return fmt.Sprintf("/webhook/feishu/%d", channelID)
}

// 确保 FeishuAdapter 实现了所需的接口
var _ port.MessageSender = (*FeishuAdapter)(nil)
var _ port.WebhookProcessor = (*FeishuAdapter)(nil)
var _ port.PlatformIdentifier = (*FeishuAdapter)(nil)
var _ port.ConfigValidator = (*FeishuAdapter)(nil)
