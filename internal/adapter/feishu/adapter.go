package feishu

import (
	"context"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
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
func (f *FeishuAdapter) ValidateConfig(config map[string]interface{}) error {
	appID, ok := config["app_id"].(string)
	if !ok || appID == "" {
		return fmt.Errorf("app_id is required")
	}

	appSecret, ok := config["app_secret"].(string)
	if !ok || appSecret == "" {
		return fmt.Errorf("app_secret is required")
	}

	return nil
}

// SendMessage 发送消息
func (f *FeishuAdapter) SendMessage(ctx context.Context, message *entity.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	client, err := f.getClient(config)
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
func (f *FeishuAdapter) buildMessageContent(message *entity.UnifiedMessage) (string, error) {
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
func (f *FeishuAdapter) getClient(config map[string]interface{}) (*lark.Client, error) {
	if f.client != nil {
		return f.client, nil
	}

	appID := config["app_id"].(string)
	appSecret := config["app_secret"].(string)

	// 创建飞书客户端
	client := lark.NewClient(appID, appSecret, lark.WithLogLevel(larkcore.LogLevelError))

	f.client = client
	return client, nil
}

// ParseFeishuConfig 解析飞书配置
func ParseFeishuConfig(config map[string]interface{}) (*FeishuConfig, error) {
	appID, ok := config["app_id"].(string)
	if !ok || appID == "" {
		return nil, fmt.Errorf("app_id is required")
	}

	appSecret, ok := config["app_secret"].(string)
	if !ok || appSecret == "" {
		return nil, fmt.Errorf("app_secret is required")
	}

	webhookURL, _ := config["webhook_url"].(string)
	encryptKey, _ := config["encrypt_key"].(string)
	verificationToken, _ := config["verification_token"].(string)

	return &FeishuConfig{
		AppID:             appID,
		AppSecret:         appSecret,
		WebhookURL:        webhookURL,
		EncryptKey:        encryptKey,
		VerificationToken: verificationToken,
	}, nil
}

// FeishuConfig 飞书配置
type FeishuConfig struct {
	AppID             string `json:"app_id"`
	AppSecret         string `json:"app_secret"`
	WebhookURL        string `json:"webhook_url"`
	EncryptKey        string `json:"encrypt_key"`
	VerificationToken string `json:"verification_token"`
}

// 确保 FeishuAdapter 实现了所需的接口
var _ port.MessageSender = (*FeishuAdapter)(nil)
var _ port.PlatformIdentifier = (*FeishuAdapter)(nil)
var _ port.ConfigValidator = (*FeishuAdapter)(nil)
