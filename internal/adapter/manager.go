package adapter

import (
	"context"
	"fmt"

	"github.com/sivdead/OmniBotGo/internal/adapter/dingtalk"
	"github.com/sivdead/OmniBotGo/internal/adapter/feishu"
	"github.com/sivdead/OmniBotGo/internal/adapter/wecom"
	"github.com/sivdead/OmniBotGo/internal/entity"
)

// Manager 平台适配器管理器
type Manager struct {
	adapters map[entity.PlatformType]entity.PlatformAdapter
}

// NewManager 创建适配器管理器实例
func NewManager() *Manager {
	adapters := make(map[entity.PlatformType]entity.PlatformAdapter)

	// 注册企业微信适配器
	adapters[entity.PlatformTypeWecom] = wecom.NewWecomAdapter()

	// 注册钉钉适配器
	adapters[entity.PlatformTypeDingtalk] = dingtalk.NewDingtalkAdapter()

	// 注册飞书适配器
	adapters[entity.PlatformTypeFeishu] = feishu.NewFeishuAdapter()

	// 可以继续注册其他平台适配器
	// adapters[entity.PlatformTypeWechatOfficial] = wechat.NewWechatAdapter()

	return &Manager{
		adapters: adapters,
	}
}

// GetAdapter 获取指定平台的适配器
func (m *Manager) GetAdapter(platformType entity.PlatformType) (entity.PlatformAdapter, error) {
	adapter, exists := m.adapters[platformType]
	if !exists {
		return nil, fmt.Errorf("unsupported platform type: %s", platformType)
	}
	return adapter, nil
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
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return err
	}
	return adapter.ValidateConfig(config)
}

// GetAccessToken 获取访问令牌
func (m *Manager) GetAccessToken(ctx context.Context, platformType entity.PlatformType, config map[string]interface{}) (*entity.AccessTokenResponse, error) {
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAccessToken(ctx, config)
}

// RefreshAccessToken 刷新访问令牌
func (m *Manager) RefreshAccessToken(ctx context.Context, platformType entity.PlatformType, config map[string]interface{}, oldToken string) (*entity.AccessTokenResponse, error) {
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return nil, err
	}
	return adapter.RefreshAccessToken(ctx, config, oldToken)
}

// VerifyWebhook 验证Webhook请求
func (m *Manager) VerifyWebhook(ctx context.Context, platformType entity.PlatformType, signature string, timestamp string, nonce string, body []byte, config map[string]interface{}) error {
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return err
	}
	return adapter.VerifyWebhook(ctx, signature, timestamp, nonce, body, config)
}

// ParseInboundMessage 解析入站消息
func (m *Manager) ParseInboundMessage(ctx context.Context, platformType entity.PlatformType, body []byte, config map[string]interface{}) (*entity.UnifiedMessage, error) {
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return nil, err
	}
	return adapter.ParseInboundMessage(ctx, body, config)
}

// SendMessage 发送消息
func (m *Manager) SendMessage(ctx context.Context, platformType entity.PlatformType, message *entity.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return err
	}
	return adapter.SendMessage(ctx, message, config, accessToken)
}

// BuildWebhookPath 构建Webhook路径
func (m *Manager) BuildWebhookPath(platformType entity.PlatformType, channelID int64) (string, error) {
	adapter, err := m.GetAdapter(platformType)
	if err != nil {
		return "", err
	}
	return adapter.BuildWebhookPath(channelID), nil
}
