package providers

import (
	"os"

	"github.com/google/wire"
	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/adapter"
	"github.com/sivdead/OmniBotGo/internal/adapter/dingtalk_enterprise"
	"github.com/sivdead/OmniBotGo/internal/adapter/dingtalk_proxy"
	"github.com/sivdead/OmniBotGo/internal/adapter/dingtalk_stream"
	"github.com/sivdead/OmniBotGo/internal/adapter/feishu"
	"github.com/sivdead/OmniBotGo/internal/adapter/wechat_official"
	"github.com/sivdead/OmniBotGo/internal/adapter/wecom"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

// AdapterSet 包含所有适配器的Provider
var AdapterSet = wire.NewSet(
	NewAdapterManager,
	NewAdapterRegistry,
	// 各平台适配器
	NewWecomAdapter,
	NewDingtalkStreamAdapter,
	NewDingtalkEnterpriseAdapter,
	NewDingtalkProxyAdapter,
	NewWechatOfficialAdapter,
	NewFeishuAdapter,
	// 绑定接口和实现
	wire.Bind(new(port.AdapterManager), new(*adapter.Manager)),
)

// AdapterRegistry 适配器注册表类型
type AdapterRegistry map[entity.PlatformType]interface{}

// NewWecomAdapter 创建企业微信适配器
func NewWecomAdapter() *wecom.WecomAdapter {
	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	return wecom.NewWecomAdapter(logger)
}

// NewDingtalkStreamAdapter 创建钉钉Stream适配器
func NewDingtalkStreamAdapter() *dingtalk_stream.DingtalkStreamAdapter {
	// 创建zerolog logger
	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	return dingtalk_stream.NewAdapter(logger)
}

// NewDingtalkEnterpriseAdapter 创建钉钉企业应用适配器
func NewDingtalkEnterpriseAdapter() *dingtalk_enterprise.DingtalkEnterpriseAdapter {
	// 创建zerolog logger
	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	return dingtalk_enterprise.NewAdapter(logger)
}

// NewDingtalkProxyAdapter 创建钉钉代理适配器
func NewDingtalkProxyAdapter(
	streamAdapter *dingtalk_stream.DingtalkStreamAdapter,
	enterpriseAdapter *dingtalk_enterprise.DingtalkEnterpriseAdapter,
) *dingtalk_proxy.DingtalkProxyAdapter {
	// 创建zerolog logger
	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	return dingtalk_proxy.NewAdapter(logger, streamAdapter, enterpriseAdapter)
}

// NewWechatOfficialAdapter 创建微信公众号适配器
func NewWechatOfficialAdapter() *wechat_official.WechatOfficialAdapter {
	return wechat_official.NewWechatOfficialAdapter()
}

// NewFeishuAdapter 创建飞书适配器
func NewFeishuAdapter() *feishu.FeishuAdapter {
	return feishu.NewFeishuAdapter()
}

// NewAdapterRegistry 创建适配器注册表
func NewAdapterRegistry(
	wecom *wecom.WecomAdapter,
	dingtalkProxy *dingtalk_proxy.DingtalkProxyAdapter,
	wechat *wechat_official.WechatOfficialAdapter,
	feishu *feishu.FeishuAdapter,
) AdapterRegistry {
	registry := make(AdapterRegistry)

	// 注册各平台适配器
	registry[entity.PlatformTypeWecom] = wecom
	registry[entity.PlatformTypeDingtalk] = dingtalkProxy // 使用代理适配器
	registry[entity.PlatformTypeWechatOfficial] = wechat
	registry[entity.PlatformTypeFeishu] = feishu

	return registry
}

// NewAdapterManager 创建适配器管理器
func NewAdapterManager(registry AdapterRegistry) *adapter.Manager {
	// 转换AdapterRegistry到map[entity.PlatformType]interface{}
	adapters := make(map[entity.PlatformType]interface{})
	for k, v := range registry {
		adapters[k] = v
	}
	return adapter.NewManagerWithRegistry(adapters)
}
