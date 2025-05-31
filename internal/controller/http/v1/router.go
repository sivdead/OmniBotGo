package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// NewTranslationRoutes -.
func NewTranslationRoutes(apiV1Group fiber.Router, t usecase.Translation, l logger.Interface) {
	r := &V1{t: t, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	translationGroup := apiV1Group.Group("/translation")

	{
		translationGroup.Get("/history", r.history)
		translationGroup.Post("/do-translate", r.doTranslate)
	}
}

// NewAPIRoutes 设置所有API路由
func NewAPIRoutes(
	apiV1Group fiber.Router,
	messageUC usecase.MessageUseCase,
	channelUC usecase.ChannelUseCase,
	botUC usecase.BotUseCase,
	webhookUC usecase.WebhookUseCase,
	t usecase.Translation,
	l logger.Interface,
) {
	v1 := NewV1Controller(t, messageUC, channelUC, botUC, webhookUC, l)

	// 消息相关路由
	messagesGroup := apiV1Group.Group("/messages")
	{
		messagesGroup.Post("/send", v1.SendMessage)
		messagesGroup.Get("/history", v1.GetMessageHistory)
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
}

// NewWebhookRoutes 设置webhook路由
func NewWebhookRoutes(
	app fiber.Router,
	webhookUC usecase.WebhookUseCase,
	l logger.Interface,
) {
	v1 := &V1{webhookUC: webhookUC, l: l, v: validator.New()}

	// Webhook路由（不在/api/v1下）
	webhookGroup := app.Group("/webhook")
	{
		webhookGroup.Get("/:platform/:channel_id", v1.VerifyWebhook)
		webhookGroup.Post("/:platform/:channel_id", v1.ProcessWebhook)
	}
}
