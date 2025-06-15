package entity

import (
	"context"
	"time"
)

// PlatformType 平台类型枚举
type PlatformType string

const (
	PlatformTypeWecom          PlatformType = "wecom"           // 企业微信
	PlatformTypeDingtalk       PlatformType = "dingtalk"        // 钉钉
	PlatformTypeWechatOfficial PlatformType = "wechat_official" // 微信公众号
	PlatformTypeFeishu         PlatformType = "feishu"          // 飞书
)

// PlatformAdapter 平台适配器接口
type PlatformAdapter interface {
	// GetPlatformType 获取平台类型
	GetPlatformType() PlatformType

	// ValidateConfig 验证平台配置
	ValidateConfig(config map[string]interface{}) error

	// GetAccessToken 获取访问令牌
	GetAccessToken(ctx context.Context, config map[string]interface{}) (*AccessTokenResponse, error)

	// RefreshAccessToken 刷新访问令牌
	RefreshAccessToken(ctx context.Context, config map[string]interface{}, oldToken string) (*AccessTokenResponse, error)

	// VerifyWebhook 验证Webhook请求
	VerifyWebhook(ctx context.Context, signature string, timestamp string, nonce string, body []byte, config map[string]interface{}) error

	// ParseInboundMessage 解析入站消息
	ParseInboundMessage(ctx context.Context, body []byte, config map[string]interface{}) (*UnifiedMessage, error)

	// SendMessage 发送消息
	SendMessage(ctx context.Context, message *UnifiedMessage, config map[string]interface{}, accessToken string) error

	// BuildWebhookPath 构建Webhook路径
	BuildWebhookPath(channelID int64) string
}

// AccessTokenResponse 访问令牌响应
type AccessTokenResponse struct {
	AccessToken string     `json:"access_token"`
	ExpiresIn   int        `json:"expires_in"` // 过期时间，秒
	ExpiresAt   *time.Time `json:"expires_at"` // 过期时间点
}

// UnifiedMessage 统一消息格式
type UnifiedMessage struct {
	// 基本信息
	MessageID   string `json:"message_id"`
	MessageType string `json:"message_type"` // text, image, audio, video, file, event

	// 发送者信息
	SenderID   string `json:"sender_id"`
	SenderName string `json:"sender_name"`
	SenderType string `json:"sender_type"` // user, group, system

	// 接收者信息
	ReceiverID   string `json:"receiver_id"`
	ReceiverName string `json:"receiver_name"`
	ReceiverType string `json:"receiver_type"` // user, group, system

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
}

// PlatformConfig 平台配置基础结构
type PlatformConfig struct {
	Platform PlatformType           `json:"platform"`
	Config   map[string]interface{} `json:"config"`
}

// WecomConfig 企业微信平台配置
type WecomConfig struct {
	CorpID     string `json:"corp_id" validate:"required"`
	AgentID    string `json:"agent_id" validate:"required"`
	Secret     string `json:"secret" validate:"required"`
	Token      string `json:"token,omitempty"`       // 用于验证URL
	AESKey     string `json:"aes_key,omitempty"`     // 用于消息加解密
	WebhookURL string `json:"webhook_url,omitempty"` // 接收消息的URL
}

// DingtalkConfig 钉钉平台配置
type DingtalkConfig struct {
	AppKey     string `json:"app_key" validate:"required"`
	AppSecret  string `json:"app_secret" validate:"required"`
	AgentID    string `json:"agent_id" validate:"required"`
	Token      string `json:"token,omitempty"`
	AESKey     string `json:"aes_key,omitempty"`
	WebhookURL string `json:"webhook_url,omitempty"`
}

// WechatOfficialConfig 微信公众号配置
type WechatOfficialConfig struct {
	AppID      string `json:"app_id" validate:"required"`
	AppSecret  string `json:"app_secret" validate:"required"`
	Token      string `json:"token,omitempty"`
	AESKey     string `json:"aes_key,omitempty"`
	WebhookURL string `json:"webhook_url,omitempty"`
}

// FeishuConfig 飞书配置
type FeishuConfig struct {
	AppID      string `json:"app_id" validate:"required"`
	AppSecret  string `json:"app_secret" validate:"required"`
	Token      string `json:"token,omitempty"`
	AESKey     string `json:"aes_key,omitempty"`
	WebhookURL string `json:"webhook_url,omitempty"`
}

// WebhookEvent Webhook事件基础结构
type WebhookEvent struct {
	EventType string                 `json:"event_type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}
