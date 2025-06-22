package adapter

import (
	"context"
	"testing"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAdapter 模拟适配器，实现多个接口
type MockAdapter struct {
	mock.Mock
}

func (m *MockAdapter) SendMessage(ctx context.Context, message *dto.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	args := m.Called(ctx, message, config, accessToken)
	return args.Error(0)
}

func (m *MockAdapter) VerifyWebhook(ctx context.Context, signature string, timestamp string, nonce string, body []byte, config map[string]interface{}) error {
	args := m.Called(ctx, signature, timestamp, nonce, body, config)
	return args.Error(0)
}

func (m *MockAdapter) ParseInboundMessage(ctx context.Context, body []byte, config map[string]interface{}) (*dto.UnifiedMessage, error) {
	args := m.Called(ctx, body, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UnifiedMessage), args.Error(1)
}

func (m *MockAdapter) BuildWebhookPath(channelID int64) string {
	args := m.Called(channelID)
	return args.String(0)
}

func (m *MockAdapter) GetAccessToken(ctx context.Context, config map[string]interface{}) (*dto.AccessTokenResponse, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AccessTokenResponse), args.Error(1)
}

func (m *MockAdapter) RefreshAccessToken(ctx context.Context, config map[string]interface{}, oldToken string) (*dto.AccessTokenResponse, error) {
	args := m.Called(ctx, config, oldToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AccessTokenResponse), args.Error(1)
}

func (m *MockAdapter) ValidateConfig(config map[string]interface{}) error {
	args := m.Called(config)
	return args.Error(0)
}

func TestNewManagerWithRegistry(t *testing.T) {
	// 创建模拟适配器
	mockWecom := &MockAdapter{}
	mockDingtalk := &MockAdapter{}

	// 创建注册表
	registry := map[entity.PlatformType]interface{}{
		entity.PlatformTypeWecom:    mockWecom,
		entity.PlatformTypeDingtalk: mockDingtalk,
	}

	// 创建管理器
	manager := NewManagerWithRegistry(registry)

	// 验证管理器不为空
	assert.NotNil(t, manager)

	// 验证可以获取适配器
	adapter, err := manager.GetAdapter(entity.PlatformTypeWecom)
	assert.NoError(t, err)
	assert.Equal(t, mockWecom, adapter)
}

func TestManager_GetAdapter(t *testing.T) {
	mockAdapter := &MockAdapter{}
	registry := map[entity.PlatformType]interface{}{
		entity.PlatformTypeWecom: mockAdapter,
	}
	manager := NewManagerWithRegistry(registry)

	t.Run("获取存在的适配器", func(t *testing.T) {
		adapter, err := manager.GetAdapter(entity.PlatformTypeWecom)
		assert.NoError(t, err)
		assert.Equal(t, mockAdapter, adapter)
	})

	t.Run("获取不存在的适配器", func(t *testing.T) {
		adapter, err := manager.GetAdapter(entity.PlatformTypeDingtalk)
		assert.Error(t, err)
		assert.Nil(t, adapter)
		assert.Contains(t, err.Error(), "unsupported platform type")
	})
}

func TestManager_GetMessageSender(t *testing.T) {
	mockAdapter := &MockAdapter{}
	registry := map[entity.PlatformType]interface{}{
		entity.PlatformTypeWecom: mockAdapter,
	}
	manager := NewManagerWithRegistry(registry)

	t.Run("获取具备消息发送能力的适配器", func(t *testing.T) {
		sender, err := manager.GetMessageSender(entity.PlatformTypeWecom)
		assert.NoError(t, err)
		assert.NotNil(t, sender)
	})

	t.Run("获取不存在的平台", func(t *testing.T) {
		sender, err := manager.GetMessageSender(entity.PlatformTypeDingtalk)
		assert.Error(t, err)
		assert.Nil(t, sender)
	})
}

func TestManager_GetWebhookProcessor(t *testing.T) {
	mockAdapter := &MockAdapter{}
	registry := map[entity.PlatformType]interface{}{
		entity.PlatformTypeWecom: mockAdapter,
	}
	manager := NewManagerWithRegistry(registry)

	processor, err := manager.GetWebhookProcessor(entity.PlatformTypeWecom)
	assert.NoError(t, err)
	assert.NotNil(t, processor)
}

func TestManager_GetTokenManager(t *testing.T) {
	mockAdapter := &MockAdapter{}
	registry := map[entity.PlatformType]interface{}{
		entity.PlatformTypeWecom: mockAdapter,
	}
	manager := NewManagerWithRegistry(registry)

	tokenManager, err := manager.GetTokenManager(entity.PlatformTypeWecom)
	assert.NoError(t, err)
	assert.NotNil(t, tokenManager)
}

func TestManager_GetConfigValidator(t *testing.T) {
	mockAdapter := &MockAdapter{}
	registry := map[entity.PlatformType]interface{}{
		entity.PlatformTypeWecom: mockAdapter,
	}
	manager := NewManagerWithRegistry(registry)

	validator, err := manager.GetConfigValidator(entity.PlatformTypeWecom)
	assert.NoError(t, err)
	assert.NotNil(t, validator)
}

func TestManager_GetSupportedPlatforms(t *testing.T) {
	registry := map[entity.PlatformType]interface{}{
		entity.PlatformTypeWecom:    &MockAdapter{},
		entity.PlatformTypeDingtalk: &MockAdapter{},
		entity.PlatformTypeFeishu:   &MockAdapter{},
	}
	manager := NewManagerWithRegistry(registry)

	platforms := manager.GetSupportedPlatforms()
	assert.Len(t, platforms, 3)

	// 验证所有平台都在列表中
	platformMap := make(map[entity.PlatformType]bool)
	for _, p := range platforms {
		platformMap[p] = true
	}

	assert.True(t, platformMap[entity.PlatformTypeWecom])
	assert.True(t, platformMap[entity.PlatformTypeDingtalk])
	assert.True(t, platformMap[entity.PlatformTypeFeishu])
}

func TestManager_ValidateConfig(t *testing.T) {
	mockAdapter := &MockAdapter{}
	registry := map[entity.PlatformType]interface{}{
		entity.PlatformTypeWecom: mockAdapter,
	}
	manager := NewManagerWithRegistry(registry)

	config := map[string]interface{}{
		"corp_id": "test_corp_id",
		"secret":  "test_secret",
	}

	// 设置模拟预期
	mockAdapter.On("ValidateConfig", config).Return(nil)

	err := manager.ValidateConfig(entity.PlatformTypeWecom, config)
	assert.NoError(t, err)

	// 验证调用
	mockAdapter.AssertExpectations(t)
}

func TestManager_SendMessage(t *testing.T) {
	mockAdapter := &MockAdapter{}
	registry := map[entity.PlatformType]interface{}{
		entity.PlatformTypeWecom: mockAdapter,
	}
	manager := NewManagerWithRegistry(registry)

	ctx := context.Background()
	message := &dto.UnifiedMessage{
		MessageID: "test_msg_123",
		Content:   "测试消息",
	}
	config := map[string]interface{}{
		"corp_id": "test_corp_id",
	}
	accessToken := "test_token"

	// 设置模拟预期
	mockAdapter.On("SendMessage", ctx, message, config, accessToken).Return(nil)

	err := manager.SendMessage(ctx, entity.PlatformTypeWecom, message, config, accessToken)
	assert.NoError(t, err)

	// 验证调用
	mockAdapter.AssertExpectations(t)
}
