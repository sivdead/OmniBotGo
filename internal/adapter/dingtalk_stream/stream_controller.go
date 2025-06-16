package dingtalk_stream

import (
	"context"
	"fmt"
	"time"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/entity"
)

// MessageHandler 消息处理接口，用于解决循环依赖
type MessageHandler interface {
	ProcessInboundMessage(ctx context.Context, message *entity.Message) error
}

// StreamController 钉钉Stream控制器
type StreamController struct {
	messageHandler MessageHandler
	logger         zerolog.Logger
	client         *client.StreamClient
	config         *DingtalkStreamConfig
	channelID      int64
	isRunning      bool
}

// NewStreamController 创建钉钉Stream控制器
func NewStreamController(
	messageHandler MessageHandler,
	logger zerolog.Logger,
	config *DingtalkStreamConfig,
	channelID int64,
) *StreamController {
	return &StreamController{
		messageHandler: messageHandler,
		logger:         logger,
		config:         config,
		channelID:      channelID,
		isRunning:      false,
	}
}

// Start 启动钉钉Stream客户端
func (c *StreamController) Start(ctx context.Context) error {
	if c.isRunning {
		return fmt.Errorf("dingtalk stream controller is already running")
	}

	// 设置钉钉SDK日志级别
	logger.SetLogger(logger.NewStdTestLogger())

	// 创建Stream客户端
	streamClient := client.NewStreamClient(client.WithAppCredential(client.NewAppCredentialConfig(c.config.ClientID, c.config.ClientSecret)))

	// 注册机器人消息处理器
	streamClient.RegisterChatBotCallbackRouter(c.handleChatBotMessage)

	// 启动客户端
	c.client = streamClient
	c.isRunning = true

	c.logger.Info().Msg("dingtalk stream controller started")

	// 在goroutine中运行客户端
	go func() {
		defer func() {
			c.isRunning = false
			c.logger.Info().Msg("dingtalk stream controller stopped")
		}()

		err := streamClient.Start(ctx)
		if err != nil {
			c.logger.Error().Err(err).Msg("dingtalk stream client error")
		}
	}()

	return nil
}

// Stop 停止钉钉Stream客户端
func (c *StreamController) Stop() error {
	if !c.isRunning {
		return nil
	}

	if c.client != nil {
		c.client.Close()
	}

	c.isRunning = false
	c.logger.Info().Msg("dingtalk stream controller stopping")

	return nil
}

// IsRunning 检查是否正在运行
func (c *StreamController) IsRunning() bool {
	return c.isRunning
}

// handleChatBotMessage 处理机器人消息
func (c *StreamController) handleChatBotMessage(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	c.logger.Info().
		Str("conversation_id", data.ConversationId).
		Str("sender_id", data.SenderStaffId).
		Str("message_type", data.Msgtype).
		Msg("received dingtalk message")

	// 直接转换为统一消息格式
	unifiedMsg, err := c.convertToUnifiedMessage(data)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to convert dingtalk message")
		return []byte(`{"text": "消息转换失败"}`), nil
	}

	// 创建消息实体
	message := &entity.Message{
		MessageID:         unifiedMsg.MessageID,
		ChannelID:         c.channelID,
		PlatformMessageID: unifiedMsg.PlatformMessageID,
		Direction:         entity.MessageDirectionInbound,
		MessageType:       unifiedMsg.MessageType,
		ContentType:       unifiedMsg.MediaType,
		SenderID:          unifiedMsg.SenderID,
		SenderName:        unifiedMsg.SenderName,
		SenderType:        unifiedMsg.SenderType,
		ReceiverID:        unifiedMsg.ReceiverID,
		ReceiverName:      unifiedMsg.ReceiverName,
		ReceiverType:      unifiedMsg.ReceiverType,
		Content:           unifiedMsg.Content,
		RawContent:        entity.JSONField(unifiedMsg.RawContent),
		MediaURL:          unifiedMsg.MediaURL,
		MediaType:         unifiedMsg.MediaType,
		FileSize:          unifiedMsg.FileSize,
		MessageStatus:     entity.MessageStatusPending,
		ConversationID:    unifiedMsg.ConversationID,
		PlatformTimestamp: unifiedMsg.PlatformTimestamp,
		ReceivedAt:        time.Now(),
	}

	// 处理消息
	err = c.messageHandler.ProcessInboundMessage(ctx, message)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to process inbound message")
		// 返回错误响应
		return []byte(`{"text": "消息处理失败，请稍后重试"}`), nil
	}

	// 这里可以根据业务逻辑决定是否返回响应
	// 目前先返回一个简单的确认消息
	if message.MessageType == "text" {
		response := fmt.Sprintf(`{"text": "收到消息：%s"}`, message.Content)
		return []byte(response), nil
	}

	return nil, nil
}

// convertToUnifiedMessage 直接转换钉钉数据模型为统一消息格式
func (c *StreamController) convertToUnifiedMessage(data *chatbot.BotCallbackDataModel) (*entity.UnifiedMessage, error) {
	unifiedMsg := &entity.UnifiedMessage{
		MessageID:         data.MsgId,
		MessageType:       data.Msgtype,
		SenderID:          data.SenderId,
		SenderName:        data.SenderNick,
		SenderType:        "user",
		ConversationID:    data.ConversationId,
		PlatformMessageID: data.MsgId,
		PlatformTimestamp: time.Unix(data.CreateAt/1000, 0),
		RawContent:        make(map[string]interface{}),
	}

	// 设置接收者信息
	if data.ConversationType == "1" { // 单聊
		unifiedMsg.ReceiverID = data.ChatbotUserId
		unifiedMsg.ReceiverType = "bot"
	} else { // 群聊
		unifiedMsg.ReceiverID = data.ConversationId
		unifiedMsg.ReceiverType = "group"
	}

	// 根据消息类型设置内容
	switch data.Msgtype {
	case "text":
		if data.Text.Content != "" {
			unifiedMsg.Content = data.Text.Content
		}
	case "picture":
		unifiedMsg.Content = "[图片]"
		unifiedMsg.MediaType = "image"
		if contentData, ok := data.Content.(map[string]interface{}); ok {
			unifiedMsg.RawContent["downloadCode"] = getStringFromMap(contentData, "downloadCode")
			unifiedMsg.RawContent["pictureType"] = getStringFromMap(contentData, "pictureType")
		}
	case "audio":
		unifiedMsg.Content = "[语音]"
		unifiedMsg.MediaType = "audio"
		if contentData, ok := data.Content.(map[string]interface{}); ok {
			unifiedMsg.RawContent["downloadCode"] = getStringFromMap(contentData, "downloadCode")
			unifiedMsg.RawContent["duration"] = getInt64FromMap(contentData, "duration")
			unifiedMsg.RawContent["recognition"] = getStringFromMap(contentData, "recognition")
		}
	case "video":
		unifiedMsg.Content = "[视频]"
		unifiedMsg.MediaType = "video"
		if contentData, ok := data.Content.(map[string]interface{}); ok {
			unifiedMsg.RawContent["downloadCode"] = getStringFromMap(contentData, "downloadCode")
			unifiedMsg.RawContent["duration"] = getInt64FromMap(contentData, "duration")
			unifiedMsg.RawContent["videoType"] = getStringFromMap(contentData, "videoType")
		}
	case "file":
		unifiedMsg.Content = "[文件]"
		unifiedMsg.MediaType = "file"
		if contentData, ok := data.Content.(map[string]interface{}); ok {
			fileName := getStringFromMap(contentData, "fileName")
			if fileName != "" {
				unifiedMsg.Content = "[文件] " + fileName
			}
			unifiedMsg.RawContent["downloadCode"] = getStringFromMap(contentData, "downloadCode")
			unifiedMsg.RawContent["fileName"] = fileName
			unifiedMsg.RawContent["fileType"] = getStringFromMap(contentData, "fileType")
		}
	case "interactive":
		if contentData, ok := data.Content.(map[string]interface{}); ok {
			title := getStringFromMap(contentData, "title")
			if title != "" {
				unifiedMsg.Content = title
			} else {
				unifiedMsg.Content = "[交互式消息]"
			}
			unifiedMsg.RawContent["actionCardId"] = getStringFromMap(contentData, "actionCardId")
			unifiedMsg.RawContent["title"] = title
			unifiedMsg.RawContent["content"] = getStringFromMap(contentData, "content")
		}
	default:
		unifiedMsg.Content = "[未知类型消息]"
	}

	// 存储原始消息数据
	unifiedMsg.RawContent["conversationType"] = data.ConversationType
	unifiedMsg.RawContent["chatbotCorpId"] = data.ChatbotCorpId
	unifiedMsg.RawContent["senderCorpId"] = data.SenderCorpId
	unifiedMsg.RawContent["senderStaffId"] = data.SenderStaffId
	unifiedMsg.RawContent["sessionWebhook"] = data.SessionWebhook
	unifiedMsg.RawContent["sessionWebhookExpiredTime"] = data.SessionWebhookExpiredTime
	unifiedMsg.RawContent["isAdmin"] = data.IsAdmin
	unifiedMsg.RawContent["isInAtList"] = data.IsInAtList
	unifiedMsg.RawContent["conversationTitle"] = data.ConversationTitle

	// 处理@用户信息
	if len(data.AtUsers) > 0 {
		atUsers := make([]map[string]interface{}, len(data.AtUsers))
		for i, atUser := range data.AtUsers {
			atUsers[i] = map[string]interface{}{
				"dingtalkId": atUser.DingtalkId,
				"staffId":    atUser.StaffId,
			}
		}
		unifiedMsg.RawContent["atUsers"] = atUsers
	}

	return unifiedMsg, nil
}

// 辅助函数：从map中安全获取string值
func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// 辅助函数：从map中安全获取int64值
func getInt64FromMap(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int64:
			return val
		case int:
			return int64(val)
		case float64:
			return int64(val)
		}
	}
	return 0
}

// GetPlatformType 获取平台类型
func (c *StreamController) GetPlatformType() entity.PlatformType {
	return entity.PlatformTypeDingtalk
}

// ValidateConfig 验证配置
func (c *StreamController) ValidateConfig(config map[string]interface{}) error {
	if clientID, ok := config["client_id"].(string); !ok || clientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if clientSecret, ok := config["client_secret"].(string); !ok || clientSecret == "" {
		return fmt.Errorf("client_secret is required")
	}
	return nil
}
