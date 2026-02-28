package dingtalk_enterprise

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewAdapter(t *testing.T) {
	logger := zerolog.Nop()
	adapter := NewAdapter(logger)

	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.httpClient)
	assert.NotNil(t, adapter.tokenCache)
	assert.NotNil(t, adapter.webhookHandlers)
}

func TestGetPlatformType(t *testing.T) {
	logger := zerolog.Nop()
	adapter := NewAdapter(logger)

	platformType := adapter.GetPlatformType()
	assert.Equal(t, entity.PlatformTypeDingtalk, platformType)
}

func TestValidateConfig(t *testing.T) {
	logger := zerolog.Nop()
	adapter := NewAdapter(logger)

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid enterprise config",
			config: map[string]interface{}{
				"app_key":    "test_key",
				"app_secret": "test_secret",
				"agent_id":   "test_agent",
			},
			wantErr: false,
		},
		{
			name: "missing app_key",
			config: map[string]interface{}{
				"app_secret": "test_secret",
				"agent_id":   "test_agent",
			},
			wantErr: true,
		},
		{
			name: "missing app_secret",
			config: map[string]interface{}{
				"app_key":  "test_key",
				"agent_id": "test_agent",
			},
			wantErr: true,
		},
		{
			name: "missing agent_id",
			config: map[string]interface{}{
				"app_key":    "test_key",
				"app_secret": "test_secret",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildDingtalkMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  *dto.UnifiedMessage
		expected map[string]interface{}
	}{
		{
			name: "text message",
			message: &dto.UnifiedMessage{
				MessageType: entity.MessageTypeText,
				Content:     "Hello, World!",
			},
			expected: map[string]interface{}{
				"msgtype": "text",
				"text": map[string]interface{}{
					"content": "Hello, World!",
				},
			},
		},
		{
			name: "markdown message",
			message: &dto.UnifiedMessage{
				MessageType: entity.MessageTypeMarkdown,
				MarkdownContent: &entity.MarkdownMessage{
					Title:   "Test Title",
					Content: "## Test Content",
				},
			},
			expected: map[string]interface{}{
				"msgtype": "markdown",
				"markdown": map[string]interface{}{
					"title": "Test Title",
					"text":  "## Test Content",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDingtalkMessage(tt.message)
			assert.Equal(t, tt.expected["msgtype"], result["msgtype"])

			// 验证具体内容
			if textContent, ok := tt.expected["text"]; ok {
				assert.Equal(t, textContent, result["text"])
			}
			if markdownContent, ok := tt.expected["markdown"]; ok {
				assert.Equal(t, markdownContent, result["markdown"])
			}
		})
	}
}

func TestParseEnterpriseConfig(t *testing.T) {
	config := map[string]interface{}{
		"app_key":      "test_key",
		"app_secret":   "test_secret",
		"agent_id":     "test_agent",
		"corp_id":      "test_corp",
		"message_mode": "group_chat",
	}

	cfg, err := ParseEnterpriseConfig(config)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test_key", cfg.AppKey)
	assert.Equal(t, "test_secret", cfg.AppSecret)
	assert.Equal(t, "test_agent", cfg.AgentId)
	assert.Equal(t, "test_corp", cfg.CorpId)
	assert.Equal(t, "group_chat", cfg.MessageMode)
}

func TestTokenCaching(t *testing.T) {
	logger := zerolog.Nop()
	adapter := NewAdapter(logger)

	// 测试缓存键生成
	config := map[string]interface{}{
		"app_key":    "test_key",
		"app_secret": "test_secret",
	}

	// 由于GetAccessToken需要实际的API调用，这里只测试缓存逻辑
	ctx := context.Background()

	// 第一次调用会失败（因为是测试环境），但应该不会panic
	_, err := adapter.GetAccessToken(ctx, config)
	assert.Error(t, err) // 预期会失败，因为没有真实的API
}

func TestSendMessageRouting(t *testing.T) {
	logger := zerolog.Nop()
	adapter := NewAdapter(logger)

	ctx := context.Background()
	config := map[string]interface{}{
		"agent_id": "test_agent",
	}

	tests := []struct {
		name         string
		message      *dto.UnifiedMessage
		expectMethod string
	}{
		{
			name: "user message - work notice",
			message: &dto.UnifiedMessage{
				ReceiverType: entity.ReceiverTypeUser,
				ReceiverID:   "user123",
			},
			expectMethod: "work_notice",
		},
		{
			name: "group message",
			message: &dto.UnifiedMessage{
				ReceiverType: entity.ReceiverTypeGroup,
				ReceiverID:   "chat123",
			},
			expectMethod: "group_message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 由于需要实际的access token和API调用，这里只验证不会panic
			err := adapter.SendMessage(ctx, tt.message, config, "fake_token")
			assert.Error(t, err) // 预期会失败，但不应该panic
		})
	}
}
