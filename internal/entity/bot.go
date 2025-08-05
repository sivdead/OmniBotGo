package entity

// Bot 机器人实体，管理单个Bot实例的配置和状态
type Bot struct {
	BaseEntity
	BotName     string    `json:"bot_name" gorm:"column:bot_name;type:varchar(100);not null;comment:机器人名称"`
	BotType     string    `json:"bot_type" gorm:"column:bot_type;type:varchar(50);not null;comment:机器人类型"`
	Description string    `json:"description" gorm:"column:description;type:text;comment:机器人描述"`
	AvatarURL   string    `json:"avatar_url" gorm:"column:avatar_url;type:varchar(500);comment:头像URL"`
	Config      JSONField `json:"config" gorm:"column:config;type:json;comment:机器人配置信息"`
	Status      Status    `json:"status" gorm:"column:status;type:tinyint;default:1;comment:状态:0-未激活,1-激活,2-已删除,3-暂停"`
	CreatedBy   string    `json:"created_by" gorm:"column:created_by;type:varchar(100);comment:创建者"`

	// 关联关系
	Channels []Channel `json:"channels,omitempty" gorm:"foreignKey:BotID;references:ID"`
}

// TableName 指定表名
func (Bot) TableName() string {
	return "bots"
}

// IsActive 检查机器人是否处于激活状态
func (b *Bot) IsActive() bool {
	return b.Status.IsActive()
}

// GetConfigValue 获取配置中的特定值
func (b *Bot) GetConfigValue(key string) interface{} {
	if b.Config == nil {
		return nil
	}
	return b.Config.Get(key)
}

// SetConfigValue 设置配置中的特定值
func (b *Bot) SetConfigValue(key string, value interface{}) {
	if b.Config == nil {
		b.Config = make(JSONField)
	}
	b.Config.Set(key, value)
}

// GetDefaultWebhookPath 获取默认的Webhook路径
func (b *Bot) GetDefaultWebhookPath() string {
	if path := b.Config.GetString("webhook_path"); path != "" {
		return path
	}
	return "/webhook/" + b.BotType + "/" + b.ID
}

// Validate 验证Bot实体数据
func (b *Bot) Validate() error {
	if b.BotName == "" {
		return NewValidationError("bot_name", "机器人名称不能为空")
	}
	if b.BotType == "" {
		return NewValidationError("bot_type", "机器人类型不能为空")
	}
	return nil
}
