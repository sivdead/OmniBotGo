package dingtalk_stream

import (
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

// DingtalkMessageType 钉钉消息类型
type DingtalkMessageType string

const (
	DingtalkMessageTypeText        DingtalkMessageType = "text"
	DingtalkMessageTypeMarkdown    DingtalkMessageType = "markdown"
	DingtalkMessageTypeActionCard  DingtalkMessageType = "actionCard"
	DingtalkMessageTypeLink        DingtalkMessageType = "link"
	DingtalkMessageTypePicture     DingtalkMessageType = "picture"
	DingtalkMessageTypeAudio       DingtalkMessageType = "audio"
	DingtalkMessageTypeVideo       DingtalkMessageType = "video"
	DingtalkMessageTypeFile        DingtalkMessageType = "file"
	DingtalkMessageTypeInteractive DingtalkMessageType = "interactive"
)

// DingtalkEventType 钉钉事件类型
type DingtalkEventType string

const (
	DingtalkEventTypeMessage      DingtalkEventType = "message"
	DingtalkEventTypeCallback     DingtalkEventType = "callback"
	DingtalkEventTypeNotification DingtalkEventType = "notification"
	DingtalkEventTypeSysMessage   DingtalkEventType = "sys_message"
)

// DingtalkConversationType 钉钉会话类型
type DingtalkConversationType string

const (
	DingtalkConversationTypeSingle DingtalkConversationType = "1" // 单聊
	DingtalkConversationTypeGroup  DingtalkConversationType = "2" // 群聊
)

// DingtalkStreamMessage 钉钉Stream消息结构
type DingtalkStreamMessage struct {
	Headers map[string]string `json:"headers"`
	Data    []byte            `json:"data"`
	Topic   string            `json:"topic"`
	Type    string            `json:"type"`
}

// DingtalkCallbackMessage 钉钉回调消息
type DingtalkCallbackMessage struct {
	MsgType                   DingtalkMessageType         `json:"msgtype"`
	ConversationType          DingtalkConversationType    `json:"conversationType"`
	ConversationID            string                      `json:"conversationId"`
	ChatbotCorpID             string                      `json:"chatbotCorpId"`
	ChatbotUserID             string                      `json:"chatbotUserId"`
	MsgID                     string                      `json:"msgId"`
	SenderNick                string                      `json:"senderNick"`
	SenderID                  string                      `json:"senderId"`
	SenderCorpID              string                      `json:"senderCorpId"`
	SessionWebhook            string                      `json:"sessionWebhook"`
	SessionWebhookExpiredTime int64                       `json:"sessionWebhookExpiredTime"`
	CreateAt                  int64                       `json:"createAt"`
	SenderStaffID             string                      `json:"senderStaffId"`
	Text                      *DingtalkTextMessage        `json:"text,omitempty"`
	File                      *DingtalkFileMessage        `json:"file,omitempty"`
	Picture                   *DingtalkPictureMessage     `json:"picture,omitempty"`
	Audio                     *DingtalkAudioMessage       `json:"audio,omitempty"`
	Video                     *DingtalkVideoMessage       `json:"video,omitempty"`
	Interactive               *DingtalkInteractiveMessage `json:"interactive,omitempty"`
}

// DingtalkTextMessage 钉钉文本消息
type DingtalkTextMessage struct {
	Content string `json:"content"`
}

// DingtalkFileMessage 钉钉文件消息
type DingtalkFileMessage struct {
	DownloadCode string `json:"downloadCode"`
	FileName     string `json:"fileName"`
	FileType     string `json:"fileType"`
}

// DingtalkPictureMessage 钉钉图片消息
type DingtalkPictureMessage struct {
	DownloadCode string `json:"downloadCode"`
	PictureType  string `json:"pictureType"`
}

// DingtalkAudioMessage 钉钉音频消息
type DingtalkAudioMessage struct {
	DownloadCode string `json:"downloadCode"`
	Duration     int64  `json:"duration"`
	Recognition  string `json:"recognition"`
}

// DingtalkVideoMessage 钉钉视频消息
type DingtalkVideoMessage struct {
	DownloadCode string `json:"downloadCode"`
	Duration     int64  `json:"duration"`
	VideoType    string `json:"videoType"`
}

// DingtalkInteractiveMessage 钉钉交互式消息
type DingtalkInteractiveMessage struct {
	ActionCardID string `json:"actionCardId"`
	Title        string `json:"title"`
	Content      string `json:"content"`
}

// DingtalkSendMessage 钉钉发送消息结构
type DingtalkSendMessage struct {
	MsgType    DingtalkMessageType     `json:"msgtype"`
	UserID     string                  `json:"userid,omitempty"`
	DeptIDList string                  `json:"dept_id_list,omitempty"`
	ToAllUser  bool                    `json:"to_all_user,omitempty"`
	Text       *DingtalkSendText       `json:"text,omitempty"`
	Markdown   *DingtalkSendMarkdown   `json:"markdown,omitempty"`
	ActionCard *DingtalkSendActionCard `json:"actionCard,omitempty"`
}

// DingtalkSendText 钉钉发送文本消息
type DingtalkSendText struct {
	Content string `json:"content"`
}

// DingtalkSendMarkdown 钉钉发送Markdown消息
type DingtalkSendMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// DingtalkSendActionCard 钉钉发送ActionCard消息
type DingtalkSendActionCard struct {
	Title          string                     `json:"title"`
	Markdown       string                     `json:"markdown"`
	SingleTitle    string                     `json:"singleTitle,omitempty"`
	SingleURL      string                     `json:"singleURL,omitempty"`
	BtnOrientation string                     `json:"btnOrientation,omitempty"`
	Buttons        []DingtalkActionCardButton `json:"btns,omitempty"`
}

// DingtalkActionCardButton ActionCard按钮
type DingtalkActionCardButton struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

// DingtalkConfig 钉钉平台配置
type DingtalkStreamConfig struct {
	AppKey         string `json:"app_key" validate:"required"`
	AppSecret      string `json:"app_secret" validate:"required"`
	ClientID       string `json:"client_id" validate:"required"`
	ClientSecret   string `json:"client_secret" validate:"required"`
	SubscriptionID string `json:"subscription_id,omitempty"`
	Topic          string `json:"topic,omitempty"`
}

// ParseDingtalkStreamConfig 解析钉钉Stream配置
func ParseDingtalkStreamConfig(config map[string]interface{}) (*DingtalkStreamConfig, error) {
	streamConfig := &DingtalkStreamConfig{}

	if clientID, ok := config["client_id"].(string); ok {
		streamConfig.ClientID = clientID
	}

	if clientSecret, ok := config["client_secret"].(string); ok {
		streamConfig.ClientSecret = clientSecret
	}

	if appKey, ok := config["app_key"].(string); ok {
		streamConfig.AppKey = appKey
	}

	if appSecret, ok := config["app_secret"].(string); ok {
		streamConfig.AppSecret = appSecret
	}

	if subscriptionID, ok := config["subscription_id"].(string); ok {
		streamConfig.SubscriptionID = subscriptionID
	}

	if topic, ok := config["topic"].(string); ok {
		streamConfig.Topic = topic
	}

	return streamConfig, nil
}

// ToUnifiedMessage 转换为统一消息格式
func (msg *DingtalkCallbackMessage) ToUnifiedMessage() *entity.UnifiedMessage {
	unifiedMsg := &entity.UnifiedMessage{
		MessageID:         msg.MsgID,
		MessageType:       string(msg.MsgType),
		SenderID:          msg.SenderID,
		SenderName:        msg.SenderNick,
		SenderType:        "user",
		ConversationID:    msg.ConversationID,
		PlatformMessageID: msg.MsgID,
		PlatformTimestamp: time.Unix(msg.CreateAt/1000, 0),
		RawContent:        make(map[string]interface{}),
	}

	// 设置接收者信息
	if msg.ConversationType == DingtalkConversationTypeSingle {
		unifiedMsg.ReceiverID = msg.ChatbotUserID
		unifiedMsg.ReceiverType = "bot"
	} else {
		unifiedMsg.ReceiverID = msg.ConversationID
		unifiedMsg.ReceiverType = "group"
	}

	// 根据消息类型设置内容
	switch msg.MsgType {
	case DingtalkMessageTypeText:
		if msg.Text != nil {
			unifiedMsg.Content = msg.Text.Content
		}
	case DingtalkMessageTypePicture:
		if msg.Picture != nil {
			unifiedMsg.Content = "[图片]"
			unifiedMsg.MediaType = "image"
			unifiedMsg.RawContent["downloadCode"] = msg.Picture.DownloadCode
			unifiedMsg.RawContent["pictureType"] = msg.Picture.PictureType
		}
	case DingtalkMessageTypeAudio:
		if msg.Audio != nil {
			unifiedMsg.Content = "[语音]"
			unifiedMsg.MediaType = "audio"
			unifiedMsg.RawContent["downloadCode"] = msg.Audio.DownloadCode
			unifiedMsg.RawContent["duration"] = msg.Audio.Duration
			unifiedMsg.RawContent["recognition"] = msg.Audio.Recognition
		}
	case DingtalkMessageTypeVideo:
		if msg.Video != nil {
			unifiedMsg.Content = "[视频]"
			unifiedMsg.MediaType = "video"
			unifiedMsg.RawContent["downloadCode"] = msg.Video.DownloadCode
			unifiedMsg.RawContent["duration"] = msg.Video.Duration
			unifiedMsg.RawContent["videoType"] = msg.Video.VideoType
		}
	case DingtalkMessageTypeFile:
		if msg.File != nil {
			unifiedMsg.Content = "[文件] " + msg.File.FileName
			unifiedMsg.MediaType = "file"
			unifiedMsg.RawContent["downloadCode"] = msg.File.DownloadCode
			unifiedMsg.RawContent["fileName"] = msg.File.FileName
			unifiedMsg.RawContent["fileType"] = msg.File.FileType
		}
	case DingtalkMessageTypeInteractive:
		if msg.Interactive != nil {
			unifiedMsg.Content = msg.Interactive.Title
			unifiedMsg.RawContent["actionCardId"] = msg.Interactive.ActionCardID
			unifiedMsg.RawContent["title"] = msg.Interactive.Title
			unifiedMsg.RawContent["content"] = msg.Interactive.Content
		}
	}

	// 存储原始消息数据
	unifiedMsg.RawContent["conversationType"] = msg.ConversationType
	unifiedMsg.RawContent["chatbotCorpId"] = msg.ChatbotCorpID
	unifiedMsg.RawContent["senderCorpId"] = msg.SenderCorpID
	unifiedMsg.RawContent["senderStaffId"] = msg.SenderStaffID
	unifiedMsg.RawContent["sessionWebhook"] = msg.SessionWebhook
	unifiedMsg.RawContent["sessionWebhookExpiredTime"] = msg.SessionWebhookExpiredTime

	return unifiedMsg
}
