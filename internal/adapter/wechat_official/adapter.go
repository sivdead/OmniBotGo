package wechat_official

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

var (
	errWechatOfficialMissingAppID     = errors.New("app_id is required in wechat official config")
	errWechatOfficialMissingAppSecret = errors.New("app_secret is required in wechat official config")
	errWechatOfficialMissingToken     = errors.New("token is required in wechat official config for webhook verification")
	errWechatOfficialInvalidSignature = errors.New("invalid signature")
	errWechatOfficialHTTPStatus       = errors.New("wechat official API HTTP error")
)

const (
	wechatOfficialMsgTypeVoice     = "voice"
	wechatOfficialMsgTypeLocation  = "location"
	wechatOfficialMsgTypeEvent     = "event"
	wechatOfficialEventSubscribe   = "subscribe"
	wechatOfficialEventUnsubscribe = "unsubscribe"
	rawKeyLocation                 = "location"
	rawKeyEvent                    = "event"
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

// SupportedMessageTypes implements port.MessageTypeSupport (T08).
func (w *WechatOfficialAdapter) SupportedMessageTypes() []string {
	return []string{
		entity.MessageTypeText,
		entity.MessageTypeImage,
		entity.MessageTypeAudio,
		entity.MessageTypeVideo,
		entity.MessageTypeNews,
		wechatOfficialMsgTypeVoice,
	}
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
	case entity.MessageTypeText:
		sendMsg.Text = &WechatOfficialTextMessage{
			Content: message.Content,
		}
	case entity.MessageTypeImage:
		sendMsg.Image = &WechatOfficialImageMessage{
			MediaID: message.MediaURL,
		}
	case entity.MessageTypeAudio, wechatOfficialMsgTypeVoice:
		// D05: entity.MessageTypeAudio → voice/media_id (not text fallback).
		sendMsg.Voice = &WechatOfficialVoiceMessage{
			MediaID: message.MediaURL,
		}
		sendMsg.MsgType = wechatOfficialMsgTypeVoice
	case entity.MessageTypeVideo:
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
		// T08: do not silently downgrade unknown types to text
		return fmt.Errorf("%w: wechat_official cannot send %q", port.ErrUnsupportedMessageType, message.MessageType)
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

	// D06: reject non-2xx before treating as sent.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: status %d", errWechatOfficialHTTPStatus, resp.StatusCode)
	}

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
		return nil, errWechatOfficialMissingAppID
	}

	appSecret, ok := config["app_secret"].(string)
	if !ok || appSecret == "" {
		return nil, errWechatOfficialMissingAppSecret
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
func (w *WechatOfficialAdapter) VerifyWebhook(_ context.Context, signature, timestamp, nonce string, _ []byte, cfg map[string]interface{}) error {
	token, ok := cfg["token"].(string)
	if !ok || token == "" {
		return errWechatOfficialMissingToken
	}

	// 微信公众号签名验证
	tmpArr := []string{token, timestamp, nonce}
	sort.Strings(tmpArr)
	tmpStr := strings.Join(tmpArr, "")

	hash := sha1.New()
	hash.Write([]byte(tmpStr))
	expectedSignature := fmt.Sprintf("%x", hash.Sum(nil))

	if signature != expectedSignature {
		return errWechatOfficialInvalidSignature
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
	case wechatOfficialMsgTypeLocation:
		applyWechatOfficialLocation(unifiedMsg, &wechatMsg)
	case "link":
		unifiedMsg.Content = fmt.Sprintf("%s\n%s\n%s", wechatMsg.Title, wechatMsg.Description, wechatMsg.URL)
	case wechatOfficialMsgTypeEvent:
		applyWechatOfficialEvent(unifiedMsg, &wechatMsg)
	}

	return unifiedMsg, nil
}

// BuildWebhookPath 构建Webhook路径
func (w *WechatOfficialAdapter) BuildWebhookPath(channelID string) string {
	return fmt.Sprintf("/webhook/wechat_official/%s", channelID)
}

// 辅助函数

func mapToWechatOfficialMessageType(msgType string) string {
	switch msgType {
	case entity.MessageTypeText:
		return "text"
	case entity.MessageTypeImage:
		return "image"
	case entity.MessageTypeAudio, wechatOfficialMsgTypeVoice:
		return wechatOfficialMsgTypeVoice
	case entity.MessageTypeVideo:
		return "video"
	case entity.MessageTypeNews:
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
	case wechatOfficialMsgTypeLocation:
		return wechatOfficialMsgTypeLocation
	case "link":
		return "link"
	case wechatOfficialMsgTypeEvent:
		return wechatOfficialMsgTypeEvent
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

func applyWechatOfficialLocation(unifiedMsg *dto.UnifiedMessage, wechatMsg *WechatOfficialMessage) {
	unifiedMsg.MessageType = entity.MessageTypeLocation

	latitude, errLat := strconv.ParseFloat(wechatMsg.LocationX, 64)
	if errLat != nil {
		latitude = 0
	}

	longitude, errLng := strconv.ParseFloat(wechatMsg.LocationY, 64)
	if errLng != nil {
		longitude = 0
	}

	unifiedMsg.LocationContent = &entity.LocationMessage{
		Latitude:  latitude,
		Longitude: longitude,
		Address:   wechatMsg.Label,
	}
	unifiedMsg.Content = wechatMsg.Label
	unifiedMsg.RawContent[rawKeyLocation] = map[string]interface{}{
		"latitude":  latitude,
		"longitude": longitude,
		"address":   wechatMsg.Label,
	}
}

func applyWechatOfficialEvent(unifiedMsg *dto.UnifiedMessage, wechatMsg *WechatOfficialMessage) {
	unifiedMsg.MessageType = entity.MessageTypeEvent
	eventType := mapWechatOfficialEventType(wechatMsg.Event)
	unifiedMsg.EventContent = &entity.EventMessage{
		EventType: eventType,
		EventKey:  wechatMsg.EventKey,
		EventData: map[string]interface{}{
			"original_event": wechatMsg.Event,
		},
	}
	unifiedMsg.Content = wechatMsg.Event

	if wechatMsg.EventKey != "" {
		unifiedMsg.Content += " - " + wechatMsg.EventKey
	}

	unifiedMsg.RawContent[rawKeyEvent] = map[string]interface{}{
		"event_type": eventType,
		"event_key":  wechatMsg.EventKey,
	}
}

func mapWechatOfficialEventType(event string) string {
	switch strings.ToLower(event) {
	case wechatOfficialEventSubscribe:
		return entity.EventTypeUserSubscribe
	case wechatOfficialEventUnsubscribe:
		return entity.EventTypeUserUnsubscribe
	case "click":
		return entity.EventTypeMenuClick
	case "view":
		return entity.EventTypeMenuView
	case wechatOfficialMsgTypeLocation:
		return entity.EventTypeLocationReport
	case "scancode_push", "scancode_waitmsg", "scan":
		return entity.EventTypeScanQRCode
	case "location_select":
		return entity.EventTypeLocationSelect
	default:
		return entity.EventTypeCustom
	}
}

// 确保 WechatOfficialAdapter 实现了所需的接口。
var (
	_ port.MessageSender      = (*WechatOfficialAdapter)(nil)
	_ port.TokenManager       = (*WechatOfficialAdapter)(nil)
	_ port.WebhookProcessor   = (*WechatOfficialAdapter)(nil)
	_ port.PlatformIdentifier = (*WechatOfficialAdapter)(nil)
	_ port.ConfigValidator    = (*WechatOfficialAdapter)(nil)
)
