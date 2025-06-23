package providers

import (
	"github.com/google/wire"
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/controller/http"
	"github.com/sivdead/OmniBotGo/internal/service"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/grpcserver"
	"github.com/sivdead/OmniBotGo/pkg/httpserver"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"github.com/sivdead/OmniBotGo/pkg/rabbitmq/rmq_rpc/server"
)

// ServerSet 服务器相关的Provider集合
var ServerSet = wire.NewSet(
	NewHTTPServer,
	NewGRPCServer,
	NewRMQServer,
	NewConfigWatcher,
)

// NewHTTPServer 创建HTTP服务器并设置路由
func NewHTTPServer(
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
) *httpserver.Server {
	httpSrv := httpserver.New(
		httpserver.Port(cfg.HTTP.Port),
		httpserver.Prefork(cfg.HTTP.UsePreforkMode),
	)

	// 设置路由
	http.NewRouter(httpSrv.App, cfg, messageUC, channelUC, botUC, systemConfigUC, platformUC, monitorUC, logUC, queueUC, processorUC, l)

	return httpSrv
}

// NewGRPCServer 创建gRPC服务器（基础版本，暂时没有路由）
func NewGRPCServer(
	cfg *config.Config,
	l logger.Interface,
) *grpcserver.Server {
	grpcSrv := grpcserver.New(grpcserver.Port(cfg.GRPC.Port))

	// TODO: 后续添加OmniBotGo的路由
	// grpc.NewRouter(grpcSrv.App, useCases, l)

	return grpcSrv
}

// NewRMQServer 创建RabbitMQ RPC服务器（基础版本，暂时没有路由）
func NewRMQServer(
	cfg *config.Config,
	l logger.Interface,
) (*server.Server, error) {
	// 如果没有配置RMQ URL，返回nil（表示不启用RMQ服务器）
	if cfg.RMQ.URL == "" {
		l.Info("RabbitMQ URL not configured, RMQ server disabled")
		return nil, nil
	}

	// TODO: 后续添加OmniBotGo的路由
	// rmqRouter := amqprpc.NewRouter(useCases, l)

	return server.New(cfg.RMQ.URL, cfg.RMQ.ServerExchange, nil, l)
}

// NewConfigWatcher 创建配置监视器
func NewConfigWatcher(
	systemRepo port.SystemConfigRepository,
	channelRepo port.ChannelRepository,
	processorRepo port.MessageProcessorRepository,
	adapterManager port.AdapterManager,
	logger logger.Interface,
) *service.ConfigWatcher {
	return service.NewConfigWatcher(systemRepo, channelRepo, processorRepo, adapterManager, logger)
}
