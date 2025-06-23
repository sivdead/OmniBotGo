package providers

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/wire"
	"github.com/sivdead/OmniBotGo/internal/service"
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
	HTTPServer        *httpserver.Server
	GRPCServer        *grpcserver.Server
	RMQServer         *server.Server
	ConnectionManager *service.ConnectionManager
	Database          database.CommonDB
	Logger            logger.Interface
}

// NewApp 创建应用程序实例
func NewApp(
	httpSrv *httpserver.Server,
	grpcSrv *grpcserver.Server,
	rmqSrv *server.Server,
	connMgr *service.ConnectionManager,
	db database.CommonDB,
	l logger.Interface,
) *App {
	return &App{
		HTTPServer:        httpSrv,
		GRPCServer:        grpcSrv,
		RMQServer:         rmqSrv,
		ConnectionManager: connMgr,
		Database:          db,
		Logger:            l,
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

	// 启动ConnectionManager（在其他服务之前启动，确保Stream连接已就绪）
	if err := a.ConnectionManager.Start(context.Background()); err != nil {
		a.Logger.Error(fmt.Errorf("app - Run - connectionManager.Start: %w", err))
	}

	// 启动所有服务器
	if a.RMQServer != nil {
		a.RMQServer.Start()
	}
	a.GRPCServer.Start()
	a.HTTPServer.Start()

	a.Logger.Info("应用程序启动完成")

	// 等待中断信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// 创建RMQ通知channel，如果RMQServer为nil则创建一个永不触发的channel
	rmqNotify := make(<-chan error)
	if a.RMQServer != nil {
		rmqNotify = a.RMQServer.Notify()
	}

	select {
	case s := <-interrupt:
		a.Logger.Info("app - Run - signal: %s", s.String())
	case err := <-a.HTTPServer.Notify():
		a.Logger.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	case err := <-a.GRPCServer.Notify():
		a.Logger.Error(fmt.Errorf("app - Run - grpcServer.Notify: %w", err))
	case err := <-rmqNotify:
		if err != nil {
			a.Logger.Error(fmt.Errorf("app - Run - rmqServer.Notify: %w", err))
		}
	}

	// 优雅关闭所有服务器
	a.shutdown()
}

// shutdown 优雅关闭所有服务器
func (a *App) shutdown() {
	a.Logger.Info("开始优雅关闭应用程序")

	// 先关闭ConnectionManager（Stream连接）
	if err := a.ConnectionManager.Stop(context.Background()); err != nil {
		a.Logger.Error(fmt.Errorf("app - shutdown - connectionManager.Stop: %w", err))
	}

	// 然后关闭其他服务器
	if err := a.HTTPServer.Shutdown(); err != nil {
		a.Logger.Error(fmt.Errorf("app - shutdown - httpServer.Shutdown: %w", err))
	}

	if err := a.GRPCServer.Shutdown(); err != nil {
		a.Logger.Error(fmt.Errorf("app - shutdown - grpcServer.Shutdown: %w", err))
	}

	if a.RMQServer != nil {
		if err := a.RMQServer.Shutdown(); err != nil {
			a.Logger.Error(fmt.Errorf("app - shutdown - rmqServer.Shutdown: %w", err))
		}
	}

	a.Logger.Info("应用程序优雅关闭完成")
}
