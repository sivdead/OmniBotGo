//go:build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/providers"
)

// InitializeApp 使用Wire进行依赖注入，初始化整个应用程序
func InitializeApp(cfg *config.Config) (*providers.App, error) {
	wire.Build(
		providers.InfrastructureSet,
		providers.AdapterSet,
		providers.RepositorySet,
		providers.UseCaseSet,
		providers.ServerSet,
		providers.AppSet,
	)
	return &providers.App{}, nil
}
