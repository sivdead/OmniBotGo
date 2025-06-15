package wecom

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

// WecomAdapter 企业微信平台适配器
type WecomAdapter struct {
	httpClient *http.Client
}

// NewWecomAdapter 创建企业微信适配器实例
func NewWecomAdapter() *WecomAdapter {
	return &WecomAdapter{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetPlatformType 获取平台类型
func (w *WecomAdapter) GetPlatformType() entity.PlatformType {
	return entity.PlatformTypeWecom
}

// ValidateConfig 验证平台配置
func (w *WecomAdapter) ValidateConfig(config map[string]interface{}) error {
	corpID, ok := config["corp_id"].(string)
	if !ok || corpID == "" {
		return fmt.Errorf("corp_id is required")
	}

	agentID, ok := config["agent_id"].(string)
	if !ok || agentID == "" {
		return fmt.Errorf("agent_id is required")
	}

	secret, ok := config["secret"].(string)
	if !ok || secret == "" {
		return fmt.Errorf("secret is required")
	}

	return nil
}

// GetAccessToken 获取访问令牌
func (w *WecomAdapter) GetAccessToken(ctx context.Context, config map[string]interface{}) (*entity.AccessTokenResponse, error) {
	if err := w.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	corpID := config["corp_id"].(string)
	secret := config["secret"].(string)

	// 构建请求URL
	reqURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		url.QueryEscape(corpID), url.QueryEscape(secret))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := w.httpClient.Do(req)
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
		return nil, fmt.Errorf("wecom API error: %d - %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &entity.AccessTokenResponse{
		AccessToken: tokenResp.AccessToken,
		ExpiresIn:   tokenResp.ExpiresIn,
		ExpiresAt:   &expiresAt,
	}, nil
}

// RefreshAccessToken 刷新访问令牌
func (w *WecomAdapter) RefreshAccessToken(ctx context.Context, config map[string]interface{}, oldToken string) (*entity.AccessTokenResponse, error) {
	// 企业微信的访问令牌刷新实际上就是重新获取
	return w.GetAccessToken(ctx, config)
}

// VerifyWebhook 验证Webhook请求
func (w *WecomAdapter) VerifyWebhook(ctx context.Context, signature string, timestamp string, nonce string, body []byte, config map[string]interface{}) error {
	token, ok := config["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("token is required for webhook verification")
	}

	// 企业微信验证逻辑：将token、timestamp、nonce三个参数进行字典序排序
	params := []string{token, timestamp, nonce}
	sort.Strings(params)

	// 将三个参数字符串拼接成一个字符串进行sha1加密
	str := strings.Join(params, "")
	hash := sha1.Sum([]byte(str))
	expectedSignature := hex.EncodeToString(hash[:])

	if signature != expectedSignature {
		return fmt.Errorf("webhook signature verification failed")
	}

	return nil
}

// ParseInboundMessage 解析入站消息
func (w *WecomAdapter) ParseInboundMessage(ctx context.Context, body []byte, config map[string]interface{}) (*entity.UnifiedMessage, error) {
	var wecomMsg WecomMessage
	if err := json.Unmarshal(body, &wecomMsg); err != nil {
		return nil, fmt.Errorf("failed to parse wecom message: %w", err)
	}

	// 转换为统一消息格式
	unifiedMsg := &entity.UnifiedMessage{
		MessageID:         generateMessageID(wecomMsg.MsgID, wecomMsg.CreateTime),
		MessageType:       mapWecomMessageType(wecomMsg.MsgType),
		SenderID:          wecomMsg.FromUserName,
		SenderName:        wecomMsg.FromUserName, // 企业微信可能需要额外API获取用户名
		SenderType:        "user",
		ReceiverID:        wecomMsg.ToUserName,
		ReceiverName:      wecomMsg.ToUserName,
		ReceiverType:      "bot",
		Content:           wecomMsg.Content,
		RawContent:        convertToMap(wecomMsg),
		ConversationID:    wecomMsg.FromUserName, // 使用发送者作为会话ID
		PlatformMessageID: strconv.FormatInt(wecomMsg.MsgID, 10),
		PlatformTimestamp: time.Unix(wecomMsg.CreateTime, 0),
	}

	// 处理媒体消息
	switch wecomMsg.MsgType {
	case "image":
		unifiedMsg.MediaURL = wecomMsg.PicURL
		unifiedMsg.MediaType = "image"
	case "voice":
		unifiedMsg.MediaURL = wecomMsg.MediaID
		unifiedMsg.MediaType = "audio"
	case "video", "shortvideo":
		unifiedMsg.MediaURL = wecomMsg.MediaID
		unifiedMsg.MediaType = "video"
	case "file":
		unifiedMsg.MediaURL = wecomMsg.MediaID
		unifiedMsg.MediaType = "file"
	}

	return unifiedMsg, nil
}

// SendMessage 发送消息
func (w *WecomAdapter) SendMessage(ctx context.Context, message *entity.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	agentID := config["agent_id"].(string)

	// 构建企业微信消息格式
	sendMsg := WecomSendMessage{
		ToUser:  message.ReceiverID,
		MsgType: mapToWecomMessageType(message.MessageType),
		AgentID: agentID,
	}

	// 根据消息类型设置内容
	switch message.MessageType {
	case "text":
		sendMsg.Text = &WecomTextMessage{
			Content: message.Content,
		}
	case "image":
		sendMsg.Image = &WecomImageMessage{
			MediaID: message.MediaURL,
		}
	// 可以继续添加其他消息类型
	default:
		// 默认作为文本消息处理
		sendMsg.Text = &WecomTextMessage{
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
	reqURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", accessToken)

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(msgData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
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
func (w *WecomAdapter) BuildWebhookPath(channelID int64) string {
	return fmt.Sprintf("/webhook/wecom/%d", channelID)
}

// 辅助函数

func generateMessageID(msgID int64, createTime int64) string {
	return fmt.Sprintf("wecom_%d_%d", msgID, createTime)
}

func mapWecomMessageType(msgType string) string {
	switch msgType {
	case "text":
		return "text"
	case "image":
		return "image"
	case "voice":
		return "audio"
	case "video", "shortvideo":
		return "video"
	case "file":
		return "file"
	case "event":
		return "event"
	default:
		return "unknown"
	}
}

func mapToWecomMessageType(msgType string) string {
	switch msgType {
	case "text":
		return "text"
	case "image":
		return "image"
	case "audio":
		return "voice"
	case "video":
		return "video"
	case "file":
		return "file"
	default:
		return "text"
	}
}

func convertToMap(msg WecomMessage) map[string]interface{} {
	data, _ := json.Marshal(msg)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}
