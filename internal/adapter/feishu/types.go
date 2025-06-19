package feishu

// FeishuMessage 飞书接收消息结构
type FeishuMessage struct {
	Schema string       `json:"schema"`
	Header FeishuHeader `json:"header"`
	Event  FeishuEvent  `json:"event"`
}

// FeishuHeader 飞书消息头
type FeishuHeader struct {
	EventID    string `json:"event_id"`
	EventType  string `json:"event_type"`
	CreateTime string `json:"create_time"`
	Token      string `json:"token"`
	AppID      string `json:"app_id"`
	TenantKey  string `json:"tenant_key"`
}

// FeishuEvent 飞书事件内容
type FeishuEvent struct {
	Message FeishuMessageContent `json:"message"`
	Sender  FeishuSender         `json:"sender"`
}

// FeishuMessageContent 飞书消息内容
type FeishuMessageContent struct {
	MessageID   string                 `json:"message_id"`
	RootID      string                 `json:"root_id"`
	ParentID    string                 `json:"parent_id"`
	CreateTime  int64                  `json:"create_time"`
	UpdateTime  int64                  `json:"update_time"`
	MessageType string                 `json:"message_type"`
	Content     map[string]interface{} `json:"content"`
	ChatID      string                 `json:"chat_id"`
	ChatType    string                 `json:"chat_type"`
}

// FeishuSender 飞书发送者信息
type FeishuSender struct {
	SenderID   FeishuSenderID `json:"sender_id"`
	SenderType string         `json:"sender_type"`
	TenantKey  string         `json:"tenant_key"`
}

// FeishuSenderID 飞书发送者ID
type FeishuSenderID struct {
	UserID  string `json:"user_id"`
	UnionID string `json:"union_id"`
	OpenID  string `json:"open_id"`
}

// FeishuSendMessage 飞书发送消息结构
type FeishuSendMessage struct {
	ReceiveID string                 `json:"receive_id"`
	MsgType   string                 `json:"msg_type"`
	Content   map[string]interface{} `json:"content"`
}

// FeishuTextMessage 飞书文本消息
type FeishuTextMessage struct {
	Text string `json:"text"`
}

// FeishuImageMessage 飞书图片消息
type FeishuImageMessage struct {
	ImageKey string `json:"image_key"`
}

// FeishuPostMessage 飞书富文本消息
type FeishuPostMessage struct {
	ZhCn FeishuPostContent `json:"zh_cn,omitempty"`
	EnUs FeishuPostContent `json:"en_us,omitempty"`
}

// FeishuPostContent 飞书富文本内容
type FeishuPostContent struct {
	Title   string          `json:"title"`
	Content [][]interface{} `json:"content"`
}

// FeishuCardMessage 飞书卡片消息
type FeishuCardMessage struct {
	Card interface{} `json:"card"`
}

// FeishuFileMessage 飞书文件消息
type FeishuFileMessage struct {
	FileKey string `json:"file_key"`
}

// FeishuAudioMessage 飞书音频消息
type FeishuAudioMessage struct {
	FileKey  string `json:"file_key"`
	Duration int    `json:"duration"`
}

// FeishuVideoMessage 飞书视频消息
type FeishuVideoMessage struct {
	FileKey  string `json:"file_key"`
	Duration int    `json:"duration"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// FeishuWebhookRequest 飞书Webhook请求参数
type FeishuWebhookRequest struct {
	Timestamp string `header:"X-Lark-Request-Timestamp"`
	Nonce     string `header:"X-Lark-Request-Nonce"`
	Signature string `header:"X-Lark-Signature"`
}

// FeishuInteractiveCard 飞书交互式卡片
type FeishuInteractiveCard struct {
	Config   FeishuCardConfig    `json:"config"`
	Header   FeishuCardHeader    `json:"header"`
	Elements []FeishuCardElement `json:"elements"`
}

// FeishuCardConfig 飞书卡片配置
type FeishuCardConfig struct {
	WideScreenMode bool `json:"wide_screen_mode"`
	EnableForward  bool `json:"enable_forward"`
}

// FeishuCardHeader 飞书卡片头部
type FeishuCardHeader struct {
	Title FeishuCardTitle `json:"title"`
}

// FeishuCardTitle 飞书卡片标题
type FeishuCardTitle struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// FeishuCardElement 飞书卡片元素
type FeishuCardElement struct {
	Tag     string      `json:"tag"`
	Text    interface{} `json:"text,omitempty"`
	Content interface{} `json:"content,omitempty"`
	Fields  interface{} `json:"fields,omitempty"`
}
