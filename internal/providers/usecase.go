package providers

import (
	"github.com/google/wire"
	"github.com/sivdead/OmniBotGo/internal/repo"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/internal/usecase/translation"
)

// UseCaseSet 包含所有usecase相关的Provider
var UseCaseSet = wire.NewSet(
	NewTranslationUseCase,
	wire.Bind(new(usecase.Translation), new(*translation.UseCase)),
)

// NewTranslationUseCase 创建TranslationUseCase实例
func NewTranslationUseCase(
	translationRepo repo.TranslationRepo,
	translationWebAPI repo.TranslationWebAPI,
) *translation.UseCase {
	return translation.New(translationRepo, translationWebAPI)
}
