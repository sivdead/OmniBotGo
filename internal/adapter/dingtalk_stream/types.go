package dingtalk_stream

import (
	"time"

	"github.com/sivdead/OmniBotGo/internal/dto"
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
	// 机器人相关事件
	DingtalkEventTypeRobotAdded   DingtalkEventType = "robot_added"
	DingtalkEventTypeRobotRemoved DingtalkEventType = "robot_removed"
	DingtalkEventTypeGroupJoin    DingtalkEventType = "group_join"
	DingtalkEventTypeGroupLeave   DingtalkEventType = "group_leave"
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
	MsgType                   DingtalkMessageType      `json:"msgtype"`
	ConversationType          DingtalkConversationType `json:"conversationType"`
	ConversationID            string                   `json:"conversationId"`
	ChatbotCorpID             string                   `json:"chatbotCorpId"`
	ChatbotUserID             string                   `json:"chatbotUserId"`
	MsgID                     string                   `json:"msgId"`
	SenderNick                string                   `json:"senderNick"`
	SenderID                  string                   `json:"senderId"`
	SenderCorpID              string                   `json:"senderCorpId"`
	SessionWebhook            string                   `json:"sessionWebhook"`
	SessionWebhookExpiredTime int64                    `json:"sessionWebhookExpiredTime"`
	CreateAt                  int64                    `json:"createAt"`
	SenderStaffID             string                   `json:"senderStaffId"`
	// 消息内容
	Text        *DingtalkTextMessage        `json:"text,omitempty"`
	File        *DingtalkFileMessage        `json:"file,omitempty"`
	Picture     *DingtalkPictureMessage     `json:"picture,omitempty"`
	Audio       *DingtalkAudioMessage       `json:"audio,omitempty"`
	Video       *DingtalkVideoMessage       `json:"video,omitempty"`
	Interactive *DingtalkInteractiveMessage `json:"interactive,omitempty"`
	// 事件相关
	EventType DingtalkEventType      `json:"eventType,omitempty"`
	EventData map[string]interface{} `json:"eventData,omitempty"`
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

// ToUnifiedMessage 转换为统一消息格式
func (msg *DingtalkCallbackMessage) ToUnifiedMessage() *dto.UnifiedMessage {
	unifiedMsg := &dto.UnifiedMessage{
		MessageID:         msg.MsgID,
		MessageType:       mapDingtalkMessageType(msg.MsgType),
		SenderID:          msg.SenderID,
		SenderName:        msg.SenderNick,
		SenderType:        entity.SenderTypeUser,
		ConversationID:    msg.ConversationID,
		PlatformMessageID: msg.MsgID,
		PlatformTimestamp: time.Unix(msg.CreateAt/1000, 0),
		RawContent:        make(map[string]interface{}),
	}

	// 设置接收者信息
	if msg.ConversationType == DingtalkConversationTypeSingle {
		unifiedMsg.ReceiverID = msg.ChatbotUserID
		unifiedMsg.ReceiverType = entity.ReceiverTypeUser
	} else {
		unifiedMsg.ReceiverID = msg.ConversationID
		unifiedMsg.ReceiverType = entity.ReceiverTypeGroup
	}

	// 处理事件消息
	if msg.EventType != "" {
		unifiedMsg.MessageType = entity.MessageTypeEvent
		eventType := mapDingtalkEventType(msg.EventType)
		unifiedMsg.EventContent = &entity.EventMessage{
			EventType: eventType,
			EventKey:  string(msg.EventType),
			EventData: msg.EventData,
		}
		unifiedMsg.Content = "[事件] " + eventType
	} else {
		// 根据消息类型设置内容
		switch msg.MsgType {
		case DingtalkMessageTypeText:
			if msg.Text != nil {
				unifiedMsg.Content = msg.Text.Content
			}

		case DingtalkMessageTypePicture:
			if msg.Picture != nil {
				unifiedMsg.Content = "[图片]"
				unifiedMsg.MediaType = entity.MediaTypeImage
				unifiedMsg.RawContent["downloadCode"] = msg.Picture.DownloadCode
				unifiedMsg.RawContent["pictureType"] = msg.Picture.PictureType
			}

		case DingtalkMessageTypeAudio:
			if msg.Audio != nil {
				unifiedMsg.Content = "[语音]"
				unifiedMsg.MediaType = entity.MediaTypeVoice
				unifiedMsg.RawContent["downloadCode"] = msg.Audio.DownloadCode
				unifiedMsg.RawContent["duration"] = msg.Audio.Duration
				unifiedMsg.RawContent["recognition"] = msg.Audio.Recognition
			}

		case DingtalkMessageTypeVideo:
			if msg.Video != nil {
				unifiedMsg.Content = "[视频]"
				unifiedMsg.MediaType = entity.MediaTypeVideo
				unifiedMsg.RawContent["downloadCode"] = msg.Video.DownloadCode
				unifiedMsg.RawContent["duration"] = msg.Video.Duration
				unifiedMsg.RawContent["videoType"] = msg.Video.VideoType
			}

		case DingtalkMessageTypeFile:
			if msg.File != nil {
				unifiedMsg.Content = "[文件] " + msg.File.FileName
				unifiedMsg.MediaType = entity.MediaTypeFile
				unifiedMsg.FileContent = &entity.FileMessage{
					FileName: msg.File.FileName,
					FileType: msg.File.FileType,
				}
				unifiedMsg.RawContent["downloadCode"] = msg.File.DownloadCode
				unifiedMsg.RawContent["fileName"] = msg.File.FileName
				unifiedMsg.RawContent["fileType"] = msg.File.FileType
			}

		case DingtalkMessageTypeInteractive:
			if msg.Interactive != nil {
				unifiedMsg.MessageType = entity.MessageTypeCard
				unifiedMsg.Content = msg.Interactive.Title
				unifiedMsg.CardContent = &entity.CardMessage{
					CardType: entity.CardTypeInteractive,
					Title:    msg.Interactive.Title,
					Content:  msg.Interactive.Content,
				}
				unifiedMsg.RawContent["actionCardId"] = msg.Interactive.ActionCardID
			}
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

// mapDingtalkMessageType 映射钉钉消息类型到统一消息类型
func mapDingtalkMessageType(msgType DingtalkMessageType) string {
	switch msgType {
	case DingtalkMessageTypeText:
		return entity.MessageTypeText
	case DingtalkMessageTypeMarkdown:
		return entity.MessageTypeMarkdown
	case DingtalkMessageTypeActionCard, DingtalkMessageTypeInteractive:
		return entity.MessageTypeCard
	case DingtalkMessageTypeLink:
		return entity.MessageTypeLink
	case DingtalkMessageTypePicture:
		return entity.MessageTypeImage
	case DingtalkMessageTypeAudio:
		return entity.MessageTypeAudio
	case DingtalkMessageTypeVideo:
		return entity.MessageTypeVideo
	case DingtalkMessageTypeFile:
		return entity.MessageTypeFile
	default:
		return entity.MessageTypeText
	}
}

// mapDingtalkEventType 映射钉钉事件类型到统一事件类型
func mapDingtalkEventType(eventType DingtalkEventType) string {
	switch eventType {
	case DingtalkEventTypeRobotAdded:
		return entity.EventTypeUserSubscribe
	case DingtalkEventTypeRobotRemoved:
		return entity.EventTypeUserUnsubscribe
	case DingtalkEventTypeGroupJoin:
		return entity.EventTypeGroupJoin
	case DingtalkEventTypeGroupLeave:
		return entity.EventTypeGroupLeave
	default:
		return entity.EventTypeCustom
	}
}
