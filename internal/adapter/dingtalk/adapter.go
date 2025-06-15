package dingtalk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

// DingtalkAdapter 钉钉平台适配器
type DingtalkAdapter struct {
	httpClient *http.Client
}

// NewDingtalkAdapter 创建钉钉适配器实例
func NewDingtalkAdapter() *DingtalkAdapter {
	return &DingtalkAdapter{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetPlatformType 获取平台类型
func (d *DingtalkAdapter) GetPlatformType() entity.PlatformType {
	return entity.PlatformTypeDingtalk
}

// ValidateConfig 验证平台配置
func (d *DingtalkAdapter) ValidateConfig(config map[string]interface{}) error {
	appKey, ok := config["app_key"].(string)
	if !ok || appKey == "" {
		return fmt.Errorf("app_key is required")
	}

	appSecret, ok := config["app_secret"].(string)
	if !ok || appSecret == "" {
		return fmt.Errorf("app_secret is required")
	}

	agentID, ok := config["agent_id"].(string)
	if !ok || agentID == "" {
		return fmt.Errorf("agent_id is required")
	}

	return nil
}

// GetAccessToken 获取访问令牌
func (d *DingtalkAdapter) GetAccessToken(ctx context.Context, config map[string]interface{}) (*entity.AccessTokenResponse, error) {
	if err := d.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	appKey := config["app_key"].(string)
	appSecret := config["app_secret"].(string)

	// 构建请求URL
	reqURL := fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s",
		url.QueryEscape(appKey), url.QueryEscape(appSecret))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenResp struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("dingtalk API error: %d - %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &entity.AccessTokenResponse{
		AccessToken: tokenResp.AccessToken,
		ExpiresIn:   tokenResp.ExpiresIn,
		ExpiresAt:   &expiresAt,
	}, nil
}

// RefreshAccessToken 刷新访问令牌
func (d *DingtalkAdapter) RefreshAccessToken(ctx context.Context, config map[string]interface{}, oldToken string) (*entity.AccessTokenResponse, error) {
	// 钉钉的访问令牌刷新实际上就是重新获取
	return d.GetAccessToken(ctx, config)
}

// VerifyWebhook 验证Webhook请求
func (d *DingtalkAdapter) VerifyWebhook(ctx context.Context, signature string, timestamp string, nonce string, body []byte, config map[string]interface{}) error {
	token, ok := config["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("token is required for webhook verification")
	}

	// 钉钉签名验证逻辑
	stringToSign := timestamp + "\n" + token
	h := hmac.New(sha256.New, []byte(token))
	h.Write([]byte(stringToSign))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if signature != expectedSignature {
		return fmt.Errorf("webhook signature verification failed")
	}

	return nil
}

// ParseInboundMessage 解析入站消息
func (d *DingtalkAdapter) ParseInboundMessage(ctx context.Context, body []byte, config map[string]interface{}) (*entity.UnifiedMessage, error) {
	var dingtalkMsg DingtalkMessage
	if err := json.Unmarshal(body, &dingtalkMsg); err != nil {
		return nil, fmt.Errorf("failed to parse dingtalk message: %w", err)
	}

	// 转换为统一消息格式
	unifiedMsg := &entity.UnifiedMessage{
		MessageID:         generateMessageID(dingtalkMsg.MsgID, dingtalkMsg.CreateAt),
		MessageType:       mapDingtalkMessageType(dingtalkMsg.MsgType),
		SenderID:          dingtalkMsg.SenderID,
		SenderName:        dingtalkMsg.SenderNick,
		SenderType:        "user",
		ReceiverID:        dingtalkMsg.ChatbotUserID,
		ReceiverName:      dingtalkMsg.ChatbotUserID,
		ReceiverType:      "bot",
		Content:           extractContent(dingtalkMsg),
		RawContent:        convertToMap(dingtalkMsg),
		ConversationID:    dingtalkMsg.ConversationID,
		PlatformMessageID: dingtalkMsg.MsgID,
		PlatformTimestamp: time.Unix(dingtalkMsg.CreateAt/1000, 0), // 钉钉使用毫秒时间戳
	}

	// 处理媒体消息
	switch dingtalkMsg.MsgType {
	case "picture":
		unifiedMsg.MediaURL = dingtalkMsg.Content.PicURL
		unifiedMsg.MediaType = "image"
	case "audio":
		unifiedMsg.MediaURL = dingtalkMsg.Content.MediaID
		unifiedMsg.MediaType = "audio"
	case "video":
		unifiedMsg.MediaURL = dingtalkMsg.Content.MediaID
		unifiedMsg.MediaType = "video"
	case "file":
		unifiedMsg.MediaURL = dingtalkMsg.Content.MediaID
		unifiedMsg.MediaType = "file"
	}

	return unifiedMsg, nil
}

// SendMessage 发送消息
func (d *DingtalkAdapter) SendMessage(ctx context.Context, message *entity.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	// 构建钉钉消息格式
	sendMsg := DingtalkSendMessage{
		MsgType: mapToDingtalkMessageType(message.MessageType),
	}

	// 根据消息类型设置内容
	switch message.MessageType {
	case "text":
		sendMsg.Text = &DingtalkTextMessage{
			Content: message.Content,
		}
	case "markdown":
		sendMsg.Markdown = &DingtalkMarkdownMessage{
			Title: "消息",
			Text:  message.Content,
		}
	// 可以继续添加其他消息类型
	default:
		// 默认作为文本消息处理
		sendMsg.Text = &DingtalkTextMessage{
			Content: message.Content,
		}
		sendMsg.MsgType = "text"
	}

	// 序列化消息
	msgData, err := json.Marshal(sendMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 发送消息
	reqURL := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", accessToken)

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(msgData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var sendResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &sendResp); err != nil {
		return fmt.Errorf("failed to parse send response: %w", err)
	}

	if sendResp.ErrCode != 0 {
		return fmt.Errorf("failed to send message: %d - %s", sendResp.ErrCode, sendResp.ErrMsg)
	}

	return nil
}

// BuildWebhookPath 构建Webhook路径
func (d *DingtalkAdapter) BuildWebhookPath(channelID int64) string {
	return fmt.Sprintf("/webhook/dingtalk/%d", channelID)
}

// 辅助函数

func generateMessageID(msgID string, createAt int64) string {
	return fmt.Sprintf("dingtalk_%s_%d", msgID, createAt)
}

func mapDingtalkMessageType(msgType string) string {
	switch msgType {
	case "text":
		return "text"
	case "picture":
		return "image"
	case "audio":
		return "audio"
	case "video":
		return "video"
	case "file":
		return "file"
	case "markdown":
		return "markdown"
	default:
		return "unknown"
	}
}

func mapToDingtalkMessageType(msgType string) string {
	switch msgType {
	case "text":
		return "text"
	case "image":
		return "picture"
	case "audio":
		return "audio"
	case "video":
		return "video"
	case "file":
		return "file"
	case "markdown":
		return "markdown"
	default:
		return "text"
	}
}

func extractContent(msg DingtalkMessage) string {
	switch msg.MsgType {
	case "text":
		return msg.Text.Content
	case "picture":
		return "[图片]"
	case "audio":
		return "[语音]"
	case "video":
		return "[视频]"
	case "file":
		return "[文件]"
	default:
		return ""
	}
}

func convertToMap(msg DingtalkMessage) map[string]interface{} {
	data, _ := json.Marshal(msg)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}
