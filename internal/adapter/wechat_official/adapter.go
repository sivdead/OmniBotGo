package wechat_official

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

// WechatOfficialAdapter 微信公众号平台适配器
type WechatOfficialAdapter struct {
	httpClient *http.Client
}

// NewWechatOfficialAdapter 创建微信公众号适配器实例
func NewWechatOfficialAdapter() *WechatOfficialAdapter {
	return &WechatOfficialAdapter{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetPlatformType 获取平台类型
func (w *WechatOfficialAdapter) GetPlatformType() entity.PlatformType {
	return entity.PlatformTypeWechatOfficial
}

// ValidateConfig 验证平台配置
func (w *WechatOfficialAdapter) ValidateConfig(cfg map[string]interface{}) error {
	return config.ValidatePlatformConfig(entity.PlatformTypeWechatOfficial, cfg)
}

// SendMessage 发送消息
func (w *WechatOfficialAdapter) SendMessage(ctx context.Context, message *dto.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	// 构建微信公众号消息格式
	sendMsg := WechatOfficialSendMessage{
		ToUser:  message.ReceiverID,
		MsgType: mapToWechatOfficialMessageType(message.MessageType),
	}

	// 根据消息类型设置内容
	switch message.MessageType {
	case "text":
		sendMsg.Text = &WechatOfficialTextMessage{
			Content: message.Content,
		}
	case "image":
		sendMsg.Image = &WechatOfficialImageMessage{
			MediaID: message.MediaURL,
		}
	case "voice":
		sendMsg.Voice = &WechatOfficialVoiceMessage{
			MediaID: message.MediaURL,
		}
	case "video":
		sendMsg.Video = &WechatOfficialVideoMessage{
			MediaID: message.MediaURL,
			Title:   message.Content,
		}
	case "news":
		// 图文消息处理
		if articles, ok := message.RawContent["articles"].([]interface{}); ok {
			news := &WechatOfficialNewsMessage{
				Articles: make([]WechatOfficialArticle, 0, len(articles)),
			}
			for _, article := range articles {
				if articleMap, ok := article.(map[string]interface{}); ok {
					news.Articles = append(news.Articles, WechatOfficialArticle{
						Title:       getString(articleMap, "title"),
						Description: getString(articleMap, "description"),
						URL:         getString(articleMap, "url"),
						PicURL:      getString(articleMap, "pic_url"),
					})
				}
			}
			sendMsg.News = news
		}
	default:
		// 默认作为文本消息处理
		sendMsg.Text = &WechatOfficialTextMessage{
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
	reqURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s", accessToken)

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

	var sendResp WechatOfficialResponse
	if err := json.Unmarshal(body, &sendResp); err != nil {
		return fmt.Errorf("failed to parse send response: %w", err)
	}

	if sendResp.ErrCode != 0 {
		return fmt.Errorf("wechat official API error: %d - %s", sendResp.ErrCode, sendResp.ErrMsg)
	}

	return nil
}

// GetAccessToken 获取访问令牌
func (w *WechatOfficialAdapter) GetAccessToken(ctx context.Context, config map[string]interface{}) (*dto.AccessTokenResponse, error) {
	appID, ok := config["app_id"].(string)
	if !ok || appID == "" {
		return nil, fmt.Errorf("app_id is required in wechat official config")
	}
	appSecret, ok := config["app_secret"].(string)
	if !ok || appSecret == "" {
		return nil, fmt.Errorf("app_secret is required in wechat official config")
	}

	reqURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)

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
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp WechatOfficialTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("wechat official API error: %d - %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &dto.AccessTokenResponse{
		AccessToken: tokenResp.AccessToken,
		ExpiresIn:   tokenResp.ExpiresIn,
		ExpiresAt:   &expiresAt,
	}, nil
}

// RefreshAccessToken 刷新访问令牌
func (w *WechatOfficialAdapter) RefreshAccessToken(ctx context.Context, config map[string]interface{}, oldToken string) (*dto.AccessTokenResponse, error) {
	// 微信公众号的access_token刷新就是重新获取
	return w.GetAccessToken(ctx, config)
}

// VerifyWebhook 验证Webhook请求
func (w *WechatOfficialAdapter) VerifyWebhook(ctx context.Context, signature string, timestamp string, nonce string, body []byte, config map[string]interface{}) error {
	token, ok := config["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("token is required in wechat official config for webhook verification")
	}

	// 微信公众号签名验证
	tmpArr := []string{token, timestamp, nonce}
	sort.Strings(tmpArr)
	tmpStr := strings.Join(tmpArr, "")

	hash := sha1.New()
	hash.Write([]byte(tmpStr))
	expectedSignature := fmt.Sprintf("%x", hash.Sum(nil))

	if signature != expectedSignature {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// ParseInboundMessage 解析入站消息
func (w *WechatOfficialAdapter) ParseInboundMessage(ctx context.Context, body []byte, config map[string]interface{}) (*dto.UnifiedMessage, error) {
	var wechatMsg WechatOfficialMessage
	if err := json.Unmarshal(body, &wechatMsg); err != nil {
		return nil, fmt.Errorf("failed to parse wechat official message: %w", err)
	}

	// 构建统一消息格式
	unifiedMsg := &dto.UnifiedMessage{
		MessageID:         fmt.Sprintf("wechat_official_%d", wechatMsg.MsgID),
		PlatformMessageID: fmt.Sprintf("%d", wechatMsg.MsgID),
		MessageType:       mapFromWechatOfficialMessageType(wechatMsg.MsgType),
		SenderID:          wechatMsg.FromUserName,
		SenderName:        wechatMsg.FromUserName,
		SenderType:        "user",
		ReceiverID:        wechatMsg.ToUserName,
		ReceiverName:      wechatMsg.ToUserName,
		ReceiverType:      "official_account",
		Content:           wechatMsg.Content,
		RawContent: map[string]interface{}{
			"original_message": wechatMsg,
		},
		PlatformTimestamp: time.Unix(wechatMsg.CreateTime, 0),
	}

	// 根据消息类型设置额外信息
	switch wechatMsg.MsgType {
	case "image":
		unifiedMsg.MediaURL = wechatMsg.PicURL
		unifiedMsg.MediaType = "image/jpeg"
	case "voice":
		unifiedMsg.MediaURL = wechatMsg.MediaID
		unifiedMsg.MediaType = "audio/" + wechatMsg.Format
	case "video":
		unifiedMsg.MediaURL = wechatMsg.MediaID
		unifiedMsg.MediaType = "video/mp4"
	case "location":
		unifiedMsg.Content = fmt.Sprintf("位置: %s (%s, %s)", wechatMsg.Label, wechatMsg.LocationX, wechatMsg.LocationY)
	case "link":
		unifiedMsg.Content = fmt.Sprintf("%s\n%s\n%s", wechatMsg.Title, wechatMsg.Description, wechatMsg.URL)
	case "event":
		unifiedMsg.MessageType = "event"
		unifiedMsg.Content = fmt.Sprintf("事件: %s", wechatMsg.Event)
		if wechatMsg.EventKey != "" {
			unifiedMsg.Content += fmt.Sprintf(" - %s", wechatMsg.EventKey)
		}
	}

	return unifiedMsg, nil
}

// BuildWebhookPath 构建Webhook路径
func (w *WechatOfficialAdapter) BuildWebhookPath(channelID int64) string {
	return fmt.Sprintf("/webhook/wechat_official/%d", channelID)
}

// 辅助函数

func mapToWechatOfficialMessageType(msgType string) string {
	switch msgType {
	case "text":
		return "text"
	case "image":
		return "image"
	case "voice":
		return "voice"
	case "video":
		return "video"
	case "news":
		return "news"
	default:
		return "text"
	}
}

func mapFromWechatOfficialMessageType(msgType string) string {
	switch msgType {
	case "text":
		return "text"
	case "image":
		return "image"
	case "voice":
		return "voice"
	case "video":
		return "video"
	case "location":
		return "location"
	case "link":
		return "link"
	case "event":
		return "event"
	default:
		return "text"
	}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// 确保 WechatOfficialAdapter 实现了所需的接口
var _ port.MessageSender = (*WechatOfficialAdapter)(nil)
var _ port.TokenManager = (*WechatOfficialAdapter)(nil)
var _ port.WebhookProcessor = (*WechatOfficialAdapter)(nil)
var _ port.PlatformIdentifier = (*WechatOfficialAdapter)(nil)
var _ port.ConfigValidator = (*WechatOfficialAdapter)(nil)
