package entity

// PlatformType 平台类型枚举
type PlatformType string

const (
	PlatformTypeWecom          PlatformType = "wecom"           // 企业微信
	PlatformTypeDingtalk       PlatformType = "dingtalk"        // 钉钉
	PlatformTypeWechatOfficial PlatformType = "wechat_official" // 微信公众号
	PlatformTypeFeishu         PlatformType = "feishu"          // 飞书
)
