// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// MessageUseCase 消息处理业务逻辑
	MessageUseCase interface {
		// ProcessInboundMessage 处理入站消息
		ProcessInboundMessage(ctx context.Context, msg *entity.Message) error
		// SendMessage 发送消息
		SendMessage(ctx context.Context, msg *entity.Message) error
		// GetMessageHistory 获取消息历史
		GetMessageHistory(ctx context.Context, params GetMessageHistoryParams) (*MessageHistoryResult, error)
		// GetMessage 根据ID获取消息
		GetMessage(ctx context.Context, id int64) (*entity.Message, error)
		// RetryFailedMessage 重试失败的消息
		RetryFailedMessage(ctx context.Context, messageID int64) error
	}

	// ChannelUseCase 通道管理业务逻辑
	ChannelUseCase interface {
		// CreateChannel 创建通道
		CreateChannel(ctx context.Context, req CreateChannelRequest) (*entity.Channel, error)
		// UpdateChannel 更新通道
		UpdateChannel(ctx context.Context, req UpdateChannelRequest) (*entity.Channel, error)
		// DeleteChannel 删除通道
		DeleteChannel(ctx context.Context, id int64) error
		// GetChannel 获取通道信息
		GetChannel(ctx context.Context, id int64) (*entity.Channel, error)
		// ListChannels 获取通道列表
		ListChannels(ctx context.Context, params ListChannelsParams) (*ChannelListResult, error)
		// UpdateChannelStatus 更新通道状态
		UpdateChannelStatus(ctx context.Context, id int64, status entity.ConnectionStatus) error
		// RefreshChannelToken 刷新通道令牌
		RefreshChannelToken(ctx context.Context, id int64) error
	}

	// BotUseCase 机器人管理业务逻辑
	BotUseCase interface {
		// CreateBot 创建机器人
		CreateBot(ctx context.Context, req CreateBotRequest) (*entity.Bot, error)
		// UpdateBot 更新机器人
		UpdateBot(ctx context.Context, req UpdateBotRequest) (*entity.Bot, error)
		// DeleteBot 删除机器人
		DeleteBot(ctx context.Context, id int64) error
		// GetBot 获取机器人信息
		GetBot(ctx context.Context, id int64) (*entity.Bot, error)
		// ListBots 获取机器人列表
		ListBots(ctx context.Context, params ListBotsParams) (*BotListResult, error)
	}

	// WebhookUseCase Webhook处理业务逻辑
	WebhookUseCase interface {
		// HandleWebhook 处理平台Webhook事件
		HandleWebhook(ctx context.Context, platform string, channelID int64, payload []byte) error
		// VerifyWebhookSignature 验证Webhook签名
		VerifyWebhookSignature(ctx context.Context, platform string, channelID int64, signature string, payload []byte) error
	}

	// Translation 翻译业务逻辑（保留原有接口）
	Translation interface {
		Translate(context.Context, entity.Translation) (entity.Translation, error)
		History(context.Context) (entity.TranslationHistory, error)
	}
)

// 请求和响应结构体

// CreateChannelRequest 创建通道请求
type CreateChannelRequest struct {
	BotID        int64                  `json:"bot_id" validate:"required"`
	PlatformType string                 `json:"platform_type" validate:"required"`
	ChannelName  string                 `json:"channel_name" validate:"required"`
	Config       map[string]interface{} `json:"config"`
}

// UpdateChannelRequest 更新通道请求
type UpdateChannelRequest struct {
	ID          int64                  `json:"id" validate:"required"`
	ChannelName *string                `json:"channel_name,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Status      *entity.Status         `json:"status,omitempty"`
}

// ListChannelsParams 获取通道列表参数
type ListChannelsParams struct {
	BotID        *int64         `json:"bot_id,omitempty"`
	PlatformType *string        `json:"platform_type,omitempty"`
	Status       *entity.Status `json:"status,omitempty"`
	Page         int            `json:"page" validate:"min=1"`
	PageSize     int            `json:"page_size" validate:"min=1,max=100"`
}

// ChannelListResult 通道列表结果
type ChannelListResult struct {
	Items      []entity.Channel `json:"items"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// CreateBotRequest 创建机器人请求
type CreateBotRequest struct {
	BotName     string                 `json:"bot_name" validate:"required"`
	BotType     string                 `json:"bot_type" validate:"required"`
	Description string                 `json:"description"`
	AvatarURL   string                 `json:"avatar_url"`
	Config      map[string]interface{} `json:"config"`
	CreatedBy   string                 `json:"created_by"`
}

// UpdateBotRequest 更新机器人请求
type UpdateBotRequest struct {
	ID          int64                  `json:"id" validate:"required"`
	BotName     *string                `json:"bot_name,omitempty"`
	Description *string                `json:"description,omitempty"`
	AvatarURL   *string                `json:"avatar_url,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Status      *entity.Status         `json:"status,omitempty"`
}

// ListBotsParams 获取机器人列表参数
type ListBotsParams struct {
	BotType   *string        `json:"bot_type,omitempty"`
	Status    *entity.Status `json:"status,omitempty"`
	CreatedBy *string        `json:"created_by,omitempty"`
	Page      int            `json:"page" validate:"min=1"`
	PageSize  int            `json:"page_size" validate:"min=1,max=100"`
}

// ListResult 通用分页列表结果
type ListResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// BotListResult 机器人列表结果（使用泛型）
type BotListResult = ListResult[entity.Bot]

// GetMessageHistoryParams 获取消息历史参数
type GetMessageHistoryParams struct {
	ChannelID     *int64                   `json:"channel_id,omitempty"`
	SenderID      *string                  `json:"sender_id,omitempty"`
	ReceiverID    *string                  `json:"receiver_id,omitempty"`
	MessageType   *string                  `json:"message_type,omitempty"`
	MessageStatus *entity.MessageStatus    `json:"message_status,omitempty"`
	Direction     *entity.MessageDirection `json:"direction,omitempty"`
	StartTime     *string                  `json:"start_time,omitempty"`
	EndTime       *string                  `json:"end_time,omitempty"`
	Page          int                      `json:"page" validate:"min=1"`
	PageSize      int                      `json:"page_size" validate:"min=1,max=100"`
}

// MessageHistoryResult 消息历史结果
type MessageHistoryResult struct {
	Items      []entity.Message `json:"items"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}
