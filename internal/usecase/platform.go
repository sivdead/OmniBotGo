package usecase

import (
	"context"
	"fmt"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// platformUC 平台管理用例实现
type platformUC struct {
	adapterManager port.AdapterManager
	channelRepo    port.ChannelRepository
	logger         logger.Interface
}

// NewPlatformUseCase 创建平台管理用例
func NewPlatformUseCase(
	adapterManager port.AdapterManager,
	channelRepo port.ChannelRepository,
	logger logger.Interface,
) PlatformUseCase {
	return &platformUC{
		adapterManager: adapterManager,
		channelRepo:    channelRepo,
		logger:         logger,
	}
}

// GetPlatforms 获取支持的平台列表
func (uc *platformUC) GetPlatforms(ctx context.Context) ([]*PlatformInfo, error) {
	platforms := []*PlatformInfo{
		{
			Type:        string(entity.PlatformTypeWecom),
			Name:        "企业微信",
			Description: "企业微信机器人平台",
			SupportedFeatures: []string{
				"text_message",
				"image_message",
				"file_message",
				"markdown_message",
				"card_message",
				"webhook",
			},
			ConfigFields: []PlatformConfigField{
				{Field: "corp_id", Label: "企业ID", Type: "string", Required: true},
				{Field: "agent_id", Label: "应用ID", Type: "string", Required: true},
				{Field: "secret", Label: "应用密钥", Type: "password", Required: true},
				{Field: "token", Label: "接收消息Token", Type: "string", Required: false},
				{Field: "encoding_aes_key", Label: "接收消息EncodingAESKey", Type: "string", Required: false},
			},
			WebhookConfig: PlatformWebhookConfig{
				URLPattern:  "/api/v1/webhook/wecom/{channel_id}",
				Method:      "POST",
				ContentType: "application/json",
			},
			Status:  "active",
			IconURL: "/static/icons/wecom.png",
		},
		{
			Type:        string(entity.PlatformTypeDingTalk),
			Name:        "钉钉",
			Description: "钉钉机器人平台",
			SupportedFeatures: []string{
				"text_message",
				"markdown_message",
				"link_message",
				"action_card_message",
				"webhook",
			},
			ConfigFields: []PlatformConfigField{
				{Field: "app_key", Label: "应用Key", Type: "string", Required: true},
				{Field: "app_secret", Label: "应用密钥", Type: "password", Required: true},
				{Field: "robot_code", Label: "机器人编码", Type: "string", Required: false},
			},
			WebhookConfig: PlatformWebhookConfig{
				URLPattern:  "/api/v1/webhook/dingtalk/{channel_id}",
				Method:      "POST",
				ContentType: "application/json",
			},
			Status:  "active",
			IconURL: "/static/icons/dingtalk.png",
		},
		{
			Type:        string(entity.PlatformTypeDingTalkStream),
			Name:        "钉钉Stream",
			Description: "钉钉Stream连接模式",
			SupportedFeatures: []string{
				"text_message",
				"markdown_message",
				"interactive_card",
				"stream_connection",
				"real_time_message",
			},
			ConfigFields: []PlatformConfigField{
				{Field: "client_id", Label: "客户端ID", Type: "string", Required: true},
				{Field: "client_secret", Label: "客户端密钥", Type: "password", Required: true},
			},
			WebhookConfig: PlatformWebhookConfig{
				URLPattern:  "", // Stream模式不使用webhook
				Method:      "",
				ContentType: "",
			},
			Status:  "active",
			IconURL: "/static/icons/dingtalk.png",
		},
		{
			Type:        string(entity.PlatformTypeDingTalkEnterprise),
			Name:        "钉钉企业应用",
			Description: "钉钉企业内部应用",
			SupportedFeatures: []string{
				"text_message",
				"markdown_message",
				"oa_message",
				"file_message",
				"webhook",
			},
			ConfigFields: []PlatformConfigField{
				{Field: "app_key", Label: "应用Key", Type: "string", Required: true},
				{Field: "app_secret", Label: "应用密钥", Type: "password", Required: true},
				{Field: "agent_id", Label: "AgentID", Type: "string", Required: true},
			},
			WebhookConfig: PlatformWebhookConfig{
				URLPattern:  "/api/v1/webhook/dingtalk-enterprise/{channel_id}",
				Method:      "POST",
				ContentType: "application/json",
			},
			Status:  "active",
			IconURL: "/static/icons/dingtalk.png",
		},
		{
			Type:        string(entity.PlatformTypeFeishu),
			Name:        "飞书",
			Description: "飞书机器人平台",
			SupportedFeatures: []string{
				"text_message",
				"rich_text_message",
				"image_message",
				"interactive_card",
				"webhook",
			},
			ConfigFields: []PlatformConfigField{
				{Field: "app_id", Label: "应用ID", Type: "string", Required: true},
				{Field: "app_secret", Label: "应用密钥", Type: "password", Required: true},
				{Field: "verification_token", Label: "验证Token", Type: "string", Required: false},
				{Field: "encrypt_key", Label: "加密Key", Type: "string", Required: false},
			},
			WebhookConfig: PlatformWebhookConfig{
				URLPattern:  "/api/v1/webhook/feishu/{channel_id}",
				Method:      "POST",
				ContentType: "application/json",
			},
			Status:  "active",
			IconURL: "/static/icons/feishu.png",
		},
		{
			Type:        string(entity.PlatformTypeWechatOfficial),
			Name:        "微信公众号",
			Description: "微信公众号平台",
			SupportedFeatures: []string{
				"text_message",
				"image_message",
				"voice_message",
				"video_message",
				"news_message",
				"template_message",
				"webhook",
			},
			ConfigFields: []PlatformConfigField{
				{Field: "app_id", Label: "AppID", Type: "string", Required: true},
				{Field: "app_secret", Label: "AppSecret", Type: "password", Required: true},
				{Field: "token", Label: "Token", Type: "string", Required: true},
				{Field: "encoding_aes_key", Label: "EncodingAESKey", Type: "string", Required: false},
			},
			WebhookConfig: PlatformWebhookConfig{
				URLPattern:  "/api/v1/webhook/wechat/{channel_id}",
				Method:      "POST",
				ContentType: "application/xml",
			},
			Status:  "active",
			IconURL: "/static/icons/wechat.png",
		},
	}

	return platforms, nil
}

// GetPlatformByType 获取平台详情
func (uc *platformUC) GetPlatformByType(ctx context.Context, platformType string) (*PlatformInfo, error) {
	platforms, err := uc.GetPlatforms(ctx)
	if err != nil {
		return nil, err
	}

	for _, platform := range platforms {
		if platform.Type == platformType {
			return platform, nil
		}
	}

	return nil, fmt.Errorf("platform type %s not found", platformType)
}

// ValidatePlatformConfig 验证平台配置
func (uc *platformUC) ValidatePlatformConfig(ctx context.Context, req ValidatePlatformConfigRequest) (*PlatformConfigValidationResult, error) {
	// 获取平台信息
	platform, err := uc.GetPlatformByType(ctx, req.PlatformType)
	if err != nil {
		return &PlatformConfigValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("invalid platform type: %s", req.PlatformType)},
		}, nil
	}

	errors := []string{}

	// 验证必填字段
	for _, field := range platform.ConfigFields {
		if field.Required {
			value, exists := req.Config[field.Field]
			if !exists || value == nil || value == "" {
				errors = append(errors, fmt.Sprintf("required field '%s' is missing", field.Field))
			}
		}
	}

	// 验证字段类型
	for key, value := range req.Config {
		// 查找字段定义
		var fieldDef *PlatformConfigField
		for _, field := range platform.ConfigFields {
			if field.Field == key {
				fieldDef = &field
				break
			}
		}

		if fieldDef == nil {
			errors = append(errors, fmt.Sprintf("unknown field '%s'", key))
			continue
		}

		// 基本类型验证
		switch fieldDef.Type {
		case "string", "password":
			if _, ok := value.(string); !ok {
				errors = append(errors, fmt.Sprintf("field '%s' must be a string", key))
			}
		case "number":
			switch value.(type) {
			case int, int32, int64, float32, float64:
				// valid number
			default:
				errors = append(errors, fmt.Sprintf("field '%s' must be a number", key))
			}
		case "boolean":
			if _, ok := value.(bool); !ok {
				errors = append(errors, fmt.Sprintf("field '%s' must be a boolean", key))
			}
		}
	}

	// 平台特定的验证
	switch req.PlatformType {
	case string(entity.PlatformTypeWecom):
		// 验证企业微信特定配置
		if corpID, ok := req.Config["corp_id"].(string); ok && len(corpID) < 10 {
			errors = append(errors, "invalid corp_id format")
		}
	case string(entity.PlatformTypeDingTalk), string(entity.PlatformTypeDingTalkEnterprise):
		// 验证钉钉特定配置
		if appKey, ok := req.Config["app_key"].(string); ok && len(appKey) < 10 {
			errors = append(errors, "invalid app_key format")
		}
	case string(entity.PlatformTypeFeishu):
		// 验证飞书特定配置
		if appID, ok := req.Config["app_id"].(string); ok && len(appID) < 10 {
			errors = append(errors, "invalid app_id format")
		}
	}

	result := &PlatformConfigValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}

	return result, nil
}

// GetPlatformStatus 获取平台状态
func (uc *platformUC) GetPlatformStatus(ctx context.Context, platformType string) (*PlatformStatusResult, error) {
	// 验证平台类型
	_, err := uc.GetPlatformByType(ctx, platformType)
	if err != nil {
		return nil, err
	}

	// 获取该平台的所有活跃通道
	channels, err := uc.channelRepo.GetByPlatformType(ctx, platformType)
	if err != nil {
		uc.logger.Error("failed to get channels by platform type", "error", err, "platform", platformType)
		return &PlatformStatusResult{
			PlatformType: platformType,
			Status:       "error",
			Message:      "Failed to check platform status",
		}, nil
	}

	// 检查是否有连接的通道
	connectedCount := 0
	for _, channel := range channels {
		if channel.ConnectionStatus == entity.ConnectionStatusConnected {
			connectedCount++
		}
	}

	status := "inactive"
	message := "No active channels"

	if len(channels) > 0 {
		if connectedCount > 0 {
			status = "active"
			message = fmt.Sprintf("%d/%d channels connected", connectedCount, len(channels))
		} else {
			status = "disconnected"
			message = fmt.Sprintf("All %d channels disconnected", len(channels))
		}
	}

	return &PlatformStatusResult{
		PlatformType: platformType,
		Status:       status,
		Message:      message,
	}, nil
}
