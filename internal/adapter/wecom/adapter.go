package wecom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
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
	case "markdown":
		sendMsg.Markdown = &WecomMarkdownMessage{
			Content: message.Content,
		}
	case "image":
		sendMsg.Image = &WecomImageMessage{
			MediaID: message.MediaURL,
		}
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

	var sendResp WecomResponse
	if err := json.Unmarshal(body, &sendResp); err != nil {
		return fmt.Errorf("failed to parse send response: %w", err)
	}

	if sendResp.ErrCode != 0 {
		return fmt.Errorf("wecom API error: %d - %s", sendResp.ErrCode, sendResp.ErrMsg)
	}

	return nil
}

// mapToWecomMessageType 映射消息类型到企业微信格式
func mapToWecomMessageType(msgType string) string {
	switch msgType {
	case "text":
		return "text"
	case "markdown":
		return "markdown"
	case "image":
		return "image"
	case "voice":
		return "voice"
	case "video":
		return "video"
	case "file":
		return "file"
	default:
		return "text"
	}
}

// ParseWecomConfig 解析企业微信配置
func ParseWecomConfig(config map[string]interface{}) (*WecomConfig, error) {
	corpID, ok := config["corp_id"].(string)
	if !ok || corpID == "" {
		return nil, fmt.Errorf("corp_id is required")
	}

	agentID, ok := config["agent_id"].(string)
	if !ok || agentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}

	secret, ok := config["secret"].(string)
	if !ok || secret == "" {
		return nil, fmt.Errorf("secret is required")
	}

	token, _ := config["token"].(string)

	return &WecomConfig{
		CorpID:  corpID,
		AgentID: agentID,
		Secret:  secret,
		Token:   token,
	}, nil
}

// WecomConfig 企业微信配置
type WecomConfig struct {
	CorpID  string `json:"corp_id"`
	AgentID string `json:"agent_id"`
	Secret  string `json:"secret"`
	Token   string `json:"token"`
}

// 确保 WecomAdapter 实现了所需的接口
var _ port.MessageSender = (*WecomAdapter)(nil)
var _ port.PlatformIdentifier = (*WecomAdapter)(nil)
var _ port.ConfigValidator = (*WecomAdapter)(nil)
