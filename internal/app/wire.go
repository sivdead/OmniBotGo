//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/providers"
)

// InitializeApp 使用Wire进行依赖注入，初始化整个应用程序
func InitializeApp(
	cfg *config.Config,
) (*providers.App, func(), error) {
	panic(wire.Build(
		providers.InfrastructureSet,
		providers.RepositorySet,
		providers.AdapterSet,
		providers.UseCaseSet,
		providers.ServerSet,
		providers.NewApp,
	))
}
