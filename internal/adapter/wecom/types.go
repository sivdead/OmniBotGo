package wecom

// WecomMessage 企业微信接收消息结构
type WecomMessage struct {
	ToUserName   string `json:"ToUserName" xml:"ToUserName"`
	FromUserName string `json:"FromUserName" xml:"FromUserName"`
	CreateTime   int64  `json:"CreateTime" xml:"CreateTime"`
	MsgType      string `json:"MsgType" xml:"MsgType"`
	MsgID        int64  `json:"MsgId" xml:"MsgId"`
	AgentID      int    `json:"AgentID" xml:"AgentID"`

	// 文本消息
	Content string `json:"Content,omitempty" xml:"Content,omitempty"`

	// 图片消息
	PicURL  string `json:"PicUrl,omitempty" xml:"PicUrl,omitempty"`
	MediaID string `json:"MediaId,omitempty" xml:"MediaId,omitempty"`

	// 语音消息
	Format string `json:"Format,omitempty" xml:"Format,omitempty"`

	// 视频消息
	ThumbMediaID string `json:"ThumbMediaId,omitempty" xml:"ThumbMediaId,omitempty"`

	// 位置消息
	LocationX string `json:"Location_X,omitempty" xml:"Location_X,omitempty"`
	LocationY string `json:"Location_Y,omitempty" xml:"Location_Y,omitempty"`
	Scale     string `json:"Scale,omitempty" xml:"Scale,omitempty"`
	Label     string `json:"Label,omitempty" xml:"Label,omitempty"`

	// 链接消息
	Title       string `json:"Title,omitempty" xml:"Title,omitempty"`
	Description string `json:"Description,omitempty" xml:"Description,omitempty"`
	URL         string `json:"Url,omitempty" xml:"Url,omitempty"`

	// 事件消息
	Event    string `json:"Event,omitempty" xml:"Event,omitempty"`
	EventKey string `json:"EventKey,omitempty" xml:"EventKey,omitempty"`
}

// WecomSendMessage 企业微信发送消息结构
type WecomSendMessage struct {
	ToUser                 string                  `json:"touser"`
	ToParty                string                  `json:"toparty,omitempty"`
	ToTag                  string                  `json:"totag,omitempty"`
	MsgType                string                  `json:"msgtype"`
	AgentID                string                  `json:"agentid"`
	Text                   *WecomTextMessage       `json:"text,omitempty"`
	Image                  *WecomImageMessage      `json:"image,omitempty"`
	Voice                  *WecomVoiceMessage      `json:"voice,omitempty"`
	Video                  *WecomVideoMessage      `json:"video,omitempty"`
	File                   *WecomFileMessage       `json:"file,omitempty"`
	TextCard               *WecomTextCardMessage   `json:"textcard,omitempty"`
	News                   *WecomNewsMessage       `json:"news,omitempty"`
	MPNews                 *WecomMPNewsMessage     `json:"mpnews,omitempty"`
	Markdown               *WecomMarkdownMessage   `json:"markdown,omitempty"`
	MiniProgramNotice      *WecomMiniProgramNotice `json:"miniprogram_notice,omitempty"`
	TemplateCard           *WecomTemplateCard      `json:"template_card,omitempty"`
	Safe                   int                     `json:"safe,omitempty"`
	EnableIDTrans          int                     `json:"enable_id_trans,omitempty"`
	EnableDuplicateCheck   int                     `json:"enable_duplicate_check,omitempty"`
	DuplicateCheckInterval int                     `json:"duplicate_check_interval,omitempty"`
}

// WecomTextMessage 文本消息
type WecomTextMessage struct {
	Content string `json:"content"`
}

// WecomImageMessage 图片消息
type WecomImageMessage struct {
	MediaID string `json:"media_id"`
}

// WecomVoiceMessage 语音消息
type WecomVoiceMessage struct {
	MediaID string `json:"media_id"`
}

// WecomVideoMessage 视频消息
type WecomVideoMessage struct {
	MediaID     string `json:"media_id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// WecomFileMessage 文件消息
type WecomFileMessage struct {
	MediaID string `json:"media_id"`
}

// WecomTextCardMessage 文本卡片消息
type WecomTextCardMessage struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	BtnTxt      string `json:"btntxt,omitempty"`
}

// WecomNewsMessage 图文消息
type WecomNewsMessage struct {
	Articles []WecomArticle `json:"articles"`
}

// WecomArticle 图文消息文章
type WecomArticle struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl,omitempty"`
}

// WecomMPNewsMessage 图文消息（mpnews类型）
type WecomMPNewsMessage struct {
	Articles []WecomMPArticle `json:"articles"`
}

// WecomMPArticle MP图文消息文章
type WecomMPArticle struct {
	Title            string `json:"title"`
	ThumbMediaID     string `json:"thumb_media_id"`
	Author           string `json:"author,omitempty"`
	ContentSourceURL string `json:"content_source_url,omitempty"`
	Content          string `json:"content"`
	Digest           string `json:"digest,omitempty"`
}

// WecomMarkdownMessage Markdown消息
type WecomMarkdownMessage struct {
	Content string `json:"content"`
}

// WecomMiniProgramNotice 小程序通知消息
type WecomMiniProgramNotice struct {
	AppID             string                              `json:"appid"`
	Page              string                              `json:"page,omitempty"`
	Title             string                              `json:"title"`
	Description       string                              `json:"description,omitempty"`
	EmphasisFirstItem bool                                `json:"emphasis_first_item,omitempty"`
	ContentItem       []WecomMiniProgramNoticeContentItem `json:"content_item,omitempty"`
}

// WecomMiniProgramNoticeContentItem 小程序通知内容项
type WecomMiniProgramNoticeContentItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// WecomTemplateCard 模板卡片消息
type WecomTemplateCard struct {
	CardType              string                        `json:"card_type"`
	Source                *WecomTemplateCardSource      `json:"source,omitempty"`
	MainTitle             *WecomTemplateCardMainTitle   `json:"main_title,omitempty"`
	EmphasisContent       *WecomTemplateCardEmphasis    `json:"emphasis_content,omitempty"`
	QuoteArea             *WecomTemplateCardQuoteArea   `json:"quote_area,omitempty"`
	SubTitleText          string                        `json:"sub_title_text,omitempty"`
	HorizontalContentList []WecomTemplateCardHorizontal `json:"horizontal_content_list,omitempty"`
	JumpList              []WecomTemplateCardJump       `json:"jump_list,omitempty"`
	CardAction            *WecomTemplateCardAction      `json:"card_action,omitempty"`
}

// WecomTemplateCardSource 模板卡片来源
type WecomTemplateCardSource struct {
	IconURL   string `json:"icon_url,omitempty"`
	Desc      string `json:"desc,omitempty"`
	DescColor int    `json:"desc_color,omitempty"`
}

// WecomTemplateCardMainTitle 模板卡片主标题
type WecomTemplateCardMainTitle struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
}

// WecomTemplateCardEmphasis 模板卡片强调内容
type WecomTemplateCardEmphasis struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
}

// WecomTemplateCardQuoteArea 模板卡片引用区域
type WecomTemplateCardQuoteArea struct {
	Type      int    `json:"type,omitempty"`
	URL       string `json:"url,omitempty"`
	AppID     string `json:"appid,omitempty"`
	PagePath  string `json:"pagepath,omitempty"`
	Title     string `json:"title,omitempty"`
	QuoteText string `json:"quote_text,omitempty"`
}

// WecomTemplateCardHorizontal 模板卡片水平内容
type WecomTemplateCardHorizontal struct {
	KeyName string `json:"keyname"`
	Value   string `json:"value"`
	Type    int    `json:"type,omitempty"`
	URL     string `json:"url,omitempty"`
	MediaID string `json:"media_id,omitempty"`
}

// WecomTemplateCardJump 模板卡片跳转
type WecomTemplateCardJump struct {
	Type     int    `json:"type"`
	URL      string `json:"url,omitempty"`
	Title    string `json:"title"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// WecomTemplateCardAction 模板卡片行动
type WecomTemplateCardAction struct {
	Type     int    `json:"type"`
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// WecomEventMessage 事件消息
type WecomEventMessage struct {
	ToUserName   string `json:"ToUserName" xml:"ToUserName"`
	FromUserName string `json:"FromUserName" xml:"FromUserName"`
	CreateTime   int64  `json:"CreateTime" xml:"CreateTime"`
	MsgType      string `json:"MsgType" xml:"MsgType"`
	Event        string `json:"Event" xml:"Event"`
	EventKey     string `json:"EventKey,omitempty" xml:"EventKey,omitempty"`
	AgentID      int    `json:"AgentID" xml:"AgentID"`
}

// WecomCallbackRequest Webhook回调请求
type WecomCallbackRequest struct {
	MsgSignature string `form:"msg_signature" query:"msg_signature"`
	Timestamp    string `form:"timestamp" query:"timestamp"`
	Nonce        string `form:"nonce" query:"nonce"`
	EchoStr      string `form:"echostr" query:"echostr"`
}

// WecomResponse 企业微信API响应
type WecomResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
