package wecom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/dto"
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
func (w *WecomAdapter) ValidateConfig(cfg map[string]interface{}) error {
	return config.ValidatePlatformConfig(entity.PlatformTypeWecom, cfg)
}

// SendMessage 发送消息
func (w *WecomAdapter) SendMessage(ctx context.Context, message *dto.UnifiedMessage, conf map[string]interface{}, accessToken string) error {
	agentID := conf["agent_id"].(string)

	// 构建企业微信消息格式
	sendMsg := WecomSendMessage{
		ToUser:  message.ReceiverID,
		MsgType: mapToWecomMessageType(message.MessageType),
		AgentID: agentID,
	}

	// 根据消息类型设置内容
	switch message.MessageType {
	case entity.MessageTypeText:
		sendMsg.Text = &WecomTextMessage{
			Content: message.Content,
		}

	case entity.MessageTypeMarkdown:
		// 优先使用结构化内容
		if message.MarkdownContent != nil {
			sendMsg.Markdown = &WecomMarkdownMessage{
				Content: message.MarkdownContent.Content,
			}
		} else {
			sendMsg.Markdown = &WecomMarkdownMessage{
				Content: message.Content,
			}
		}

	case entity.MessageTypeImage:
		sendMsg.Image = &WecomImageMessage{
			MediaID: message.MediaURL,
		}

	case entity.MessageTypeVideo:
		sendMsg.Video = &WecomVideoMessage{
			MediaID: message.MediaURL,
		}

	case entity.MessageTypeFile:
		// 优先使用结构化内容
		if message.FileContent != nil {
			sendMsg.File = &WecomFileMessage{
				MediaID: message.FileContent.FileURL,
			}
		} else {
			sendMsg.File = &WecomFileMessage{
				MediaID: message.MediaURL,
			}
		}

	case entity.MessageTypeCard:
		// 转换为文本卡片消息
		if message.CardContent != nil {
			sendMsg.MsgType = "textcard"
			sendMsg.TextCard = &WecomTextCardMessage{
				Title:       message.CardContent.Title,
				Description: message.CardContent.Content,
			}
			// 如果有按钮，使用第一个按钮的URL
			if len(message.CardContent.Buttons) > 0 {
				sendMsg.TextCard.URL = message.CardContent.Buttons[0].ActionURL
				sendMsg.TextCard.BtnTxt = message.CardContent.Buttons[0].Title
			}
		}

	case entity.MessageTypeNews:
		// 图文消息
		if message.NewsContent != nil {
			sendMsg.MsgType = "news"
			articles := make([]WecomArticle, 0, len(message.NewsContent.Articles))
			for _, article := range message.NewsContent.Articles {
				articles = append(articles, WecomArticle{
					Title:       article.Title,
					Description: article.Description,
					URL:         article.URL,
					PicURL:      article.PicURL,
				})
			}
			sendMsg.News = &WecomNewsMessage{
				Articles: articles,
			}
		}

	case entity.MessageTypeTemplate:
		// 模板卡片消息
		if message.TemplateContent != nil {
			sendMsg.MsgType = "template_card"
			// 这里可以根据模板内容构建更复杂的模板卡片
			sendMsg.TemplateCard = &WecomTemplateCard{
				CardType: "text_notice",
				MainTitle: &WecomTemplateCardMainTitle{
					Title: "通知",
				},
			}
		}

	case entity.MessageTypeEvent:
		// 事件消息通常不发送，而是接收
		return fmt.Errorf("不支持发送事件类型消息")

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
	case entity.MessageTypeText:
		return "text"
	case entity.MessageTypeMarkdown:
		return "markdown"
	case entity.MessageTypeImage:
		return "image"
	case entity.MessageTypeAudio:
		return "voice"
	case entity.MessageTypeVideo:
		return "video"
	case entity.MessageTypeFile:
		return "file"
	case entity.MessageTypeCard:
		return "textcard"
	case entity.MessageTypeNews:
		return "news"
	default:
		return "text"
	}
}

// 确保 WecomAdapter 实现了所需的接口
var _ port.MessageSender = (*WecomAdapter)(nil)
var _ port.PlatformIdentifier = (*WecomAdapter)(nil)
var _ port.ConfigValidator = (*WecomAdapter)(nil)
