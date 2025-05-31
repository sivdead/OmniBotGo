package v1

import (
	"github.com/go-playground/validator/v10"
	v1 "github.com/sivdead/OmniBotGo/docs/proto/v1"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// V1 -.
type V1 struct {
	v1.TranslationServer

	t usecase.Translation
	l logger.Interface
	v *validator.Validate
}
