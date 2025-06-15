package feishu

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

// FeishuAdapter 飞书平台适配器
type FeishuAdapter struct {
	httpClient *http.Client
}

// NewFeishuAdapter 创建飞书适配器实例
func NewFeishuAdapter() *FeishuAdapter {
	return &FeishuAdapter{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetPlatformType 获取平台类型
func (f *FeishuAdapter) GetPlatformType() entity.PlatformType {
	return entity.PlatformTypeFeishu
}

// ValidateConfig 验证平台配置
func (f *FeishuAdapter) ValidateConfig(config map[string]interface{}) error {
	appID, ok := config["app_id"].(string)
	if !ok || appID == "" {
		return fmt.Errorf("app_id is required")
	}

	appSecret, ok := config["app_secret"].(string)
	if !ok || appSecret == "" {
		return fmt.Errorf("app_secret is required")
	}

	return nil
}

// GetAccessToken 获取访问令牌
func (f *FeishuAdapter) GetAccessToken(ctx context.Context, config map[string]interface{}) (*entity.AccessTokenResponse, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	appID := config["app_id"].(string)
	appSecret := config["app_secret"].(string)

	// 构建请求体
	reqBody := map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal", strings.NewReader(string(reqData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenResp struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.Code != 0 {
		return nil, fmt.Errorf("feishu API error: %d - %s", tokenResp.Code, tokenResp.Msg)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.Expire) * time.Second)

	return &entity.AccessTokenResponse{
		AccessToken: tokenResp.TenantAccessToken,
		ExpiresIn:   tokenResp.Expire,
		ExpiresAt:   &expiresAt,
	}, nil
}

// RefreshAccessToken 刷新访问令牌
func (f *FeishuAdapter) RefreshAccessToken(ctx context.Context, config map[string]interface{}, oldToken string) (*entity.AccessTokenResponse, error) {
	// 飞书的访问令牌刷新实际上就是重新获取
	return f.GetAccessToken(ctx, config)
}

// VerifyWebhook 验证Webhook请求
func (f *FeishuAdapter) VerifyWebhook(ctx context.Context, signature string, timestamp string, nonce string, body []byte, config map[string]interface{}) error {
	token, ok := config["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("token is required for webhook verification")
	}

	// 飞书签名验证逻辑
	stringToSign := timestamp + nonce + string(body)
	hash := sha1.Sum([]byte(stringToSign + token))
	expectedSignature := hex.EncodeToString(hash[:])

	if signature != expectedSignature {
		return fmt.Errorf("webhook signature verification failed")
	}

	return nil
}

// ParseInboundMessage 解析入站消息
func (f *FeishuAdapter) ParseInboundMessage(ctx context.Context, body []byte, config map[string]interface{}) (*entity.UnifiedMessage, error) {
	var feishuMsg FeishuMessage
	if err := json.Unmarshal(body, &feishuMsg); err != nil {
		return nil, fmt.Errorf("failed to parse feishu message: %w", err)
	}

	// 转换为统一消息格式
	unifiedMsg := &entity.UnifiedMessage{
		MessageID:         generateMessageID(feishuMsg.Event.Message.MessageID, feishuMsg.Event.Message.CreateTime),
		MessageType:       mapFeishuMessageType(feishuMsg.Event.Message.MessageType),
		SenderID:          feishuMsg.Event.Sender.SenderID.UserID,
		SenderName:        feishuMsg.Event.Sender.SenderID.UserID, // 飞书可能需要额外API获取用户名
		SenderType:        "user",
		ReceiverID:        feishuMsg.Event.Message.ChatID,
		ReceiverName:      feishuMsg.Event.Message.ChatID,
		ReceiverType:      "chat",
		Content:           extractFeishuContent(feishuMsg.Event.Message),
		RawContent:        convertToMap(feishuMsg),
		ConversationID:    feishuMsg.Event.Message.ChatID,
		PlatformMessageID: feishuMsg.Event.Message.MessageID,
		PlatformTimestamp: time.Unix(0, feishuMsg.Event.Message.CreateTime*1000000), // 飞书使用毫秒时间戳
	}

	return unifiedMsg, nil
}

// SendMessage 发送消息
func (f *FeishuAdapter) SendMessage(ctx context.Context, message *entity.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	// 构建飞书消息格式
	sendMsg := FeishuSendMessage{
		ReceiveID: message.ReceiverID,
		MsgType:   mapToFeishuMessageType(message.MessageType),
	}

	// 根据消息类型设置内容
	switch message.MessageType {
	case "text":
		sendMsg.Content = map[string]interface{}{
			"text": message.Content,
		}
	case "markdown":
		sendMsg.Content = map[string]interface{}{
			"text": message.Content,
		}
	// 可以继续添加其他消息类型
	default:
		// 默认作为文本消息处理
		sendMsg.Content = map[string]interface{}{
			"text": message.Content,
		}
		sendMsg.MsgType = "text"
	}

	// 序列化消息
	msgData, err := json.Marshal(sendMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 发送消息
	reqURL := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=chat_id"

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(msgData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var sendResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	if err := json.Unmarshal(body, &sendResp); err != nil {
		return fmt.Errorf("failed to parse send response: %w", err)
	}

	if sendResp.Code != 0 {
		return fmt.Errorf("failed to send message: %d - %s", sendResp.Code, sendResp.Msg)
	}

	return nil
}

// BuildWebhookPath 构建Webhook路径
func (f *FeishuAdapter) BuildWebhookPath(channelID int64) string {
	return fmt.Sprintf("/webhook/feishu/%d", channelID)
}

// 辅助函数

func generateMessageID(msgID string, createTime int64) string {
	return fmt.Sprintf("feishu_%s_%d", msgID, createTime)
}

func mapFeishuMessageType(msgType string) string {
	switch msgType {
	case "text":
		return "text"
	case "image":
		return "image"
	case "audio":
		return "audio"
	case "video":
		return "video"
	case "file":
		return "file"
	case "post":
		return "markdown"
	default:
		return "unknown"
	}
}

func mapToFeishuMessageType(msgType string) string {
	switch msgType {
	case "text":
		return "text"
	case "image":
		return "image"
	case "audio":
		return "audio"
	case "video":
		return "video"
	case "file":
		return "file"
	case "markdown":
		return "post"
	default:
		return "text"
	}
}

func extractFeishuContent(msg FeishuMessageContent) string {
	if msg.Content != nil {
		if text, ok := msg.Content["text"].(string); ok {
			return text
		}
	}
	return ""
}

func convertToMap(msg FeishuMessage) map[string]interface{} {
	data, _ := json.Marshal(msg)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}
