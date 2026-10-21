package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/internal/usecase/service"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// messageUseCase 消息处理业务逻辑实现
type messageUseCase struct {
	messageRepo    port.MessageRepository
	channelRepo    port.ChannelRepository
	adapterManager port.AdapterManager
	routingUC      RoutingUseCase
	msgService     *service.MessageService
	queueRepo      port.MessageQueueRepository
	logger         logger.Interface
}

// NewMessageUseCase 创建消息处理业务逻辑实例
func NewMessageUseCase(
	messageRepo port.MessageRepository,
	channelRepo port.ChannelRepository,
	adapterManager port.AdapterManager,
	routingUC RoutingUseCase,
	queueRepo port.MessageQueueRepository,
	logger logger.Interface,
) MessageUseCase {
	return &messageUseCase{
		messageRepo:    messageRepo,
		channelRepo:    channelRepo,
		adapterManager: adapterManager,
		routingUC:      routingUC,
		msgService:     service.NewMessageService(),
		queueRepo:      queueRepo,
		logger:         logger,
	}
}

const (
	successfulProcessorsKey  = "_successful_processor_ids"
	configKeyWebhookURL      = "webhook_url"
	payloadKeyProcessorID    = "processor_id"
	payloadKeyMessageID      = "message_id"
	payloadKeyChannelID      = "channel_id"
	configKeyStreamMode      = "stream_mode"
	orderByCreatedAt         = "created_at"
	httpMethodPOST           = "POST"
	actionHealthy            = "healthy"
	jsonKeyID                = "id"
	jsonKeyFunction          = "function"
	jsonKeyArgs              = "args"
	defaultAIHistoryLimit    = 10
	webhookHTTPTimeoutSec    = 30
	httpStatusTooManyReqs    = 429
	httpStatusRequestTimeout = 408
	webhookQueuePriority     = 50
	webhookQueueMaxRetries   = 5
)

var (
	errChannelNotReady         = errors.New("通道未就绪")
	errChannelInactive         = errors.New("通道未激活")
	errWebhookDeliveryFailed   = errors.New("webhook delivery failed")
	errRetryExhausted          = errors.New("retry exhausted")
	errUnsupportedMsgDirection = errors.New("unsupported message direction")
	errMessageNotFailed        = errors.New("消息状态不是失败状态")
	errAIProviderRequired      = errors.New("AI provider is required")
	errAIModelRequired         = errors.New("AI model is required")
	errAIAPIKeyRequired        = errors.New("API key is required")
	errAIModelNotToolCalling   = errors.New("created model does not implement ToolCallingChatModel interface")
	errAIEmptyResponse         = errors.New("empty response from AI model")
	errAIEmptyStreamResponse   = errors.New("empty response from AI model stream")
	errUnknownToolFunction     = errors.New("unknown tool function")
	errUnknownProcessorType    = errors.New("未知的处理器类型")
	errWebhookURLMissing       = errors.New("Webhook处理器缺少webhook_url配置")
	errWebhookURLInvalid       = errors.New("无效的webhook_url配置")
	errAutoReplyContentMissing = errors.New("自动回复处理器缺少reply_content配置")
	errAutoReplyContentInvalid = errors.New("无效的reply_content配置")
	// ErrReplyRequiresInbound is returned when ReplyToOriginal is given a non-inbound message.
	ErrReplyRequiresInbound = errors.New("reply requires an inbound message")
	// ErrInboundMessageNotFound is returned when the inbound ref cannot be resolved.
	ErrInboundMessageNotFound = errors.New("inbound message not found")
)

// generateMessageID 生成消息ID
func (uc *messageUseCase) generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// enqueueFailedMessage 将失败的消息加入重试队列
func (uc *messageUseCase) enqueueFailedMessage(ctx context.Context, msg *entity.Message) error {
	queue := &entity.MessageQueue{
		QueueName:   "message_retry",
		MessageID:   msg.MessageID,
		Priority:    100,
		MaxRetries:  3,
		Status:      entity.QueueStatusPending,
		ScheduledAt: time.Now().Add(5 * time.Minute), // 5分钟后重试
		Payload: entity.JSONField{
			"message_id":   msg.ID,
			"channel_id":   msg.ChannelID,
			"retry_reason": "processing_failed",
		},
	}

	return uc.queueRepo.Create(ctx, queue)
}
