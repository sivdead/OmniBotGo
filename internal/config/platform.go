// Package config 定义应用程序的配置结构
package config

import (
	"fmt"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

// PlatformConfig 平台配置基础结构
type PlatformConfig struct {
	Platform entity.PlatformType    `json:"platform"`
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

// DingtalkStreamConfig 钉钉Stream模式配置
type DingtalkStreamConfig struct {
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
}

// ValidatePlatformConfig 验证平台配置
func ValidatePlatformConfig(platformType entity.PlatformType, config map[string]interface{}) error {
	switch platformType {
	case entity.PlatformTypeWecom:
		return validateWecomConfig(config)
	case entity.PlatformTypeDingtalk:
		return validateDingtalkConfig(config)
	case entity.PlatformTypeWechatOfficial:
		return validateWechatOfficialConfig(config)
	case entity.PlatformTypeFeishu:
		return validateFeishuConfig(config)
	default:
		return fmt.Errorf("unknown platform type: %s", platformType)
	}
}

// validateWecomConfig 验证企业微信配置
func validateWecomConfig(config map[string]interface{}) error {
	requiredFields := []string{"corp_id", "agent_id", "secret"}
	for _, field := range requiredFields {
		if val, ok := config[field].(string); !ok || val == "" {
			return fmt.Errorf("%s is required for wecom platform", field)
		}
	}
	return nil
}

// validateDingtalkConfig 验证钉钉配置
func validateDingtalkConfig(config map[string]interface{}) error {
	// Stream模式只需要client_id和client_secret
	if clientID, hasClientID := config["client_id"].(string); hasClientID && clientID != "" {
		if clientSecret, hasClientSecret := config["client_secret"].(string); !hasClientSecret || clientSecret == "" {
			return fmt.Errorf("client_secret is required for dingtalk stream mode")
		}
		return nil
	}

	// 普通模式需要app_key, app_secret, agent_id
	requiredFields := []string{"app_key", "app_secret", "agent_id"}
	for _, field := range requiredFields {
		if val, ok := config[field].(string); !ok || val == "" {
			return fmt.Errorf("%s is required for dingtalk platform", field)
		}
	}
	return nil
}

// validateWechatOfficialConfig 验证微信公众号配置
func validateWechatOfficialConfig(config map[string]interface{}) error {
	requiredFields := []string{"app_id", "app_secret", "token"}
	for _, field := range requiredFields {
		if val, ok := config[field].(string); !ok || val == "" {
			return fmt.Errorf("%s is required for wechat official platform", field)
		}
	}
	return nil
}

// validateFeishuConfig 验证飞书配置
func validateFeishuConfig(config map[string]interface{}) error {
	requiredFields := []string{"app_id", "app_secret"}
	for _, field := range requiredFields {
		if val, ok := config[field].(string); !ok || val == "" {
			return fmt.Errorf("%s is required for feishu platform", field)
		}
	}
	return nil
}

// ParseWecomConfig 解析企业微信配置
func ParseWecomConfig(config map[string]interface{}) (*WecomConfig, error) {
	if err := validateWecomConfig(config); err != nil {
		return nil, err
	}

	return &WecomConfig{
		CorpID:     config["corp_id"].(string),
		AgentID:    config["agent_id"].(string),
		Secret:     config["secret"].(string),
		Token:      getStringOrDefault(config, "token", ""),
		AESKey:     getStringOrDefault(config, "aes_key", ""),
		WebhookURL: getStringOrDefault(config, "webhook_url", ""),
	}, nil
}

// ParseDingtalkConfig 解析钉钉配置
func ParseDingtalkConfig(config map[string]interface{}) (*DingtalkConfig, error) {
	if err := validateDingtalkConfig(config); err != nil {
		return nil, err
	}

	return &DingtalkConfig{
		AppKey:     getStringOrDefault(config, "app_key", ""),
		AppSecret:  getStringOrDefault(config, "app_secret", ""),
		AgentID:    getStringOrDefault(config, "agent_id", ""),
		Token:      getStringOrDefault(config, "token", ""),
		AESKey:     getStringOrDefault(config, "aes_key", ""),
		WebhookURL: getStringOrDefault(config, "webhook_url", ""),
	}, nil
}

// ParseDingtalkStreamConfig 解析钉钉Stream模式配置
func ParseDingtalkStreamConfig(config map[string]interface{}) (*DingtalkStreamConfig, error) {
	clientID, ok := config["client_id"].(string)
	if !ok || clientID == "" {
		return nil, fmt.Errorf("client_id is required for dingtalk stream mode")
	}

	clientSecret, ok := config["client_secret"].(string)
	if !ok || clientSecret == "" {
		return nil, fmt.Errorf("client_secret is required for dingtalk stream mode")
	}

	return &DingtalkStreamConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

// ParseWechatOfficialConfig 解析微信公众号配置
func ParseWechatOfficialConfig(config map[string]interface{}) (*WechatOfficialConfig, error) {
	if err := validateWechatOfficialConfig(config); err != nil {
		return nil, err
	}

	return &WechatOfficialConfig{
		AppID:      config["app_id"].(string),
		AppSecret:  config["app_secret"].(string),
		Token:      config["token"].(string),
		AESKey:     getStringOrDefault(config, "aes_key", ""),
		WebhookURL: getStringOrDefault(config, "webhook_url", ""),
	}, nil
}

// ParseFeishuConfig 解析飞书配置
func ParseFeishuConfig(config map[string]interface{}) (*FeishuConfig, error) {
	if err := validateFeishuConfig(config); err != nil {
		return nil, err
	}

	return &FeishuConfig{
		AppID:      config["app_id"].(string),
		AppSecret:  config["app_secret"].(string),
		Token:      getStringOrDefault(config, "token", ""),
		AESKey:     getStringOrDefault(config, "aes_key", ""),
		WebhookURL: getStringOrDefault(config, "webhook_url", ""),
	}, nil
}

// getStringOrDefault 获取字符串配置值或返回默认值
func getStringOrDefault(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}
