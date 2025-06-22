// Package v1 implements routing paths. Each services in own file.
package http

import (
	"net/http"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "github.com/sivdead/OmniBotGo/docs" // Swagger docs.
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/controller/http/middleware"
	v1 "github.com/sivdead/OmniBotGo/internal/controller/http/v1"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// NewRouter -.
// Swagger spec:
// @title       OmniBotGo API
// @description 统一消息机器人管理平台API
// @version     1.0
// @host        localhost:8080
// @BasePath    /api/v1
func NewRouter(
	app *fiber.App,
	cfg *config.Config,
	messageUC usecase.MessageUseCase,
	channelUC usecase.ChannelUseCase,
	botUC usecase.BotUseCase,
	systemConfigUC usecase.SystemConfigUseCase,
	platformUC usecase.PlatformUseCase,
	monitorUC usecase.MonitorUseCase,
	logUC usecase.LogUseCase,
	queueUC usecase.QueueUseCase,
	processorUC usecase.ProcessorUseCase,
	l logger.Interface,
) {
	// Options
	app.Use(middleware.Logger(l))
	app.Use(middleware.Recovery(l))

	// Prometheus metrics
	if cfg.Metrics.Enabled {
		prometheus := fiberprometheus.New("my-service-name")
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	// Swagger
	if cfg.Swagger.Enabled {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// K8s probe
	app.Get("/healthz", func(ctx *fiber.Ctx) error { return ctx.SendStatus(http.StatusOK) })

	// 创建V1控制器实例（避免重复创建）
	v1Controller := v1.NewV1Controller(messageUC, channelUC, botUC, systemConfigUC, platformUC, monitorUC, logUC, queueUC, processorUC, l, cfg)

	// API路由
	apiV1Group := app.Group("/api/v1")
	{
		v1.NewAPIRoutes(apiV1Group, messageUC, channelUC, botUC, systemConfigUC, platformUC, monitorUC, logUC, queueUC, processorUC, l, cfg)
	}

	// Webhook路由（在根路径下）
	v1.NewWebhookRoutes(app, v1Controller)

	// 健康检查路由
	v1.NewHealthRoutes(app, v1Controller)
}
