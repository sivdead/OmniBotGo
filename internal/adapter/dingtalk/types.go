package dingtalk

// DingtalkMessage 钉钉接收消息结构
type DingtalkMessage struct {
	MsgType                   string `json:"msgtype"`
	MsgID                     string `json:"msgId"`
	CreateAt                  int64  `json:"createAt"`
	ConversationType          string `json:"conversationType"`
	ConversationID            string `json:"conversationId"`
	ChatbotUserID             string `json:"chatbotUserId"`
	SenderID                  string `json:"senderId"`
	SenderNick                string `json:"senderNick"`
	SenderCorpID              string `json:"senderCorpId"`
	SenderStaffID             string `json:"senderStaffId"`
	SessionWebhook            string `json:"sessionWebhook"`
	SessionWebhookExpiredTime int64  `json:"sessionWebhookExpiredTime"`
	IsAdmin                   bool   `json:"isAdmin"`
	IsInAtList                bool   `json:"isInAtList"`

	// 文本消息
	Text *DingtalkTextContent `json:"text,omitempty"`

	// 通用内容字段
	Content *DingtalkContent `json:"content,omitempty"`
}

// DingtalkTextContent 钉钉文本消息内容
type DingtalkTextContent struct {
	Content string `json:"content"`
}

// DingtalkContent 钉钉消息内容
type DingtalkContent struct {
	// 图片消息
	PicURL       string `json:"picUrl,omitempty"`
	DownloadCode string `json:"downloadCode,omitempty"`

	// 媒体消息
	MediaID  string `json:"mediaId,omitempty"`
	FileName string `json:"fileName,omitempty"`
	FileSize int64  `json:"fileSize,omitempty"`
	Duration int    `json:"duration,omitempty"`

	// 链接消息
	Title      string `json:"title,omitempty"`
	Text       string `json:"text,omitempty"`
	MessageURL string `json:"messageUrl,omitempty"`
}

// DingtalkSendMessage 钉钉发送消息结构
type DingtalkSendMessage struct {
	MsgType    string                     `json:"msgtype"`
	Text       *DingtalkTextMessage       `json:"text,omitempty"`
	Link       *DingtalkLinkMessage       `json:"link,omitempty"`
	Markdown   *DingtalkMarkdownMessage   `json:"markdown,omitempty"`
	ActionCard *DingtalkActionCardMessage `json:"actionCard,omitempty"`
	FeedCard   *DingtalkFeedCardMessage   `json:"feedCard,omitempty"`
	At         *DingtalkAtMessage         `json:"at,omitempty"`
}

// DingtalkTextMessage 钉钉文本消息
type DingtalkTextMessage struct {
	Content string `json:"content"`
}

// DingtalkLinkMessage 钉钉链接消息
type DingtalkLinkMessage struct {
	Text       string `json:"text"`
	Title      string `json:"title"`
	PicURL     string `json:"picUrl,omitempty"`
	MessageURL string `json:"messageUrl"`
}

// DingtalkMarkdownMessage 钉钉Markdown消息
type DingtalkMarkdownMessage struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// DingtalkActionCardMessage 钉钉ActionCard消息
type DingtalkActionCardMessage struct {
	Title          string                     `json:"title"`
	Text           string                     `json:"text"`
	BtnOrientation string                     `json:"btnOrientation,omitempty"`
	SingleTitle    string                     `json:"singleTitle,omitempty"`
	SingleURL      string                     `json:"singleURL,omitempty"`
	Btns           []DingtalkActionCardButton `json:"btns,omitempty"`
}

// DingtalkActionCardButton ActionCard按钮
type DingtalkActionCardButton struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

// DingtalkFeedCardMessage 钉钉FeedCard消息
type DingtalkFeedCardMessage struct {
	Links []DingtalkFeedCardLink `json:"links"`
}

// DingtalkFeedCardLink FeedCard链接
type DingtalkFeedCardLink struct {
	Title      string `json:"title"`
	MessageURL string `json:"messageURL"`
	PicURL     string `json:"picURL,omitempty"`
}

// DingtalkAtMessage 钉钉@消息配置
type DingtalkAtMessage struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	AtUserIds []string `json:"atUserIds,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

// DingtalkWebhookRequest 钉钉Webhook请求参数
type DingtalkWebhookRequest struct {
	Timestamp string `form:"timestamp" query:"timestamp"`
	Sign      string `form:"sign" query:"sign"`
}

// DingtalkEventMessage 钉钉事件消息
type DingtalkEventMessage struct {
	MsgType   string                 `json:"msgtype"`
	EventType string                 `json:"eventType"`
	TimeStamp int64                  `json:"timeStamp"`
	Data      map[string]interface{} `json:"data"`
}
