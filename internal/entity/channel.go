package entity

import (
	"time"
)

// Channel 平台通道实体，管理单个平台连接的配置和状态
type Channel struct {
	BaseEntity
	BotID                int64            `json:"bot_id" gorm:"column:bot_id;not null;index;comment:所属机器人ID"`
	PlatformType         string           `json:"platform_type" gorm:"column:platform_type;type:varchar(50);not null;comment:平台类型"`
	ChannelName          string           `json:"channel_name" gorm:"column:channel_name;type:varchar(100);not null;comment:通道名称"`
	WebhookPath          string           `json:"webhook_path" gorm:"column:webhook_path;type:varchar(200);comment:Webhook路径"`
	Config               JSONField        `json:"config" gorm:"column:config;type:json;comment:平台配置信息"`
	AccessToken          string           `json:"access_token" gorm:"column:access_token;type:varchar(500);comment:访问令牌"`
	AccessTokenExpiresAt *time.Time       `json:"access_token_expires_at" gorm:"column:access_token_expires_at;comment:令牌过期时间"`
	ConnectionStatus     ConnectionStatus `json:"connection_status" gorm:"column:connection_status;type:tinyint;default:0;comment:连接状态:0-断开,1-已连接,2-连接中,3-错误"`
	LastConnectedAt      *time.Time       `json:"last_connected_at" gorm:"column:last_connected_at;comment:最后连接时间"`
	Status               Status           `json:"status" gorm:"column:status;type:tinyint;default:1;comment:状态:0-未激活,1-激活,2-已删除,3-暂停"`

	// 关联关系
	Bot            *Bot            `json:"bot,omitempty" gorm:"foreignKey:BotID;references:ID"`
	Messages       []Message       `json:"messages,omitempty" gorm:"foreignKey:ChannelID;references:ID"`
	ConnectionLogs []ConnectionLog `json:"connection_logs,omitempty" gorm:"foreignKey:ChannelID;references:ID"`
	APICallLogs    []APICallLog    `json:"api_call_logs,omitempty" gorm:"foreignKey:ChannelID;references:ID"`
}

// TableName 指定表名
func (Channel) TableName() string {
	return "channels"
}

// IsActive 检查通道是否处于激活状态
func (c *Channel) IsActive() bool {
	return c.Status.IsActive()
}

// IsConnected 检查通道是否已连接
func (c *Channel) IsConnected() bool {
	return c.ConnectionStatus == ConnectionStatusConnected
}

// IsTokenExpired 检查访问令牌是否已过期
func (c *Channel) IsTokenExpired() bool {
	if c.AccessTokenExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.AccessTokenExpiresAt)
}

// GetConfigValue 获取配置中的特定值
func (c *Channel) GetConfigValue(key string) interface{} {
	if c.Config == nil {
		return nil
	}
	return c.Config.Get(key)
}

// SetConfigValue 设置配置中的特定值
func (c *Channel) SetConfigValue(key string, value interface{}) {
	if c.Config == nil {
		c.Config = make(JSONField)
	}
	c.Config.Set(key, value)
}

// GetWebhookURL 获取完整的Webhook URL
func (c *Channel) GetWebhookURL(baseURL string) string {
	if c.WebhookPath == "" {
		return ""
	}
	return baseURL + c.WebhookPath
}

// UpdateConnectionStatus 更新连接状态
func (c *Channel) UpdateConnectionStatus(status ConnectionStatus) {
	c.ConnectionStatus = status
	if status == ConnectionStatusConnected {
		now := time.Now()
		c.LastConnectedAt = &now
	}
}

// Validate 验证Channel实体数据
func (c *Channel) Validate() error {
	if c.BotID <= 0 {
		return NewValidationError("bot_id", "机器人ID必须大于0")
	}
	if c.PlatformType == "" {
		return NewValidationError("platform_type", "平台类型不能为空")
	}
	if c.ChannelName == "" {
		return NewValidationError("channel_name", "通道名称不能为空")
	}
	return nil
}
