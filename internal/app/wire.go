//go:build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/sivdead/OmniBotGo/config"
	"github.com/sivdead/OmniBotGo/internal/providers"
)

// InitializeApp 使用Wire进行依赖注入，初始化整个应用程序
func InitializeApp(cfg *config.Config) (*providers.App, error) {
	wire.Build(
		providers.InfrastructureSet,
		providers.ServerSet,
		providers.AppSet,
		// TODO: 后续当服务器路由集成UseCase时，重新添加以下provider sets:
		// providers.RepositorySet,
		// providers.UseCaseSet,
	)
	return &providers.App{}, nil
}
