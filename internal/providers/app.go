package providers

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/wire"
	"github.com/sivdead/OmniBotGo/pkg/database"
	"github.com/sivdead/OmniBotGo/pkg/grpcserver"
	"github.com/sivdead/OmniBotGo/pkg/httpserver"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"github.com/sivdead/OmniBotGo/pkg/rabbitmq/rmq_rpc/server"
)

// AppSet 包含应用程序的Provider
var AppSet = wire.NewSet(
	NewApp,
)

// App 应用程序主结构体
type App struct {
	HTTPServer *httpserver.Server
	GRPCServer *grpcserver.Server
	RMQServer  *server.Server
	Database   database.CommonDB
	Logger     logger.Interface
}

// NewApp 创建应用程序实例
func NewApp(
	httpSrv *httpserver.Server,
	grpcSrv *grpcserver.Server,
	rmqSrv *server.Server,
	db database.CommonDB,
	l logger.Interface,
) *App {
	return &App{
		HTTPServer: httpSrv,
		GRPCServer: grpcSrv,
		RMQServer:  rmqSrv,
		Database:   db,
		Logger:     l,
	}
}

// Run 启动应用程序
func (a *App) Run() {
	// 延迟关闭数据库连接
	defer func() {
		if err := a.Database.Close(); err != nil {
			a.Logger.Error(fmt.Errorf("app - Run - database.Close: %w", err))
		}
	}()

	// 启动所有服务器
	a.RMQServer.Start()
	a.GRPCServer.Start()
	a.HTTPServer.Start()

	// 等待中断信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		a.Logger.Info("app - Run - signal: %s", s.String())
	case err := <-a.HTTPServer.Notify():
		a.Logger.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	case err := <-a.GRPCServer.Notify():
		a.Logger.Error(fmt.Errorf("app - Run - grpcServer.Notify: %w", err))
	case err := <-a.RMQServer.Notify():
		a.Logger.Error(fmt.Errorf("app - Run - rmqServer.Notify: %w", err))
	}

	// 优雅关闭所有服务器
	a.shutdown()
}

// shutdown 优雅关闭所有服务器
func (a *App) shutdown() {
	if err := a.HTTPServer.Shutdown(); err != nil {
		a.Logger.Error(fmt.Errorf("app - shutdown - httpServer.Shutdown: %w", err))
	}

	if err := a.GRPCServer.Shutdown(); err != nil {
		a.Logger.Error(fmt.Errorf("app - shutdown - grpcServer.Shutdown: %w", err))
	}

	if err := a.RMQServer.Shutdown(); err != nil {
		a.Logger.Error(fmt.Errorf("app - shutdown - rmqServer.Shutdown: %w", err))
	}
}
