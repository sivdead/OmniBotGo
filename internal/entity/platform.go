package entity

// PlatformType 平台类型枚举
type PlatformType string

const (
	PlatformTypeWecom              PlatformType = "wecom"               // 企业微信
	PlatformTypeDingtalk           PlatformType = "dingtalk"            // 钉钉
	PlatformTypeDingTalk           PlatformType = "dingtalk"            // 钉钉（别名）
	PlatformTypeDingTalkStream     PlatformType = "dingtalk_stream"     // 钉钉Stream
	PlatformTypeDingTalkEnterprise PlatformType = "dingtalk_enterprise" // 钉钉企业应用
	PlatformTypeWechatOfficial     PlatformType = "wechat_official"     // 微信公众号
	PlatformTypeFeishu             PlatformType = "feishu"              // 飞书
)
