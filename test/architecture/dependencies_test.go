package architecture_test

import (
	"go/build"
	"strings"
	"testing"
)

// 定义架构层级和允许的依赖规则
var (
	// 层级定义
	entityLayer     = "github.com/sivdead/OmniBotGo/internal/entity"
	dtoLayer        = "github.com/sivdead/OmniBotGo/internal/dto"
	usecaseLayer    = "github.com/sivdead/OmniBotGo/internal/usecase"
	repoLayer       = "github.com/sivdead/OmniBotGo/internal/repo"
	controllerLayer = "github.com/sivdead/OmniBotGo/internal/controller"
	adapterLayer    = "github.com/sivdead/OmniBotGo/internal/adapter"
	serviceLayer    = "github.com/sivdead/OmniBotGo/internal/service"
	configLayer     = "github.com/sivdead/OmniBotGo/internal/config"

	// 定义每个层级允许导入的包
	allowedImports = map[string][]string{
		entityLayer: {
			// entity层不应该依赖任何内部包
		},
		dtoLayer: {
			// dto层只能依赖entity层
			entityLayer,
		},
		usecaseLayer: {
			// usecase层可以依赖entity、dto和自己的port接口
			entityLayer,
			dtoLayer,
			usecaseLayer + "/port",
			usecaseLayer + "/service",
		},
		repoLayer: {
			// repo层可以依赖entity、dto和usecase/port（实现接口）
			entityLayer,
			dtoLayer,
			usecaseLayer + "/port",
		},
		controllerLayer: {
			// controller层可以依赖entity、dto、usecase
			entityLayer,
			dtoLayer,
			usecaseLayer,
		},
		adapterLayer: {
			// adapter层可以依赖entity、dto、usecase/port、config
			entityLayer,
			dtoLayer,
			usecaseLayer + "/port",
			configLayer,
		},
		serviceLayer: {
			// service层可以依赖entity、dto、usecase/port
			entityLayer,
			dtoLayer,
			usecaseLayer + "/port",
			usecaseLayer,
		},
	}
)

// TestArchitectureDependencies 测试架构依赖是否符合整洁架构原则
func TestArchitectureDependencies(t *testing.T) {
	for layer, allowed := range allowedImports {
		t.Run(layer, func(t *testing.T) {
			// 获取包信息
			pkg, err := build.Import(layer, "", build.ImportComment)
			if err != nil {
				t.Skipf("跳过 %s: %v", layer, err)
				return
			}

			// 检查所有导入
			for _, imp := range pkg.Imports {
				// 跳过标准库和第三方包
				if !strings.HasPrefix(imp, "github.com/sivdead/OmniBotGo/internal") {
					continue
				}

				// 检查是否在允许列表中
				if !isAllowedImport(imp, allowed) {
					t.Errorf("%s 不应该导入 %s", layer, imp)
				}
			}
		})
	}
}

// isAllowedImport 检查导入是否被允许
func isAllowedImport(imp string, allowed []string) bool {
	for _, a := range allowed {
		if strings.HasPrefix(imp, a) {
			return true
		}
	}
	return false
}

// TestEntityIndependence 测试entity层的独立性
func TestEntityIndependence(t *testing.T) {
	pkg, err := build.Import(entityLayer, "", build.ImportComment)
	if err != nil {
		t.Skipf("跳过entity层测试: %v", err)
		return
	}

	// entity层不应该导入任何内部包
	for _, imp := range pkg.Imports {
		if strings.HasPrefix(imp, "github.com/sivdead/OmniBotGo/internal") {
			t.Errorf("entity层不应该导入内部包: %s", imp)
		}
	}
}

// TestUseCasePortPattern 测试usecase层的端口模式
func TestUseCasePortPattern(t *testing.T) {
	// 检查repo层是否只依赖usecase/port，而不是usecase本身
	pkg, err := build.Import(repoLayer, "", build.ImportComment)
	if err != nil {
		t.Skipf("跳过repo层测试: %v", err)
		return
	}

	for _, imp := range pkg.Imports {
		if imp == usecaseLayer {
			t.Errorf("repo层不应该直接导入usecase层，应该导入usecase/port")
		}
	}
}
