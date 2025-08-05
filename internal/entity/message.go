package entity

import (
	"time"
)

// Message 消息实体，存储统一格式的消息数据
type Message struct {
	BaseEntity
	MessageID         string           `json:"message_id" gorm:"column:message_id;type:varchar(100);uniqueIndex;not null;comment:消息唯一ID"`
	ChannelID         string            `json:"channel_id" gorm:"column:channel_id;not null;index;comment:所属通道ID"`
	PlatformMessageID string           `json:"platform_message_id" gorm:"column:platform_message_id;type:varchar(200);comment:平台消息ID"`
	Direction         MessageDirection `json:"direction" gorm:"column:direction;type:tinyint;not null;comment:消息方向:1-入站,2-出站"`
	MessageType       string           `json:"message_type" gorm:"column:message_type;type:varchar(50);not null;comment:消息类型"`
	ContentType       string           `json:"content_type" gorm:"column:content_type;type:varchar(50);comment:内容类型"`
	SenderID          string           `json:"sender_id" gorm:"column:sender_id;type:varchar(100);index;comment:发送者ID"`
	SenderName        string           `json:"sender_name" gorm:"column:sender_name;type:varchar(100);comment:发送者名称"`
	SenderType        string           `json:"sender_type" gorm:"column:sender_type;type:varchar(50);comment:发送者类型"`
	ReceiverID        string           `json:"receiver_id" gorm:"column:receiver_id;type:varchar(100);index;comment:接收者ID"`
	ReceiverName      string           `json:"receiver_name" gorm:"column:receiver_name;type:varchar(100);comment:接收者名称"`
	ReceiverType      string           `json:"receiver_type" gorm:"column:receiver_type;type:varchar(50);comment:接收者类型"`
	Content           string           `json:"content" gorm:"column:content;type:text;comment:消息内容"`
	RawContent        JSONField        `json:"raw_content" gorm:"column:raw_content;type:json;comment:原始内容"`
	UnifiedContent    JSONField        `json:"unified_content" gorm:"column:unified_content;type:json;comment:统一格式内容"`
	MediaURL          string           `json:"media_url" gorm:"column:media_url;type:varchar(500);comment:媒体文件URL"`
	MediaType         string           `json:"media_type" gorm:"column:media_type;type:varchar(50);comment:媒体类型"`
	FileSize          int64            `json:"file_size" gorm:"column:file_size;comment:文件大小"`
	MessageStatus     MessageStatus    `json:"message_status" gorm:"column:message_status;type:tinyint;default:0;comment:消息状态:0-待处理,1-处理中,2-已处理,3-已发送,4-失败,5-过期"`
	RetryCount        int              `json:"retry_count" gorm:"column:retry_count;default:0;comment:重试次数"`
	ErrorMessage      string           `json:"error_message" gorm:"column:error_message;type:text;comment:错误消息"`
	ParentMessageID   *string          `json:"parent_message_id" gorm:"column:parent_message_id;index;comment:父消息ID"`
	ConversationID    string           `json:"conversation_id" gorm:"column:conversation_id;type:varchar(100);index;comment:会话ID"`
	BackendRequestID  string           `json:"backend_request_id" gorm:"column:backend_request_id;type:varchar(100);comment:后端请求ID"`
	PlatformTimestamp time.Time        `json:"platform_timestamp" gorm:"column:platform_timestamp;comment:平台时间戳"`
	ReceivedAt        time.Time        `json:"received_at" gorm:"column:received_at;autoCreateTime;comment:接收时间"`
	ProcessedAt       *time.Time       `json:"processed_at" gorm:"column:processed_at;comment:处理时间"`
	SentAt            *time.Time       `json:"sent_at" gorm:"column:sent_at;comment:发送时间"`

	// 关联关系
	Channel       *Channel  `json:"channel,omitempty" gorm:"foreignKey:ChannelID;references:ID"`
	ParentMessage *Message  `json:"parent_message,omitempty" gorm:"foreignKey:ParentMessageID;references:ID"`
	Replies       []Message `json:"replies,omitempty" gorm:"foreignKey:ParentMessageID;references:ID"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}
