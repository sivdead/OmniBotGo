package providers

import (
	"github.com/google/wire"
	"github.com/sivdead/OmniBotGo/config"
	"github.com/sivdead/OmniBotGo/pkg/grpcserver"
	"github.com/sivdead/OmniBotGo/pkg/httpserver"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"github.com/sivdead/OmniBotGo/pkg/rabbitmq/rmq_rpc/server"
)

// ServerSet 包含所有服务器相关的Provider
var ServerSet = wire.NewSet(
	NewHTTPServer,
	NewGRPCServer,
	NewRMQServer,
)

// NewHTTPServer 创建HTTP服务器（基础版本，暂时没有路由）
func NewHTTPServer(
	cfg *config.Config,
	l logger.Interface,
) *httpserver.Server {
	httpSrv := httpserver.New(
		httpserver.Port(cfg.HTTP.Port),
		httpserver.Prefork(cfg.HTTP.UsePreforkMode),
	)

	// TODO: 后续添加OmniBotGo的路由
	// http.NewRouter(httpSrv.App, cfg, useCases, l)

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
	// TODO: 后续添加OmniBotGo的路由
	// rmqRouter := amqprpc.NewRouter(useCases, l)

	return server.New(cfg.RMQ.URL, cfg.RMQ.ServerExchange, nil, l)
}
