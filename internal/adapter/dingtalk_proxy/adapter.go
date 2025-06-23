package dingtalk_proxy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/adapter/dingtalk_enterprise"
	"github.com/sivdead/OmniBotGo/internal/adapter/dingtalk_stream"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

// DingtalkProxyAdapter 钉钉适配器代理，根据配置选择具体实现
type DingtalkProxyAdapter struct {
	logger            zerolog.Logger
	streamAdapter     *dingtalk_stream.DingtalkStreamAdapter
	enterpriseAdapter *dingtalk_enterprise.DingtalkEnterpriseAdapter
}

// NewAdapter 创建钉钉代理适配器
func NewAdapter(logger zerolog.Logger,
	streamAdapter *dingtalk_stream.DingtalkStreamAdapter,
	enterpriseAdapter *dingtalk_enterprise.DingtalkEnterpriseAdapter) *DingtalkProxyAdapter {
	return &DingtalkProxyAdapter{
		logger:            logger,
		streamAdapter:     streamAdapter,
		enterpriseAdapter: enterpriseAdapter,
	}
}

// GetPlatformType 实现PlatformIdentifier接口
func (a *DingtalkProxyAdapter) GetPlatformType() entity.PlatformType {
	return entity.PlatformTypeDingtalk
}

// ValidateConfig 实现ConfigValidator接口
func (a *DingtalkProxyAdapter) ValidateConfig(cfg map[string]interface{}) error {
	// 判断是Stream模式还是企业应用模式
	if isStreamMode(cfg) {
		return a.streamAdapter.ValidateConfig(cfg)
	}
	return a.enterpriseAdapter.ValidateConfig(cfg)
}

// GetAccessToken 实现TokenManager接口
func (a *DingtalkProxyAdapter) GetAccessToken(ctx context.Context, config map[string]interface{}) (*dto.AccessTokenResponse, error) {
	// Stream模式不需要access token
	if isStreamMode(config) {
		return nil, fmt.Errorf("stream mode does not support access token")
	}
	return a.enterpriseAdapter.GetAccessToken(ctx, config)
}

// RefreshAccessToken 实现TokenManager接口
func (a *DingtalkProxyAdapter) RefreshAccessToken(ctx context.Context, config map[string]interface{}, oldToken string) (*dto.AccessTokenResponse, error) {
	if isStreamMode(config) {
		return nil, fmt.Errorf("stream mode does not support access token")
	}
	return a.enterpriseAdapter.RefreshAccessToken(ctx, config, oldToken)
}

// SendMessage 实现MessageSender接口
func (a *DingtalkProxyAdapter) SendMessage(ctx context.Context, message *dto.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	if isStreamMode(config) {
		return a.streamAdapter.SendMessage(ctx, message, config, accessToken)
	}
	return a.enterpriseAdapter.SendMessage(ctx, message, config, accessToken)
}

// ProcessWebhookMessage 实现WebhookProcessor接口
func (a *DingtalkProxyAdapter) ProcessWebhookMessage(ctx context.Context, request *http.Request, channelID int64) (*dto.UnifiedMessage, error) {
	// 只有企业应用模式支持Webhook
	if request.Header.Get("x-dingtalk-signature") != "" {
		return a.enterpriseAdapter.ProcessWebhookMessage(ctx, request, channelID)
	}
	return nil, fmt.Errorf("invalid webhook request")
}

// VerifyWebhook 实现WebhookProcessor接口
func (a *DingtalkProxyAdapter) VerifyWebhook(ctx context.Context, signature, timestamp, nonce string, body []byte, config entity.JSONField) error {
	// 根据配置判断使用哪种验证方式
	cfg := make(map[string]interface{})
	if err := config.Scan(&cfg); err != nil {
		return err
	}

	if isStreamMode(cfg) {
		// Stream模式不需要验证Webhook
		return fmt.Errorf("stream mode does not support webhook")
	}

	// 企业应用模式验证
	// 实现企业应用的Webhook验证
	// 获取app_secret从配置中进行签名验证
	appSecret, ok := cfg["app_secret"].(string)
	if !ok || appSecret == "" {
		return fmt.Errorf("app_secret is required for webhook verification")
	}
	
	// 使用企业应用适配器的验证逻辑
	if !verifySignature(timestamp, nonce, appSecret, signature) {
		return fmt.Errorf("webhook signature verification failed")
	}
	
	return nil
}

// ParseInboundMessage 实现WebhookProcessor接口
func (a *DingtalkProxyAdapter) ParseInboundMessage(ctx context.Context, body []byte, config entity.JSONField) (*dto.UnifiedMessage, error) {
	cfg := make(map[string]interface{})
	if err := config.Scan(&cfg); err != nil {
		return nil, err
	}

	if isStreamMode(cfg) {
		return nil, fmt.Errorf("stream mode does not support webhook parsing")
	}

	// 解析企业应用的Webhook消息
	var webhookData dingtalk_enterprise.DingtalkWebhookData
	if err := json.Unmarshal(body, &webhookData); err != nil {
		return nil, fmt.Errorf("failed to parse webhook data: %w", err)
	}

	return webhookData.ToUnifiedMessage(), nil
}

// Start 实现StreamAdapter接口
func (a *DingtalkProxyAdapter) Start(ctx context.Context, messageHandler port.MessageHandler, cfg map[string]interface{}) error {
	if !isStreamMode(cfg) {
		return fmt.Errorf("enterprise mode does not support stream connection")
	}
	return a.streamAdapter.Start(ctx, messageHandler, cfg)
}

// Stop 实现StreamAdapter接口
func (a *DingtalkProxyAdapter) Stop(ctx context.Context) error {
	return a.streamAdapter.Stop(ctx)
}

// IsConnected 实现StreamAdapter接口
func (a *DingtalkProxyAdapter) IsConnected() bool {
	return a.streamAdapter.IsConnected()
}

// isStreamMode 判断是否为Stream模式
func isStreamMode(config map[string]interface{}) bool {
	// 如果配置了client_id和client_secret，则为Stream模式
	clientID, hasClientID := config["client_id"].(string)
	clientSecret, hasClientSecret := config["client_secret"].(string)

	return hasClientID && hasClientSecret && clientID != "" && clientSecret != ""
}

// verifySignature 验证钉钉签名
func verifySignature(timestamp, nonce, token, signature string) bool {
	if token == "" || signature == "" {
		return false
	}
	
	// 将timestamp、nonce、token按字典序排序
	params := []string{timestamp, nonce, token}
	sort.Strings(params)
	
	// 拼接成字符串
	data := strings.Join(params, "")
	
	// 进行sha256加密
	h := sha256.New()
	h.Write([]byte(data))
	hash := hex.EncodeToString(h.Sum(nil))
	
	// 将加密结果与signature对比
	return hash == signature
}
