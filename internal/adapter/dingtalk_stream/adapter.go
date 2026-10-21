package dingtalk_stream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

var errStreamMissingChannelID = errors.New("channel_id is required in dingtalk stream config")

// DingtalkStreamAdapter 钉钉Stream适配器，实现MessageSender, StreamAdapter, PlatformIdentifier, ConfigValidator接口
type DingtalkStreamAdapter struct {
	logger           zerolog.Logger
	streamController *StreamController
	config           map[string]interface{}
	isConnected      bool
	mu               sync.RWMutex
}

// NewAdapter 创建钉钉Stream适配器
func NewAdapter(logger zerolog.Logger) *DingtalkStreamAdapter {
	return &DingtalkStreamAdapter{
		logger:      logger,
		isConnected: false,
	}
}

// NewStreamSession 实现 port.StreamSessionFactory（C06：每通道独立会话）。
func (a *DingtalkStreamAdapter) NewStreamSession() port.StreamAdapter {
	return NewAdapter(a.logger)
}

// GetPlatformType 实现PlatformIdentifier接口
func (a *DingtalkStreamAdapter) GetPlatformType() entity.PlatformType {
	return entity.PlatformTypeDingtalk
}

// ValidateConfig 实现ConfigValidator接口
func (a *DingtalkStreamAdapter) ValidateConfig(cfg map[string]interface{}) error {
	return config.ValidatePlatformConfig(entity.PlatformTypeDingtalk, cfg)
}

// SupportedMessageTypes implements port.MessageTypeSupport (T08).
func (a *DingtalkStreamAdapter) SupportedMessageTypes() []string {
	return []string{
		entity.MessageTypeText,
		entity.MessageTypeMarkdown,
		entity.MessageTypeCard,
		entity.MessageTypeLink,
	}
}

// SendMessage 实现MessageSender接口
func (a *DingtalkStreamAdapter) SendMessage(ctx context.Context, message *dto.UnifiedMessage, config map[string]interface{}, accessToken string) error {
	// 从统一消息体中获取sessionWebhook
	sessionWebhook, ok := message.RawContent["sessionWebhook"].(string)
	if !ok || sessionWebhook == "" {
		return fmt.Errorf("sessionWebhook not found in message RawContent for dingtalk stream mode")
	}

	// 根据消息类型构建请求体
	var requestBody interface{}

	switch message.MessageType {
	case entity.MessageTypeText:
		requestBody = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]interface{}{
				"content": message.Content,
			},
		}

	case entity.MessageTypeMarkdown:
		title := "消息"
		// 优先使用结构化内容
		if message.MarkdownContent != nil {
			title = message.MarkdownContent.Title
			if title == "" {
				title = "Markdown消息"
			}
			requestBody = map[string]interface{}{
				"msgtype": "markdown",
				"markdown": map[string]interface{}{
					"title": title,
					"text":  message.MarkdownContent.Content,
				},
			}
		} else {
			if titleValue, ok := message.RawContent["title"].(string); ok {
				title = titleValue
			}
			requestBody = map[string]interface{}{
				"msgtype": "markdown",
				"markdown": map[string]interface{}{
					"title": title,
					"text":  message.Content,
				},
			}
		}

	case entity.MessageTypeCard:
		// 转换为ActionCard
		if message.CardContent != nil {
			actionCard := map[string]interface{}{
				"msgtype": "actionCard",
				"actionCard": map[string]interface{}{
					"title": message.CardContent.Title,
					"text":  message.CardContent.Content,
				},
			}

			// 处理按钮
			if len(message.CardContent.Buttons) == 1 {
				// 单个按钮
				actionCard["actionCard"].(map[string]interface{})["singleTitle"] = message.CardContent.Buttons[0].Title
				actionCard["actionCard"].(map[string]interface{})["singleURL"] = message.CardContent.Buttons[0].ActionURL
			} else if len(message.CardContent.Buttons) > 1 {
				// 多个按钮
				btns := make([]map[string]interface{}, 0, len(message.CardContent.Buttons))
				for _, btn := range message.CardContent.Buttons {
					btns = append(btns, map[string]interface{}{
						"title":     btn.Title,
						"actionURL": btn.ActionURL,
					})
				}
				actionCard["actionCard"].(map[string]interface{})["btns"] = btns
				actionCard["actionCard"].(map[string]interface{})["btnOrientation"] = "0" // 竖向排列
			}

			requestBody = actionCard
		} else {
			// 兼容旧格式
			requestBody = a.buildLegacyActionCard(message)
		}

	case entity.MessageTypeLink:
		// 链接消息
		linkData := map[string]interface{}{
			"msgtype": "link",
			"link": map[string]interface{}{
				"title":      message.Content,
				"text":       "",
				"messageUrl": "",
				"picUrl":     "",
			},
		}

		// 从RawContent中提取链接信息
		if title, ok := message.RawContent["title"].(string); ok {
			linkData["link"].(map[string]interface{})["title"] = title
		}
		if text, ok := message.RawContent["text"].(string); ok {
			linkData["link"].(map[string]interface{})["text"] = text
		}
		if messageUrl, ok := message.RawContent["messageUrl"].(string); ok {
			linkData["link"].(map[string]interface{})["messageUrl"] = messageUrl
		}
		if picUrl, ok := message.RawContent["picUrl"].(string); ok {
			linkData["link"].(map[string]interface{})["picUrl"] = picUrl
		}

		requestBody = linkData

	case entity.MessageTypeFile, entity.MessageTypeImage:
		// T08: stream mode cannot send binary media via sessionWebhook — do not fake as text
		return fmt.Errorf("%w: dingtalk stream cannot send %q", port.ErrUnsupportedMessageType, message.MessageType)

	default:
		return fmt.Errorf("%w: dingtalk stream cannot send %q", port.ErrUnsupportedMessageType, message.MessageType)
	}

	// 序列化请求体
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 发送HTTP请求到sessionWebhook
	req, err := http.NewRequestWithContext(ctx, "POST", sessionWebhook, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	var sendResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&sendResp); err != nil {
		return fmt.Errorf("failed to decode send response: %w", err)
	}

	if sendResp.ErrCode != 0 {
		return fmt.Errorf("dingtalk send message error: %s (code: %d)", sendResp.ErrMsg, sendResp.ErrCode)
	}

	a.logger.Info().
		Str("message_id", message.MessageID).
		Str("message_type", message.MessageType).
		Msg("dingtalk message sent successfully")

	return nil
}

// buildLegacyActionCard 构建旧格式的ActionCard（向后兼容）
func (a *DingtalkStreamAdapter) buildLegacyActionCard(message *dto.UnifiedMessage) map[string]interface{} {
	actionCard := map[string]interface{}{
		"msgtype": "actionCard",
		"actionCard": map[string]interface{}{
			"title": message.RawContent["title"],
			"text":  message.Content,
		},
	}

	// 添加可选字段
	if singleTitle, ok := message.RawContent["singleTitle"].(string); ok {
		actionCard["actionCard"].(map[string]interface{})["singleTitle"] = singleTitle
	}
	if singleURL, ok := message.RawContent["singleURL"].(string); ok {
		actionCard["actionCard"].(map[string]interface{})["singleURL"] = singleURL
	}
	if btnOrientation, ok := message.RawContent["btnOrientation"].(string); ok {
		actionCard["actionCard"].(map[string]interface{})["btnOrientation"] = btnOrientation
	}
	if btns, ok := message.RawContent["btns"].([]interface{}); ok {
		actionCard["actionCard"].(map[string]interface{})["btns"] = btns
	}

	return actionCard
}

// Start 实现StreamAdapter接口 - 启动Stream连接
func (a *DingtalkStreamAdapter) Start(ctx context.Context, messageHandler port.MessageHandler, cfg map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isConnected {
		return fmt.Errorf("dingtalk stream adapter is already connected")
	}

	// 保存配置
	a.config = cfg

	// 验证配置
	if err := a.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// 解析配置
	streamConfig, err := config.ParseDingtalkStreamConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to parse dingtalk stream config: %w", err)
	}

	// 从配置中获取channelID，channel_id为必填项
	channelID, ok := cfg["channel_id"].(string)
	if !ok || channelID == "" {
		return errStreamMissingChannelID
	}

	// T02: pass port.MessageHandler directly (no entity round-trip adapter)
	a.streamController = NewStreamController(
		messageHandler,
		a.logger,
		streamConfig,
		channelID,
	)

	// 启动Stream控制器
	if err := a.streamController.Start(ctx); err != nil {
		return fmt.Errorf("failed to start stream controller: %w", err)
	}

	a.isConnected = true

	a.logger.Info().
		Str("client_id", streamConfig.ClientID).
		Msg("dingtalk stream adapter started")

	return nil
}

// Stop 实现StreamAdapter接口 - 停止Stream连接
func (a *DingtalkStreamAdapter) Stop(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isConnected {
		return nil
	}

	if a.streamController != nil {
		if err := a.streamController.Stop(); err != nil {
			a.logger.Error().Err(err).Msg("error stopping stream controller")
		}
	}

	a.isConnected = false
	a.streamController = nil

	a.logger.Info().Msg("dingtalk stream adapter stopped")

	return nil
}

// IsConnected 实现StreamAdapter接口 - 检查连接状态
func (a *DingtalkStreamAdapter) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.isConnected && a.streamController != nil && a.streamController.IsRunning()
}
