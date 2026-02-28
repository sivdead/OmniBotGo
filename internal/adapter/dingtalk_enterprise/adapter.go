package dingtalk_enterprise

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

var (
	errDingtalkMissingAppKey        = errors.New("app_key is required in dingtalk enterprise config")
	errDingtalkMissingAppSecret     = errors.New("app_secret is required in dingtalk enterprise config")
	errDingtalkMissingWebhookSecret = errors.New("app_secret is required in dingtalk enterprise config for webhook verification")
	errDingtalkWebhookSigFailed     = errors.New("webhook signature verification failed")
)

const (
	// API URLs
	tokenURL        = "https://oapi.dingtalk.com/gettoken"
	workNoticeURL   = "https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2"
	groupMessageURL = "https://oapi.dingtalk.com/chat/send"
	dingMessageURL  = "https://oapi.dingtalk.com/topapi/message/send_to_conversation"
	userInfoURL     = "https://oapi.dingtalk.com/topapi/v2/user/get"

	// Token过期时间（秒）
	tokenExpireTime = 7200
)

// DingtalkEnterpriseAdapter 钉钉企业应用适配器
type DingtalkEnterpriseAdapter struct {
	logger          zerolog.Logger
	httpClient      *http.Client
	tokenCache      map[string]*tokenCacheItem
	tokenMutex      sync.RWMutex
	webhookHandlers map[string]port.MessageHandler
	handlerMutex    sync.RWMutex
}

// tokenCacheItem 令牌缓存项
type tokenCacheItem struct {
	Token      string
	ExpireTime time.Time
}

// NewAdapter 创建钉钉企业应用适配器
func NewAdapter(logger zerolog.Logger) *DingtalkEnterpriseAdapter {
	return &DingtalkEnterpriseAdapter{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tokenCache:      make(map[string]*tokenCacheItem),
		webhookHandlers: make(map[string]port.MessageHandler),
	}
}

// GetPlatformType 实现PlatformIdentifier接口
func (a *DingtalkEnterpriseAdapter) GetPlatformType() entity.PlatformType {
	return entity.PlatformTypeDingtalk
}

// ValidateConfig 实现ConfigValidator接口
func (a *DingtalkEnterpriseAdapter) ValidateConfig(cfg map[string]interface{}) error {
	return config.ValidatePlatformConfig(entity.PlatformTypeDingtalk, cfg)
}

// GetAccessToken 实现TokenManager接口。
func (a *DingtalkEnterpriseAdapter) GetAccessToken(ctx context.Context, config map[string]interface{}) (*dto.AccessTokenResponse, error) {
	appKey, ok := config["app_key"].(string)
	if !ok || appKey == "" {
		return nil, errDingtalkMissingAppKey
	}

	appSecret, ok := config["app_secret"].(string)
	if !ok || appSecret == "" {
		return nil, errDingtalkMissingAppSecret
	}

	// 检查缓存
	cacheKey := fmt.Sprintf("%s:%s", appKey, appSecret)
	a.tokenMutex.RLock()
	if cached, ok := a.tokenCache[cacheKey]; ok && cached.ExpireTime.After(time.Now()) {
		a.tokenMutex.RUnlock()
		return &dto.AccessTokenResponse{
			AccessToken: cached.Token,
			ExpiresIn:   int(time.Until(cached.ExpireTime).Seconds()),
		}, nil
	}
	a.tokenMutex.RUnlock()

	// 获取新令牌
	params := url.Values{}
	params.Add("appkey", appKey)
	params.Add("appsecret", appSecret)

	reqURL := fmt.Sprintf("%s?%s", tokenURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("dingtalk API error: %s (code: %d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	// 缓存令牌
	a.tokenMutex.Lock()
	a.tokenCache[cacheKey] = &tokenCacheItem{
		Token:      tokenResp.AccessToken,
		ExpireTime: time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}
	a.tokenMutex.Unlock()

	return &dto.AccessTokenResponse{
		AccessToken: tokenResp.AccessToken,
		ExpiresIn:   tokenResp.ExpiresIn,
	}, nil
}

// RefreshAccessToken 实现TokenManager接口
func (a *DingtalkEnterpriseAdapter) RefreshAccessToken(ctx context.Context, config map[string]interface{}, oldToken string) (*dto.AccessTokenResponse, error) {
	// 钉钉的access_token刷新机制是重新获取
	return a.GetAccessToken(ctx, config)
}

// SendMessage 实现MessageSender接口
func (a *DingtalkEnterpriseAdapter) SendMessage(ctx context.Context, message *dto.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	// 根据消息类型和接收者类型选择发送方式
	switch message.ReceiverType {
	case entity.ReceiverTypeUser:
		return a.sendWorkNotice(ctx, message, config, accessToken)
	case entity.ReceiverTypeGroup:
		return a.sendGroupMessage(ctx, message, config, accessToken)
	default:
		return fmt.Errorf("unsupported receiver type: %s", message.ReceiverType)
	}
}

// ProcessWebhookMessage 实现WebhookProcessor接口
func (a *DingtalkEnterpriseAdapter) ProcessWebhookMessage(ctx context.Context, request *http.Request, channelID int64) (*dto.UnifiedMessage, error) {
	// 读取请求体
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// 验证签名（如果提供了签名头）
	signature := request.Header.Get("sign")
	timestamp := request.Header.Get("timestamp")
	if signature != "" && timestamp != "" {
		// 这里需要从通道配置中获取app_secret进行验证
		// 暂时跳过签名验证，实际部署时应该启用
		a.logger.Warn().
			Str("signature", signature).
			Str("timestamp", timestamp).
			Msg("钉钉webhook签名验证已跳过，生产环境请启用")
	}

	// 解析消息
	var webhookData DingtalkWebhookData
	if err := json.Unmarshal(body, &webhookData); err != nil {
		return nil, fmt.Errorf("failed to parse webhook data: %w", err)
	}

	// 转换为统一消息格式
	return webhookData.ToUnifiedMessage(), nil
}

// RegisterWebhookHandler 注册Webhook处理器
func (a *DingtalkEnterpriseAdapter) RegisterWebhookHandler(channelID string, handler port.MessageHandler) {
	a.handlerMutex.Lock()
	defer a.handlerMutex.Unlock()
	a.webhookHandlers[channelID] = handler
}

// VerifyWebhook 实现WebhookProcessor接口。
func (a *DingtalkEnterpriseAdapter) VerifyWebhook(_ context.Context, signature, timestamp, _ string, _ []byte, cfg map[string]interface{}) error {
	appSecret, ok := cfg["app_secret"].(string)
	if !ok || appSecret == "" {
		return errDingtalkMissingWebhookSecret
	}

	if !a.verifyWebhookSignature(timestamp, appSecret, signature) {
		a.logger.Warn().
			Str("signature", signature).
			Str("timestamp", timestamp).
			Msg("钉钉企业应用Webhook签名验证失败")

		return errDingtalkWebhookSigFailed
	}

	return nil
}

// ParseInboundMessage 实现WebhookProcessor接口。
func (a *DingtalkEnterpriseAdapter) ParseInboundMessage(_ context.Context, body []byte, _ map[string]interface{}) (*dto.UnifiedMessage, error) {
	var webhookData DingtalkWebhookData
	if err := json.Unmarshal(body, &webhookData); err != nil {
		return nil, fmt.Errorf("failed to parse dingtalk enterprise webhook data: %w", err)
	}

	return webhookData.ToUnifiedMessage(), nil
}

// BuildWebhookPath 实现WebhookProcessor接口。
func (a *DingtalkEnterpriseAdapter) BuildWebhookPath(channelID int64) string {
	return fmt.Sprintf("/webhook/dingtalk/%d", channelID)
}

// 编译期接口实现检查。
var (
	_ port.MessageSender      = (*DingtalkEnterpriseAdapter)(nil)
	_ port.TokenManager       = (*DingtalkEnterpriseAdapter)(nil)
	_ port.WebhookProcessor   = (*DingtalkEnterpriseAdapter)(nil)
	_ port.PlatformIdentifier = (*DingtalkEnterpriseAdapter)(nil)
	_ port.ConfigValidator    = (*DingtalkEnterpriseAdapter)(nil)
)

// sendWorkNotice 发送工作通知
func (a *DingtalkEnterpriseAdapter) sendWorkNotice(ctx context.Context, message *dto.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	agentID, _ := config["agent_id"].(string)
	if agentID == "" {
		return fmt.Errorf("agent_id is required for work notice")
	}

	// 构建请求
	msg := buildDingtalkMessage(message)

	reqBody := map[string]interface{}{
		"agent_id":    agentID,
		"userid_list": message.ReceiverID,
		"msg":         msg,
	}

	// 发送请求
	return a.sendAPIRequest(ctx, workNoticeURL, accessToken, reqBody)
}

// sendGroupMessage 发送群消息
func (a *DingtalkEnterpriseAdapter) sendGroupMessage(ctx context.Context, message *dto.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	// 构建消息
	msg := buildDingtalkMessage(message)

	reqBody := map[string]interface{}{
		"chatid": message.ReceiverID,
		"msg":    msg,
	}

	// 发送请求
	return a.sendAPIRequest(ctx, groupMessageURL, accessToken, reqBody)
}

// sendAPIRequest 发送API请求
func (a *DingtalkEnterpriseAdapter) sendAPIRequest(ctx context.Context, apiURL, accessToken string, reqBody interface{}) error {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	fullURL := fmt.Sprintf("%s?access_token=%s", apiURL, accessToken)

	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		TaskID  int64  `json:"task_id,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.ErrCode != 0 {
		return fmt.Errorf("dingtalk API error: %s (code: %d)", apiResp.ErrMsg, apiResp.ErrCode)
	}

	a.logger.Info().
		Int64("task_id", apiResp.TaskID).
		Msg("dingtalk message sent successfully")

	return nil
}

// buildDingtalkMessage 构建钉钉消息格式
func buildDingtalkMessage(message *dto.UnifiedMessage) map[string]interface{} {
	switch message.MessageType {
	case entity.MessageTypeText:
		return map[string]interface{}{
			"msgtype": "text",
			"text": map[string]interface{}{
				"content": message.Content,
			},
		}

	case entity.MessageTypeMarkdown:
		title := "消息"
		content := message.Content
		if message.MarkdownContent != nil {
			title = message.MarkdownContent.Title
			content = message.MarkdownContent.Content
		}
		return map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]interface{}{
				"title": title,
				"text":  content,
			},
		}

	case entity.MessageTypeCard:
		if message.CardContent != nil {
			actionCard := map[string]interface{}{
				"msgtype": "action_card",
				"action_card": map[string]interface{}{
					"title":        message.CardContent.Title,
					"markdown":     message.CardContent.Content,
					"single_title": "查看详情",
					"single_url":   "https://www.dingtalk.com",
				},
			}

			// 处理按钮
			if len(message.CardContent.Buttons) > 0 {
				if len(message.CardContent.Buttons) == 1 {
					actionCard["action_card"].(map[string]interface{})["single_title"] = message.CardContent.Buttons[0].Title
					actionCard["action_card"].(map[string]interface{})["single_url"] = message.CardContent.Buttons[0].ActionURL
				} else {
					btns := make([]map[string]interface{}, 0, len(message.CardContent.Buttons))
					for _, btn := range message.CardContent.Buttons {
						btns = append(btns, map[string]interface{}{
							"title":      btn.Title,
							"action_url": btn.ActionURL,
						})
					}
					actionCard["action_card"].(map[string]interface{})["btn_json_list"] = btns
					delete(actionCard["action_card"].(map[string]interface{}), "single_title")
					delete(actionCard["action_card"].(map[string]interface{}), "single_url")
				}
			}

			return actionCard
		}

	case entity.MessageTypeLink:
		return map[string]interface{}{
			"msgtype": "link",
			"link": map[string]interface{}{
				"title":      message.Content,
				"text":       "",
				"picUrl":     "",
				"messageUrl": "",
			},
		}

	case entity.MessageTypeImage:
		// 企业应用发送图片需要先上传获取media_id
		return map[string]interface{}{
			"msgtype": "image",
			"image": map[string]interface{}{
				"media_id": message.RawContent["media_id"],
			},
		}

	case entity.MessageTypeFile:
		// 企业应用发送文件需要先上传获取media_id
		return map[string]interface{}{
			"msgtype": "file",
			"file": map[string]interface{}{
				"media_id": message.RawContent["media_id"],
			},
		}

	case entity.MessageTypeNews:
		// 图文消息
		if message.NewsContent != nil {
			articles := make([]map[string]interface{}, 0, len(message.NewsContent.Articles))
			for _, article := range message.NewsContent.Articles {
				articles = append(articles, map[string]interface{}{
					"title":       article.Title,
					"description": article.Description,
					"url":         article.URL,
					"picurl":      article.PicURL,
				})
			}
			return map[string]interface{}{
				"msgtype": "news",
				"news": map[string]interface{}{
					"articles": articles,
				},
			}
		}
	}

	// 默认返回文本消息
	return map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": message.Content,
		},
	}
}

// verifySignature 验证钉钉签名
func verifySignature(timestamp, secret, signature string) bool {
	// 钉钉签名验证算法：
	// 1. 把timestamp + "\n" + secret当做签名字符串
	// 2. 使用HmacSHA256算法计算签名
	// 3. 然后进行Base64 encode，得到最终的签名

	// 构建签名字符串
	stringToSign := timestamp + "\n" + secret

	// 使用HMAC-SHA256计算签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))

	// Base64编码
	calculatedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 比较签名
	return calculatedSignature == signature
}

// verifyWebhookSignature 验证钉钉Webhook签名
func (a *DingtalkEnterpriseAdapter) verifyWebhookSignature(timestamp, secret, signature string) bool {
	if secret == "" {
		a.logger.Warn().Msg("钉钉webhook secret为空，跳过签名验证")
		return true
	}

	// 验证时间戳（防重放攻击）
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		a.logger.Error().Err(err).Msg("无效的时间戳格式")
		return false
	}

	// 检查时间戳是否在合理范围内（5分钟内）
	now := time.Now().Unix() * 1000 // 钉钉使用毫秒时间戳
	if abs(now-ts) > 5*60*1000 {
		a.logger.Warn().
			Int64("timestamp", ts).
			Int64("now", now).
			Msg("时间戳超出允许范围")
		return false
	}

	return verifySignature(timestamp, secret, signature)
}

// abs 返回整数的绝对值
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
