// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

// 通用查询参数
type ListParams struct {
	Page     int                    `json:"page" validate:"min=1"`
	PageSize int                    `json:"page_size" validate:"min=1,max=100"`
	OrderBy  string                 `json:"order_by"`
	Filters  map[string]interface{} `json:"filters"`
}

type PaginatedResult struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// BotRepo Bot相关Repository接口
type BotRepo interface {
	Create(ctx context.Context, bot *entity.Bot) error
	GetByID(ctx context.Context, id int64) (*entity.Bot, error)
	GetByName(ctx context.Context, name string) (*entity.Bot, error)
	Update(ctx context.Context, bot *entity.Bot) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params ListParams) (*PaginatedResult, error)
	ListActive(ctx context.Context) ([]*entity.Bot, error)
	Exists(ctx context.Context, id int64) (bool, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	GetActiveChannelCount(ctx context.Context, id int64) (int64, error)
}

// ChannelRepo Channel相关Repository接口
type ChannelRepo interface {
	Create(ctx context.Context, channel *entity.Channel) error
	GetByID(ctx context.Context, id int64) (*entity.Channel, error)
	GetByBotID(ctx context.Context, botID int64) ([]*entity.Channel, error)
	GetByPlatformType(ctx context.Context, platformType string) ([]*entity.Channel, error)
	GetByWebhookPath(ctx context.Context, path string) (*entity.Channel, error)
	Update(ctx context.Context, channel *entity.Channel) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params ListParams) (*PaginatedResult, error)
	ListActive(ctx context.Context) ([]*entity.Channel, error)
	UpdateConnectionStatus(ctx context.Context, id int64, status entity.ConnectionStatus) error
	UpdateAccessToken(ctx context.Context, id int64, token string, expiresAt *time.Time) error
	Exists(ctx context.Context, id int64) (bool, error)
}

// MessageRepo Message相关Repository接口
type MessageRepo interface {
	Create(ctx context.Context, message *entity.Message) error
	GetByID(ctx context.Context, id int64) (*entity.Message, error)
	GetByMessageID(ctx context.Context, messageID string) (*entity.Message, error)
	GetByChannelID(ctx context.Context, channelID int64, params ListParams) (*PaginatedResult, error)
	GetByConversationID(ctx context.Context, conversationID string, params ListParams) (*PaginatedResult, error)
	GetPendingMessages(ctx context.Context, limit int) ([]*entity.Message, error)
	GetFailedMessages(ctx context.Context, limit int) ([]*entity.Message, error)
	Update(ctx context.Context, message *entity.Message) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params ListParams) (*PaginatedResult, error)
	UpdateStatus(ctx context.Context, id int64, status entity.MessageStatus) error
	IncrementRetryCount(ctx context.Context, id int64) error
	MarkAsProcessed(ctx context.Context, id int64) error
	MarkAsSent(ctx context.Context, id int64) error
	MarkAsFailed(ctx context.Context, id int64, errorMsg string) error
	Exists(ctx context.Context, id int64) (bool, error)
	ExistsByMessageID(ctx context.Context, messageID string) (bool, error)
}

// MessageProcessorRepo MessageProcessor相关Repository接口
type MessageProcessorRepo interface {
	Create(ctx context.Context, processor *entity.MessageProcessor) error
	GetByID(ctx context.Context, id int64) (*entity.MessageProcessor, error)
	GetByName(ctx context.Context, name string) (*entity.MessageProcessor, error)
	GetByType(ctx context.Context, processorType string) ([]*entity.MessageProcessor, error)
	Update(ctx context.Context, processor *entity.MessageProcessor) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params ListParams) (*PaginatedResult, error)
	ListActive(ctx context.Context) ([]*entity.MessageProcessor, error)
	ListByPriority(ctx context.Context) ([]*entity.MessageProcessor, error)
	Exists(ctx context.Context, id int64) (bool, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
}

// MessageRoutingRuleRepo MessageRoutingRule相关Repository接口
type MessageRoutingRuleRepo interface {
	Create(ctx context.Context, rule *entity.MessageRoutingRule) error
	GetByID(ctx context.Context, id int64) (*entity.MessageRoutingRule, error)
	GetByProcessorID(ctx context.Context, processorID int64) ([]*entity.MessageRoutingRule, error)
	GetActiveRules(ctx context.Context) ([]*entity.MessageRoutingRule, error)
	GetMatchingRules(ctx context.Context, botID int64, platformType string, channelID int64, messageType string) ([]*entity.MessageRoutingRule, error)
	GetFallbackRules(ctx context.Context) ([]*entity.MessageRoutingRule, error)
	Update(ctx context.Context, rule *entity.MessageRoutingRule) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params ListParams) (*PaginatedResult, error)
	ListByPriority(ctx context.Context) ([]*entity.MessageRoutingRule, error)
	Exists(ctx context.Context, id int64) (bool, error)
}

// SystemConfigRepo SystemConfig相关Repository接口
type SystemConfigRepo interface {
	Create(ctx context.Context, config *entity.SystemConfig) error
	GetByKey(ctx context.Context, key string) (*entity.SystemConfig, error)
	GetByGroup(ctx context.Context, group string) ([]*entity.SystemConfig, error)
	GetAllByGroup(ctx context.Context, groups []string) ([]*entity.SystemConfig, error)
	GetUserEditableConfigs(ctx context.Context) ([]*entity.SystemConfig, error)
	GetSystemConfigs(ctx context.Context) ([]*entity.SystemConfig, error)
	Update(ctx context.Context, config *entity.SystemConfig) error
	UpdateValue(ctx context.Context, key, value string) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params ListParams) (*PaginatedResult, error)
	Exists(ctx context.Context, key string) (bool, error)
}

// MessageQueueRepo MessageQueue相关Repository接口
type MessageQueueRepo interface {
	Create(ctx context.Context, queue *entity.MessageQueue) error
	GetByID(ctx context.Context, id int64) (*entity.MessageQueue, error)
	GetPendingJobs(ctx context.Context, queueName string, limit int) ([]*entity.MessageQueue, error)
	GetRetryableJobs(ctx context.Context, limit int) ([]*entity.MessageQueue, error)
	GetExpiredJobs(ctx context.Context, timeout int64) ([]*entity.MessageQueue, error)
	Update(ctx context.Context, queue *entity.MessageQueue) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params ListParams) (*PaginatedResult, error)
	MarkAsRunning(ctx context.Context, id int64) error
	MarkAsCompleted(ctx context.Context, id int64) error
	MarkAsFailed(ctx context.Context, id int64, errorMsg string) error
	MarkForRetry(ctx context.Context, id int64, nextScheduleTime int64) error
	MarkAsCancelled(ctx context.Context, id int64) error
	Exists(ctx context.Context, id int64) (bool, error)
}

// ConnectionLogRepo ConnectionLog相关Repository接口
type ConnectionLogRepo interface {
	Create(ctx context.Context, log *entity.ConnectionLog) error
	GetByChannelID(ctx context.Context, channelID int64, params ListParams) (*PaginatedResult, error)
	GetRecentLogs(ctx context.Context, limit int) ([]*entity.ConnectionLog, error)
	GetErrorLogs(ctx context.Context, channelID int64, limit int) ([]*entity.ConnectionLog, error)
	List(ctx context.Context, params ListParams) (*PaginatedResult, error)
	Delete(ctx context.Context, id int64) error
	DeleteOldLogs(ctx context.Context, beforeDays int) error
}

// APICallLogRepo APICallLog相关Repository接口
type APICallLogRepo interface {
	Create(ctx context.Context, log *entity.APICallLog) error
	GetByRequestID(ctx context.Context, requestID string) (*entity.APICallLog, error)
	GetByChannelID(ctx context.Context, channelID int64, params ListParams) (*PaginatedResult, error)
	GetByProcessorID(ctx context.Context, processorID int64, params ListParams) (*PaginatedResult, error)
	GetSlowCalls(ctx context.Context, thresholdMs int, limit int) ([]*entity.APICallLog, error)
	GetFailedCalls(ctx context.Context, limit int) ([]*entity.APICallLog, error)
	GetRecentCalls(ctx context.Context, limit int) ([]*entity.APICallLog, error)
	List(ctx context.Context, params ListParams) (*PaginatedResult, error)
	Delete(ctx context.Context, id int64) error
	DeleteOldLogs(ctx context.Context, beforeDays int) error
	GetStatistics(ctx context.Context, channelID *int64, processorID *int64) (*CallStatistics, error)
}

// CallStatistics API调用统计信息
type CallStatistics struct {
	TotalCalls   int64   `json:"total_calls"`
	SuccessCalls int64   `json:"success_calls"`
	FailedCalls  int64   `json:"failed_calls"`
	SuccessRate  float64 `json:"success_rate"`
	AvgDuration  float64 `json:"avg_duration_ms"`
	MaxDuration  int     `json:"max_duration_ms"`
	MinDuration  int     `json:"min_duration_ms"`
}
