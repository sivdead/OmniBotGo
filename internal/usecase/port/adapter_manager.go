package port

import (
	"context"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
)

// AdapterManager 适配器管理器接口
type AdapterManager interface {
	// GetAdapter 获取指定平台的适配器
	GetAdapter(platformType entity.PlatformType) (interface{}, error)

	// GetMessageSender 获取具备消息发送能力的适配器
	GetMessageSender(platformType entity.PlatformType) (MessageSender, error)

	// GetWebhookProcessor 获取具备Webhook处理能力的适配器
	GetWebhookProcessor(platformType entity.PlatformType) (WebhookProcessor, error)

	// GetTokenManager 获取具备Token管理能力的适配器
	GetTokenManager(platformType entity.PlatformType) (TokenManager, error)

	// GetStreamAdapter 获取具备Stream连接能力的适配器
	GetStreamAdapter(platformType entity.PlatformType) (StreamAdapter, error)

	// GetConfigValidator 获取具备配置验证能力的适配器
	GetConfigValidator(platformType entity.PlatformType) (ConfigValidator, error)

	// GetSupportedPlatforms 获取支持的平台列表
	GetSupportedPlatforms() []entity.PlatformType

	// ValidateConfig 验证平台配置
	ValidateConfig(platformType entity.PlatformType, config map[string]interface{}) error

	// SendMessage 发送消息
	SendMessage(ctx context.Context, platformType entity.PlatformType, message *dto.UnifiedMessage, config map[string]interface{}, accessToken string) error
}
