package port

import (
	"context"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

// MessageSender 发送消息的能力接口
type MessageSender interface {
	// SendMessage 发送消息
	SendMessage(ctx context.Context, message *entity.UnifiedMessage, config map[string]interface{}, accessToken string) error
}

// WebhookProcessor Webhook处理的能力接口
type WebhookProcessor interface {
	// VerifyWebhook 验证Webhook请求
	VerifyWebhook(ctx context.Context, signature string, timestamp string, nonce string, body []byte, config map[string]interface{}) error
	// ParseInboundMessage 解析入站消息
	ParseInboundMessage(ctx context.Context, body []byte, config map[string]interface{}) (*entity.UnifiedMessage, error)
	// BuildWebhookPath 构建Webhook路径
	BuildWebhookPath(channelID int64) string
}

// TokenManager Token管理的能力接口
type TokenManager interface {
	// GetAccessToken 获取访问令牌
	GetAccessToken(ctx context.Context, config map[string]interface{}) (*entity.AccessTokenResponse, error)
	// RefreshAccessToken 刷新访问令牌
	RefreshAccessToken(ctx context.Context, config map[string]interface{}, oldToken string) (*entity.AccessTokenResponse, error)
}

// StreamAdapter Stream连接适配器接口（用于钉钉Stream等主动连接模式）
type StreamAdapter interface {
	// Start 启动Stream连接
	Start(ctx context.Context, messageHandler MessageHandler, config map[string]interface{}) error
	// Stop 停止Stream连接
	Stop(ctx context.Context) error
	// IsConnected 检查连接状态
	IsConnected() bool
}

// MessageHandler Stream消息处理回调函数
type MessageHandler func(ctx context.Context, message *entity.UnifiedMessage) error

// ConfigValidator 配置验证能力接口
type ConfigValidator interface {
	// ValidateConfig 验证平台配置
	ValidateConfig(config map[string]interface{}) error
}

// PlatformIdentifier 平台识别能力接口
type PlatformIdentifier interface {
	// GetPlatformType 获取平台类型
	GetPlatformType() entity.PlatformType
}
