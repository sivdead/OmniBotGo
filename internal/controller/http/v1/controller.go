package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// 类型别名，用于简化引用
type (
	MessageUseCase = usecase.MessageUseCase
	ChannelUseCase = usecase.ChannelUseCase
	BotUseCase     = usecase.BotUseCase
	WebhookUseCase = usecase.WebhookUseCase
)

// V1 HTTP控制器结构体
type V1 struct {
	// 现有接口
	t usecase.Translation
	l logger.Interface
	v *validator.Validate

	// 新增业务逻辑接口
	messageUC MessageUseCase
	channelUC ChannelUseCase
	botUC     BotUseCase
	webhookUC WebhookUseCase
}

// NewV1Controller 创建V1控制器实例
func NewV1Controller(
	t usecase.Translation,
	messageUC MessageUseCase,
	channelUC ChannelUseCase,
	botUC BotUseCase,
	webhookUC WebhookUseCase,
	l logger.Interface,
) *V1 {
	return &V1{
		t:         t,
		messageUC: messageUC,
		channelUC: channelUC,
		botUC:     botUC,
		webhookUC: webhookUC,
		l:         l,
		v:         validator.New(),
	}
}
