package adapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

var errAIModelNotImplemented = errors.New("AI model management is not implemented yet")

// Manager 平台适配器管理器
type Manager struct {
	adapters map[entity.PlatformType]interface{}
}

// NewManager 创建适配器管理器实例（兼容旧代码，建议使用NewManagerWithRegistry）
func NewManager() *Manager {
	return &Manager{
		adapters: make(map[entity.PlatformType]interface{}),
	}
}

// NewManagerWithRegistry 使用适配器注册表创建管理器
func NewManagerWithRegistry(registry map[entity.PlatformType]interface{}) *Manager {
	return &Manager{
		adapters: registry,
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
func (m *Manager) SendMessage(ctx context.Context, platformType entity.PlatformType, message *dto.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	sender, err := m.GetMessageSender(platformType)
	if err != nil {
		return err
	}
	return sender.SendMessage(ctx, message, config, accessToken)
}

// VerifyWebhook 验证Webhook签名
func (m *Manager) VerifyWebhook(ctx context.Context, platformType entity.PlatformType, signature, timestamp, nonce string, body []byte, config entity.JSONField) error {
	processor, err := m.GetWebhookProcessor(platformType)
	if err != nil {
		return err
	}

	return processor.VerifyWebhook(ctx, signature, timestamp, nonce, body, map[string]interface{}(config))
}

// ParseInboundMessage 解析入站消息
func (m *Manager) ParseInboundMessage(ctx context.Context, platformType entity.PlatformType, body []byte, config entity.JSONField) (*dto.UnifiedMessage, error) {
	processor, err := m.GetWebhookProcessor(platformType)
	if err != nil {
		return nil, err
	}

	return processor.ParseInboundMessage(ctx, body, map[string]interface{}(config))
}

// BuildWebhookPath 构建Webhook路径
func (m *Manager) BuildWebhookPath(platformType entity.PlatformType, channelID int64) (string, error) {
	return fmt.Sprintf("/webhook/%s/%d", string(platformType), channelID), nil
}

// GetAIModelManager 获取AI模型管理器
func (m *Manager) GetAIModelManager() port.AIModelManager {
	return &notImplementedAIModelManager{}
}

// notImplementedAIModelManager 占位实现，所有方法返回"未实现"错误。
type notImplementedAIModelManager struct{}

func (n *notImplementedAIModelManager) CreateChatModel(_ context.Context, _ string, _ map[string]interface{}) (interface{}, error) {
	return nil, errAIModelNotImplemented
}

func (n *notImplementedAIModelManager) GetSupportedProviders() []string {
	return nil
}

func (n *notImplementedAIModelManager) ValidateConfig(_ string, _ map[string]interface{}) error {
	return errAIModelNotImplemented
}
