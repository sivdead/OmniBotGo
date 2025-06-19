package wechat_official

// WechatOfficialMessage 微信公众号接收消息结构
type WechatOfficialMessage struct {
	ToUserName   string `json:"ToUserName" xml:"ToUserName"`
	FromUserName string `json:"FromUserName" xml:"FromUserName"`
	CreateTime   int64  `json:"CreateTime" xml:"CreateTime"`
	MsgType      string `json:"MsgType" xml:"MsgType"`
	MsgID        int64  `json:"MsgId" xml:"MsgId"`

	// 文本消息
	Content string `json:"Content,omitempty" xml:"Content,omitempty"`

	// 图片消息
	PicURL  string `json:"PicUrl,omitempty" xml:"PicUrl,omitempty"`
	MediaID string `json:"MediaId,omitempty" xml:"MediaId,omitempty"`

	// 语音消息
	Format      string `json:"Format,omitempty" xml:"Format,omitempty"`
	Recognition string `json:"Recognition,omitempty" xml:"Recognition,omitempty"`

	// 视频消息
	ThumbMediaID string `json:"ThumbMediaId,omitempty" xml:"ThumbMediaId,omitempty"`

	// 小视频消息（同视频消息）

	// 地理位置消息
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

	// 菜单事件
	MenuID string `json:"MenuId,omitempty" xml:"MenuId,omitempty"`

	// 扫码事件
	Ticket string `json:"Ticket,omitempty" xml:"Ticket,omitempty"`

	// 上报地理位置事件
	Latitude  string `json:"Latitude,omitempty" xml:"Latitude,omitempty"`
	Longitude string `json:"Longitude,omitempty" xml:"Longitude,omitempty"`
	Precision string `json:"Precision,omitempty" xml:"Precision,omitempty"`
}

// WechatOfficialSendMessage 微信公众号发送消息结构
type WechatOfficialSendMessage struct {
	ToUser        string                            `json:"touser"`
	MsgType       string                            `json:"msgtype"`
	Text          *WechatOfficialTextMessage        `json:"text,omitempty"`
	Image         *WechatOfficialImageMessage       `json:"image,omitempty"`
	Voice         *WechatOfficialVoiceMessage       `json:"voice,omitempty"`
	Video         *WechatOfficialVideoMessage       `json:"video,omitempty"`
	Music         *WechatOfficialMusicMessage       `json:"music,omitempty"`
	News          *WechatOfficialNewsMessage        `json:"news,omitempty"`
	MPNews        *WechatOfficialMPNewsMessage      `json:"mpnews,omitempty"`
	WxCard        *WechatOfficialWxCardMessage      `json:"wxcard,omitempty"`
	MiniProgram   *WechatOfficialMiniProgramMessage `json:"miniprogrampage,omitempty"`
	MsgMenu       *WechatOfficialMsgMenuMessage     `json:"msgmenu,omitempty"`
	CustomService *WechatOfficialCustomService      `json:"customservice,omitempty"`
}

// WechatOfficialTextMessage 文本消息
type WechatOfficialTextMessage struct {
	Content string `json:"content"`
}

// WechatOfficialImageMessage 图片消息
type WechatOfficialImageMessage struct {
	MediaID string `json:"media_id"`
}

// WechatOfficialVoiceMessage 语音消息
type WechatOfficialVoiceMessage struct {
	MediaID string `json:"media_id"`
}

// WechatOfficialVideoMessage 视频消息
type WechatOfficialVideoMessage struct {
	MediaID      string `json:"media_id"`
	ThumbMediaID string `json:"thumb_media_id,omitempty"`
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
}

// WechatOfficialMusicMessage 音乐消息
type WechatOfficialMusicMessage struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	MusicURL     string `json:"musicurl"`
	HQMusicURL   string `json:"hqmusicurl"`
	ThumbMediaID string `json:"thumb_media_id"`
}

// WechatOfficialNewsMessage 图文消息
type WechatOfficialNewsMessage struct {
	Articles []WechatOfficialArticle `json:"articles"`
}

// WechatOfficialArticle 图文消息文章
type WechatOfficialArticle struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl"`
}

// WechatOfficialMPNewsMessage 图文消息（素材库）
type WechatOfficialMPNewsMessage struct {
	MediaID string `json:"media_id"`
}

// WechatOfficialWxCardMessage 卡券消息
type WechatOfficialWxCardMessage struct {
	CardID string `json:"card_id"`
}

// WechatOfficialMiniProgramMessage 小程序消息
type WechatOfficialMiniProgramMessage struct {
	Title        string `json:"title"`
	AppID        string `json:"appid"`
	PagePath     string `json:"pagepath"`
	ThumbMediaID string `json:"thumb_media_id"`
}

// WechatOfficialMsgMenuMessage 菜单消息
type WechatOfficialMsgMenuMessage struct {
	HeadContent string                          `json:"head_content"`
	List        []WechatOfficialMsgMenuListItem `json:"list"`
	TailContent string                          `json:"tail_content"`
}

// WechatOfficialMsgMenuListItem 菜单消息列表项
type WechatOfficialMsgMenuListItem struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

// WechatOfficialCustomService 客服账号
type WechatOfficialCustomService struct {
	KfAccount string `json:"kf_account"`
}

// WechatOfficialResponse 微信公众号API响应
type WechatOfficialResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// WechatOfficialTokenResponse 微信公众号Token响应
type WechatOfficialTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// WechatOfficialEventMessage 事件消息
type WechatOfficialEventMessage struct {
	ToUserName   string `json:"ToUserName" xml:"ToUserName"`
	FromUserName string `json:"FromUserName" xml:"FromUserName"`
	CreateTime   int64  `json:"CreateTime" xml:"CreateTime"`
	MsgType      string `json:"MsgType" xml:"MsgType"`
	Event        string `json:"Event" xml:"Event"`
	EventKey     string `json:"EventKey,omitempty" xml:"EventKey,omitempty"`
}

// WechatOfficialCallbackRequest Webhook回调请求
type WechatOfficialCallbackRequest struct {
	Signature string `form:"signature" query:"signature"`
	Timestamp string `form:"timestamp" query:"timestamp"`
	Nonce     string `form:"nonce" query:"nonce"`
	EchoStr   string `form:"echostr" query:"echostr"`
}

// WechatOfficialConfig 微信公众号配置
type WechatOfficialConfig struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
	Token     string `json:"token"`
	AESKey    string `json:"aes_key,omitempty"`
}
