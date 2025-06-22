package feishu

import (
	"context"
	"fmt"

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

	appID := cfg["app_id"].(string)
	appSecret := cfg["app_secret"].(string)

	// 创建飞书客户端
	client := lark.NewClient(appID, appSecret, lark.WithLogLevel(larkcore.LogLevelError))

	f.client = client
	return client, nil
}

// 确保 FeishuAdapter 实现了所需的接口
var _ port.MessageSender = (*FeishuAdapter)(nil)
var _ port.PlatformIdentifier = (*FeishuAdapter)(nil)
var _ port.ConfigValidator = (*FeishuAdapter)(nil)
