package entity

// MessageProcessor 消息处理器实体，定义消息处理逻辑
type MessageProcessor struct {
	BaseEntity
	ProcessorName string    `json:"processor_name" gorm:"column:processor_name;type:varchar(100);not null;comment:处理器名称"`
	ProcessorType string    `json:"processor_type" gorm:"column:processor_type;type:varchar(50);not null;comment:处理器类型"`
	Config        JSONField `json:"config" gorm:"column:config;type:json;comment:处理器配置"`
	Priority      int       `json:"priority" gorm:"column:priority;default:100;comment:优先级"`
	Status        Status    `json:"status" gorm:"column:status;type:tinyint;default:1;comment:状态:0-未激活,1-激活,2-已删除,3-暂停"`

	// 关联关系
	RoutingRules []MessageRoutingRule `json:"routing_rules,omitempty" gorm:"foreignKey:ProcessorID;references:ID"`
	APICallLogs  []APICallLog         `json:"api_call_logs,omitempty" gorm:"foreignKey:ProcessorID;references:ID"`
}

// TableName 指定表名
func (MessageProcessor) TableName() string {
	return "message_processors"
}

// IsActive 检查处理器是否处于激活状态
func (mp *MessageProcessor) IsActive() bool {
	return mp.Status.IsActive()
}

// GetConfigValue 获取配置中的特定值
func (mp *MessageProcessor) GetConfigValue(key string) interface{} {
	if mp.Config == nil {
		return nil
	}
	return mp.Config.Get(key)
}

// SetConfigValue 设置配置中的特定值
func (mp *MessageProcessor) SetConfigValue(key string, value interface{}) {
	if mp.Config == nil {
		mp.Config = make(JSONField)
	}
	mp.Config.Set(key, value)
}

// Validate 验证MessageProcessor实体数据
func (mp *MessageProcessor) Validate() error {
	if mp.ProcessorName == "" {
		return NewValidationError("processor_name", "处理器名称不能为空")
	}
	if mp.ProcessorType == "" {
		return NewValidationError("processor_type", "处理器类型不能为空")
	}
	return nil
}

// MessageRoutingRule 消息路由规则实体，定义消息路由逻辑
type MessageRoutingRule struct {
	BaseEntity
	RuleName        string    `json:"rule_name" gorm:"column:rule_name;type:varchar(100);not null;comment:规则名称"`
	RuleDescription string    `json:"rule_description" gorm:"column:rule_description;type:text;comment:规则描述"`
	BotIDs          JSONField `json:"bot_ids" gorm:"column:bot_ids;type:json;comment:适用的Bot ID列表"`
	PlatformTypes   JSONField `json:"platform_types" gorm:"column:platform_types;type:json;comment:适用的平台类型列表"`
	ChannelIDs      JSONField `json:"channel_ids" gorm:"column:channel_ids;type:json;comment:适用的通道ID列表"`
	MessageTypes    JSONField `json:"message_types" gorm:"column:message_types;type:json;comment:适用的消息类型列表"`
	SenderPatterns  JSONField `json:"sender_patterns" gorm:"column:sender_patterns;type:json;comment:发送者匹配模式"`
	ContentPatterns JSONField `json:"content_patterns" gorm:"column:content_patterns;type:json;comment:内容匹配模式"`
	ProcessorID     int64     `json:"processor_id" gorm:"column:processor_id;not null;index;comment:处理器ID"`
	RouteType       RouteType `json:"route_type" gorm:"column:route_type;type:tinyint;not null;comment:路由类型:1-直接,2-转发,3-广播,4-条件"`
	Priority        int       `json:"priority" gorm:"column:priority;default:100;comment:优先级"`
	IsFallback      bool      `json:"is_fallback" gorm:"column:is_fallback;default:false;comment:是否为兜底规则"`
	ConditionLogic  string    `json:"condition_logic" gorm:"column:condition_logic;type:varchar(50);comment:条件逻辑(AND/OR)"`
	Status          Status    `json:"status" gorm:"column:status;type:tinyint;default:1;comment:状态:0-未激活,1-激活,2-已删除,3-暂停"`

	// 关联关系
	MessageProcessor *MessageProcessor `json:"message_processor,omitempty" gorm:"foreignKey:ProcessorID;references:ID"`
}

// TableName 指定表名
func (MessageRoutingRule) TableName() string {
	return "message_routing_rules"
}

// IsActive 检查路由规则是否处于激活状态
func (mrr *MessageRoutingRule) IsActive() bool {
	return mrr.Status.IsActive()
}

// MatchesBotID 检查是否匹配指定的Bot ID
func (mrr *MessageRoutingRule) MatchesBotID(botID int64) bool {
	if mrr.BotIDs == nil {
		return true // 空配置表示匹配所有
	}

	botIDs, ok := mrr.BotIDs["ids"]
	if !ok {
		return true
	}

	if idList, ok := botIDs.([]interface{}); ok {
		for _, id := range idList {
			if idFloat, ok := id.(float64); ok && int64(idFloat) == botID {
				return true
			}
		}
	}
	return false
}

// MatchesPlatformType 检查是否匹配指定的平台类型
func (mrr *MessageRoutingRule) MatchesPlatformType(platformType string) bool {
	if mrr.PlatformTypes == nil {
		return true
	}

	types, ok := mrr.PlatformTypes["types"]
	if !ok {
		return true
	}

	if typeList, ok := types.([]interface{}); ok {
		for _, t := range typeList {
			if typeStr, ok := t.(string); ok && typeStr == platformType {
				return true
			}
		}
	}
	return false
}

// MatchesChannelID 检查是否匹配指定的通道ID
func (mrr *MessageRoutingRule) MatchesChannelID(channelID int64) bool {
	if mrr.ChannelIDs == nil {
		return true
	}

	channelIDs, ok := mrr.ChannelIDs["ids"]
	if !ok {
		return true
	}

	if idList, ok := channelIDs.([]interface{}); ok {
		for _, id := range idList {
			if idFloat, ok := id.(float64); ok && int64(idFloat) == channelID {
				return true
			}
		}
	}
	return false
}

// MatchesMessageType 检查是否匹配指定的消息类型
func (mrr *MessageRoutingRule) MatchesMessageType(messageType string) bool {
	if mrr.MessageTypes == nil {
		return true
	}

	types, ok := mrr.MessageTypes["types"]
	if !ok {
		return true
	}

	if typeList, ok := types.([]interface{}); ok {
		for _, t := range typeList {
			if typeStr, ok := t.(string); ok && typeStr == messageType {
				return true
			}
		}
	}
	return false
}

// Validate 验证MessageRoutingRule实体数据
func (mrr *MessageRoutingRule) Validate() error {
	if mrr.RuleName == "" {
		return NewValidationError("rule_name", "规则名称不能为空")
	}
	if mrr.ProcessorID <= 0 {
		return NewValidationError("processor_id", "处理器ID必须大于0")
	}
	return nil
}
