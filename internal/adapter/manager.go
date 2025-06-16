package adapter

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/adapter/dingtalk_stream"
	"github.com/sivdead/OmniBotGo/internal/adapter/feishu"
	"github.com/sivdead/OmniBotGo/internal/adapter/wecom"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

// Manager 平台适配器管理器
type Manager struct {
	adapters map[entity.PlatformType]interface{}
}

// NewManager 创建适配器管理器实例
func NewManager() *Manager {
	adapters := make(map[entity.PlatformType]interface{})

	// 注册企业微信适配器
	adapters[entity.PlatformTypeWecom] = wecom.NewWecomAdapter()

	// 注册钉钉Stream适配器
	adapters[entity.PlatformTypeDingtalk] = dingtalk_stream.NewAdapter(zerolog.New(os.Stdout))

	// 注册飞书适配器
	adapters[entity.PlatformTypeFeishu] = feishu.NewFeishuAdapter()

	// 可以继续注册其他平台适配器

	return &Manager{
		adapters: adapters,
	}
}

// GetAdapter 获取指定平台的适配器（原始接口）
func (m *Manager) GetAdapter(platformType entity.PlatformType) (interface{}, error) {
	adapter, exists := m.adapters[platformType]
	if !exists {
		return nil, fmt.Errorf("unsupported platform type: %s", platformType)
	}
	return adapter, nil
}

// GetMessageSender 获取具备消息发送能力的适配器
func (m *Manager) GetMessageSender(platformType entity.PlatformType) (port.MessageSender, error) {
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return nil, err
	}

	if messageSender, ok := adapter.(port.MessageSender); ok {
		return messageSender, nil
	}

	return nil, fmt.Errorf("platform %s does not support message sending", platformType)
}

// GetWebhookProcessor 获取具备Webhook处理能力的适配器
func (m *Manager) GetWebhookProcessor(platformType entity.PlatformType) (port.WebhookProcessor, error) {
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return nil, err
	}

	if webhookProcessor, ok := adapter.(port.WebhookProcessor); ok {
		return webhookProcessor, nil
	}

	return nil, fmt.Errorf("platform %s does not support webhook processing", platformType)
}

// GetTokenManager 获取具备Token管理能力的适配器
func (m *Manager) GetTokenManager(platformType entity.PlatformType) (port.TokenManager, error) {
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return nil, err
	}

	if tokenManager, ok := adapter.(port.TokenManager); ok {
		return tokenManager, nil
	}

	return nil, fmt.Errorf("platform %s does not support token management", platformType)
}

// GetStreamAdapter 获取具备Stream连接能力的适配器
func (m *Manager) GetStreamAdapter(platformType entity.PlatformType) (port.StreamAdapter, error) {
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return nil, err
	}

	if streamAdapter, ok := adapter.(port.StreamAdapter); ok {
		return streamAdapter, nil
	}

	return nil, fmt.Errorf("platform %s does not support stream connections", platformType)
}

// GetConfigValidator 获取具备配置验证能力的适配器
func (m *Manager) GetConfigValidator(platformType entity.PlatformType) (port.ConfigValidator, error) {
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return nil, err
	}

	if configValidator, ok := adapter.(port.ConfigValidator); ok {
		return configValidator, nil
	}

	return nil, fmt.Errorf("platform %s does not support config validation", platformType)
}

// GetSupportedPlatforms 获取支持的平台列表
func (m *Manager) GetSupportedPlatforms() []entity.PlatformType {
	platforms := make([]entity.PlatformType, 0, len(m.adapters))
	for platformType := range m.adapters {
		platforms = append(platforms, platformType)
	}
	return platforms
}

// ValidateConfig 验证平台配置
func (m *Manager) ValidateConfig(platformType entity.PlatformType, config map[string]interface{}) error {
	validator, err := m.GetConfigValidator(platformType)
	if err != nil {
		return err
	}
	return validator.ValidateConfig(config)
}

// SendMessage 发送消息
func (m *Manager) SendMessage(ctx context.Context, platformType entity.PlatformType, message *entity.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	sender, err := m.GetMessageSender(platformType)
	if err != nil {
		return err
	}
	return sender.SendMessage(ctx, message, config, accessToken)
}

// 以下是为了保持向后兼容性的方法，逐步迁移后可以删除

// GetAccessToken 获取访问令牌（兼容性方法）
func (m *Manager) GetAccessToken(ctx context.Context, platformType entity.PlatformType, config map[string]interface{}) (*entity.AccessTokenResponse, error) {
	tokenManager, err := m.GetTokenManager(platformType)
	if err != nil {
		return nil, err
	}
	return tokenManager.GetAccessToken(ctx, config)
}

// RefreshAccessToken 刷新访问令牌（兼容性方法）
func (m *Manager) RefreshAccessToken(ctx context.Context, platformType entity.PlatformType, config map[string]interface{}, oldToken string) (*entity.AccessTokenResponse, error) {
	tokenManager, err := m.GetTokenManager(platformType)
	if err != nil {
		return nil, err
	}
	return tokenManager.RefreshAccessToken(ctx, config, oldToken)
}

// VerifyWebhook 验证Webhook请求（兼容性方法）
func (m *Manager) VerifyWebhook(ctx context.Context, platformType entity.PlatformType, signature string, timestamp string, nonce string, body []byte, config map[string]interface{}) error {
	processor, err := m.GetWebhookProcessor(platformType)
	if err != nil {
		return err
	}
	return processor.VerifyWebhook(ctx, signature, timestamp, nonce, body, config)
}

// ParseInboundMessage 解析入站消息（兼容性方法）
func (m *Manager) ParseInboundMessage(ctx context.Context, platformType entity.PlatformType, body []byte, config map[string]interface{}) (*entity.UnifiedMessage, error) {
	processor, err := m.GetWebhookProcessor(platformType)
	if err != nil {
		return nil, err
	}
	return processor.ParseInboundMessage(ctx, body, config)
}

// BuildWebhookPath 构建Webhook路径（兼容性方法）
func (m *Manager) BuildWebhookPath(platformType entity.PlatformType, channelID int64) (string, error) {
	processor, err := m.GetWebhookProcessor(platformType)
	if err != nil {
		return "", err
	}
	return processor.BuildWebhookPath(channelID), nil
}
