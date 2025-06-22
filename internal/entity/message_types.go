package entity

// MessageType 消息类型常量
const (
	// 基础消息类型
	MessageTypeText     = "text"     // 文本消息
	MessageTypeImage    = "image"    // 图片消息
	MessageTypeAudio    = "audio"    // 音频消息
	MessageTypeVideo    = "video"    // 视频消息
	MessageTypeFile     = "file"     // 文件消息
	MessageTypeLocation = "location" // 位置消息
	MessageTypeLink     = "link"     // 链接消息

	// 富文本消息类型
	MessageTypeMarkdown = "markdown" // Markdown消息
	MessageTypeCard     = "card"     // 卡片消息
	MessageTypeNews     = "news"     // 图文消息

	// 交互消息类型
	MessageTypeMenu     = "menu"     // 菜单消息
	MessageTypeTemplate = "template" // 模板消息

	// 事件消息类型
	MessageTypeEvent = "event" // 事件消息
)

// EventType 事件类型常量
const (
	// 用户事件
	EventTypeUserSubscribe   = "user_subscribe"   // 用户关注/订阅
	EventTypeUserUnsubscribe = "user_unsubscribe" // 用户取消关注/退订
	EventTypeUserUpdate      = "user_update"      // 用户信息更新
	EventTypeUserEnter       = "user_enter"       // 用户进入会话
	EventTypeUserLeave       = "user_leave"       // 用户离开会话

	// 群组事件
	EventTypeGroupJoin   = "group_join"   // 加入群组
	EventTypeGroupLeave  = "group_leave"  // 离开群组
	EventTypeGroupCreate = "group_create" // 创建群组
	EventTypeGroupUpdate = "group_update" // 更新群组
	EventTypeGroupDelete = "group_delete" // 删除群组
	EventTypeGroupKick   = "group_kick"   // 被踢出群组

	// 菜单事件
	EventTypeMenuClick = "menu_click" // 菜单点击
	EventTypeMenuView  = "menu_view"  // 菜单跳转

	// 扫码事件
	EventTypeScanQRCode    = "scan_qrcode"    // 扫描二维码
	EventTypeScanSubscribe = "scan_subscribe" // 扫码关注

	// 位置事件
	EventTypeLocationReport = "location_report" // 位置上报
	EventTypeLocationSelect = "location_select" // 位置选择

	// 业务事件
	EventTypeCheckIn  = "check_in" // 签到/打卡
	EventTypeApproval = "approval" // 审批

	// 自定义事件
	EventTypeCustom = "custom" // 自定义事件
)

// MediaType 媒体类型常量
const (
	MediaTypeImage    = "image"    // 图片
	MediaTypeVoice    = "voice"    // 语音
	MediaTypeVideo    = "video"    // 视频
	MediaTypeFile     = "file"     // 文件
	MediaTypeDocument = "document" // 文档
)

// CardType 卡片类型常量
const (
	CardTypeText        = "text"        // 文本卡片
	CardTypeImage       = "image"       // 图片卡片
	CardTypeButton      = "button"      // 按钮卡片
	CardTypeActionCard  = "action_card" // 动作卡片（钉钉）
	CardTypeFeedCard    = "feed_card"   // 订阅卡片（钉钉）
	CardTypeMarkdown    = "markdown"    // Markdown卡片
	CardTypeInteractive = "interactive" // 交互式卡片（飞书）
)

// SenderType 发送者类型常量
const (
	SenderTypeUser   = "user"   // 用户
	SenderTypeGroup  = "group"  // 群组
	SenderTypeSystem = "system" // 系统
	SenderTypeBot    = "bot"    // 机器人
)

// ReceiverType 接收者类型常量
const (
	ReceiverTypeUser  = "user"  // 用户
	ReceiverTypeGroup = "group" // 群组
	ReceiverTypeAll   = "all"   // 所有人
	ReceiverTypeBot   = "bot"   // 机器人
)

// MessageStructures 消息结构定义

// MarkdownMessage Markdown消息结构
type MarkdownMessage struct {
	Title   string `json:"title,omitempty"` // 标题（可选）
	Content string `json:"content"`         // Markdown内容
}

// CardMessage 卡片消息结构
type CardMessage struct {
	CardType string                 `json:"card_type"`         // 卡片类型
	Title    string                 `json:"title"`             // 标题
	Content  string                 `json:"content"`           // 内容
	Buttons  []CardButton           `json:"buttons,omitempty"` // 按钮列表
	Extra    map[string]interface{} `json:"extra,omitempty"`   // 扩展字段
}

// CardButton 卡片按钮
type CardButton struct {
	Title      string `json:"title"`       // 按钮标题
	ActionURL  string `json:"action_url"`  // 点击跳转URL
	ActionType string `json:"action_type"` // 动作类型
}

// NewsMessage 图文消息结构
type NewsMessage struct {
	Articles []NewsArticle `json:"articles"` // 图文列表
}

// NewsArticle 图文项
type NewsArticle struct {
	Title       string `json:"title"`       // 标题
	Description string `json:"description"` // 描述
	URL         string `json:"url"`         // 跳转链接
	PicURL      string `json:"pic_url"`     // 图片链接
}

// FileMessage 文件消息结构
type FileMessage struct {
	FileName string `json:"file_name"` // 文件名
	FileURL  string `json:"file_url"`  // 文件URL
	FileSize int64  `json:"file_size"` // 文件大小
	FileType string `json:"file_type"` // 文件类型
}

// LocationMessage 位置消息结构
type LocationMessage struct {
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
	Precision float64 `json:"precision"` // 精度
	Address   string  `json:"address"`   // 地址描述
}

// EventMessage 事件消息结构
type EventMessage struct {
	EventType string                 `json:"event_type"` // 事件类型
	EventKey  string                 `json:"event_key"`  // 事件键值
	EventData map[string]interface{} `json:"event_data"` // 事件数据
}

// TemplateMessage 模板消息结构
type TemplateMessage struct {
	TemplateID   string                 `json:"template_id"`   // 模板ID
	TemplateData map[string]interface{} `json:"template_data"` // 模板数据
	URL          string                 `json:"url,omitempty"` // 跳转链接
}

// IsRichMessageType 判断是否为富文本消息类型
func IsRichMessageType(messageType string) bool {
	switch messageType {
	case MessageTypeMarkdown, MessageTypeCard, MessageTypeNews, MessageTypeTemplate:
		return true
	default:
		return false
	}
}

// IsEventMessageType 判断是否为事件消息类型
func IsEventMessageType(messageType string) bool {
	return messageType == MessageTypeEvent
}

// IsMediaMessageType 判断是否为媒体消息类型
func IsMediaMessageType(messageType string) bool {
	switch messageType {
	case MessageTypeImage, MessageTypeAudio, MessageTypeVideo, MessageTypeFile:
		return true
	default:
		return false
	}
}
