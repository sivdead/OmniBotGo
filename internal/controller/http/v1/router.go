package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/adapter"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// NewAPIRoutes 设置所有API路由
func NewAPIRoutes(
	apiV1Group fiber.Router,
	messageUC usecase.MessageUseCase,
	channelUC usecase.ChannelUseCase,
	botUC usecase.BotUseCase,
	l logger.Interface,
) {
	v1 := NewV1Controller(messageUC, channelUC, botUC, l)

	// 消息相关路由
	messagesGroup := apiV1Group.Group("/messages")
	{
		messagesGroup.Post("/send", v1.SendMessage)
		messagesGroup.Get("/", v1.GetMessageHistory)
		messagesGroup.Get("/:id", v1.GetMessage)
		messagesGroup.Post("/:id/retry", v1.RetryFailedMessage)
	}

	// 通道相关路由
	channelsGroup := apiV1Group.Group("/channels")
	{
		channelsGroup.Post("", v1.CreateChannel)
		channelsGroup.Get("", v1.GetChannels)
		channelsGroup.Get("/:id", v1.GetChannelByID)
		channelsGroup.Put("/:id", v1.UpdateChannel)
		channelsGroup.Delete("/:id", v1.DeleteChannel)
		channelsGroup.Patch("/:id/status", v1.UpdateChannelStatus)
		channelsGroup.Post("/:id/refresh-token", v1.RefreshChannelToken)
	}

	// 机器人相关路由
	botsGroup := apiV1Group.Group("/bots")
	{
		botsGroup.Post("", v1.CreateBot)
		botsGroup.Get("", v1.GetBots)
		botsGroup.Get("/:id", v1.GetBotByID)
		botsGroup.Put("/:id", v1.UpdateBot)
		botsGroup.Delete("/:id", v1.DeleteBot)
	}

	// 日志管理路由
	logsGroup := apiV1Group.Group("/logs")
	{
		// 连接日志
		logsGroup.Get("/connections", v1.GetConnectionLogs)
		logsGroup.Get("/connections/:id", v1.GetConnectionLogByID)

		// API调用日志
		logsGroup.Get("/api-calls", v1.GetAPICallLogs)
		logsGroup.Get("/api-calls/:id", v1.GetAPICallLogByID)
		logsGroup.Get("/api-calls/stats", v1.GetAPICallStats)
	}

	// 系统配置路由
	configsGroup := apiV1Group.Group("/configs")
	{
		configsGroup.Get("", v1.GetSystemConfigs)
		configsGroup.Get("/:key", v1.GetSystemConfigByKey)
		configsGroup.Post("", v1.CreateSystemConfig)
		configsGroup.Put("/:key", v1.UpdateSystemConfig)
		configsGroup.Delete("/:key", v1.DeleteSystemConfig)
		configsGroup.Get("/groups/:group", v1.GetSystemConfigsByGroup)
	}

	// 处理器管理路由
	processorsGroup := apiV1Group.Group("/processors")
	{
		processorsGroup.Post("", v1.CreateProcessor)
		processorsGroup.Get("", v1.GetProcessors)
		processorsGroup.Get("/:id", v1.GetProcessorByID)
		processorsGroup.Put("/:id", v1.UpdateProcessor)
		processorsGroup.Delete("/:id", v1.DeleteProcessor)
		processorsGroup.Patch("/:id/status", v1.UpdateProcessorStatus)

		// 路由规则管理
		processorsGroup.Post("/:id/rules", v1.CreateRoutingRule)
		processorsGroup.Get("/:id/rules", v1.GetRoutingRules)
		processorsGroup.Put("/:id/rules/:rule_id", v1.UpdateRoutingRule)
		processorsGroup.Delete("/:id/rules/:rule_id", v1.DeleteRoutingRule)
	}

	// 队列管理路由
	queuesGroup := apiV1Group.Group("/queues")
	{
		queuesGroup.Get("", v1.GetQueueMessages)
		queuesGroup.Get("/:id", v1.GetQueueMessageByID)
		queuesGroup.Post("/:id/retry", v1.RetryQueueMessage)
		queuesGroup.Patch("/:id/cancel", v1.CancelQueueMessage)
		queuesGroup.Get("/stats", v1.GetQueueStats)
		queuesGroup.Delete("/failed", v1.ClearFailedMessages)
	}

	// 平台管理路由（如果需要动态配置平台）
	platformsGroup := apiV1Group.Group("/platforms")
	{
		platformsGroup.Get("", v1.GetPlatforms)
		platformsGroup.Get("/:type", v1.GetPlatformByType)
		platformsGroup.Post("/:type/validate", v1.ValidatePlatformConfig)
		platformsGroup.Get("/:type/status", v1.GetPlatformStatus)
	}

	// 监控和统计路由
	monitorGroup := apiV1Group.Group("/monitor")
	{
		monitorGroup.Get("/overview", v1.GetSystemOverview)
		monitorGroup.Get("/metrics", v1.GetSystemMetrics)
		monitorGroup.Get("/health-detailed", v1.GetDetailedHealth)
	}
}

// NewWebhookRoutes 设置Webhook路由（在根路由层调用）
func NewWebhookRoutes(
	app fiber.Router,
	messageUC usecase.MessageUseCase,
	channelUC usecase.ChannelUseCase,
	botUC usecase.BotUseCase,
	l logger.Interface,
) {
	v1 := NewV1Controller(messageUC, channelUC, botUC, l)
	adapterManager := adapter.NewManager()
	webhookController := NewWebhookController(v1, adapterManager)

	// Webhook路由
	webhookGroup := app.Group("/webhook")
	{
		webhookGroup.Get("/:platform/:channel_id", webhookController.GetWebhookInfo)
		webhookGroup.Post("/:platform/:channel_id", webhookController.HandleWebhook)
	}
}

// NewHealthRoutes 设置健康检查路由（在根路由层调用）
func NewHealthRoutes(
	app fiber.Router,
	messageUC usecase.MessageUseCase,
	channelUC usecase.ChannelUseCase,
	botUC usecase.BotUseCase,
	l logger.Interface,
) {
	v1 := NewV1Controller(messageUC, channelUC, botUC, l)
	adapterManager := adapter.NewManager()
	healthController := NewHealthController(v1, adapterManager)

	// 健康检查路由
	app.Get("/healthz", healthController.GetHealth)
	app.Get("/health", healthController.GetHealthSimple)
	app.Get("/ready", healthController.GetReadiness)
	app.Get("/live", healthController.GetLiveness)
}
