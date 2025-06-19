// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
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
		// CreateStreamMessageHandler 创建用于Stream适配器的消息处理回调函数
		CreateStreamMessageHandler() port.MessageHandler
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

	// SystemConfigUseCase 系统配置业务逻辑
	SystemConfigUseCase interface {
		// CreateSystemConfig 创建系统配置
		CreateSystemConfig(ctx context.Context, req CreateSystemConfigRequest) (*entity.SystemConfig, error)
		// GetSystemConfig 获取系统配置
		GetSystemConfig(ctx context.Context, key string) (*entity.SystemConfig, error)
		// UpdateSystemConfig 更新系统配置
		UpdateSystemConfig(ctx context.Context, req UpdateSystemConfigRequest) (*entity.SystemConfig, error)
		// DeleteSystemConfig 删除系统配置
		DeleteSystemConfig(ctx context.Context, key string) error
		// ListSystemConfigs 获取系统配置列表
		ListSystemConfigs(ctx context.Context, params ListSystemConfigsParams) (*SystemConfigListResult, error)
		// GetSystemConfigsByGroup 根据组获取系统配置
		GetSystemConfigsByGroup(ctx context.Context, group string) ([]*entity.SystemConfig, error)
	}

	// PlatformUseCase 平台管理业务逻辑
	PlatformUseCase interface {
		// GetPlatforms 获取支持的平台列表
		GetPlatforms(ctx context.Context) ([]*PlatformInfo, error)
		// GetPlatformByType 获取平台详情
		GetPlatformByType(ctx context.Context, platformType string) (*PlatformInfo, error)
		// ValidatePlatformConfig 验证平台配置
		ValidatePlatformConfig(ctx context.Context, req ValidatePlatformConfigRequest) (*PlatformConfigValidationResult, error)
		// GetPlatformStatus 获取平台状态
		GetPlatformStatus(ctx context.Context, platformType string) (*PlatformStatusResult, error)
	}

	// MonitorUseCase 监控业务逻辑
	MonitorUseCase interface {
		// GetSystemOverview 获取系统概览
		GetSystemOverview(ctx context.Context) (*SystemOverviewResult, error)
		// GetSystemMetrics 获取系统指标
		GetSystemMetrics(ctx context.Context, params GetSystemMetricsParams) (*SystemMetricsResult, error)
		// GetDetailedHealth 获取详细健康检查
		GetDetailedHealth(ctx context.Context) (*DetailedHealthResult, error)
	}

	// LogUseCase 日志业务逻辑
	LogUseCase interface {
		// ListConnectionLogs 获取连接日志列表
		ListConnectionLogs(ctx context.Context, params ListConnectionLogsParams) (*ConnectionLogListResult, error)
		// GetConnectionLog 获取连接日志详情
		GetConnectionLog(ctx context.Context, id int64) (*entity.ConnectionLog, error)
		// ListAPICallLogs 获取API调用日志列表
		ListAPICallLogs(ctx context.Context, params ListAPICallLogsParams) (*APICallLogListResult, error)
		// GetAPICallLog 获取API调用日志详情
		GetAPICallLog(ctx context.Context, id int64) (*entity.APICallLog, error)
		// GetAPICallStats 获取API调用统计
		GetAPICallStats(ctx context.Context, params GetAPICallStatsParams) (*APICallStatsResult, error)
	}

	// QueueUseCase 队列业务逻辑
	QueueUseCase interface {
		// ListQueueMessages 获取队列消息列表
		ListQueueMessages(ctx context.Context, params ListQueueMessagesParams) (*QueueMessageListResult, error)
		// GetQueueMessage 获取队列消息详情
		GetQueueMessage(ctx context.Context, id int64) (*entity.MessageQueue, error)
		// RetryQueueMessage 重试队列消息
		RetryQueueMessage(ctx context.Context, id int64) error
		// CancelQueueMessage 取消队列消息
		CancelQueueMessage(ctx context.Context, id int64) error
		// GetQueueStats 获取队列统计
		GetQueueStats(ctx context.Context, params GetQueueStatsParams) (*QueueStatsResult, error)
		// CleanCompletedMessages 清理已完成消息
		CleanCompletedMessages(ctx context.Context, params CleanCompletedMessagesParams) error
	}

	// ProcessorUseCase 处理器业务逻辑
	ProcessorUseCase interface {
		// CreateProcessor 创建处理器
		CreateProcessor(ctx context.Context, req CreateProcessorRequest) (*entity.MessageProcessor, error)
		// ListProcessors 获取处理器列表
		ListProcessors(ctx context.Context, params ListProcessorsParams) (*ProcessorListResult, error)
		// GetProcessor 获取处理器详情
		GetProcessor(ctx context.Context, id int64) (*entity.MessageProcessor, error)
		// UpdateProcessor 更新处理器
		UpdateProcessor(ctx context.Context, req UpdateProcessorRequest) (*entity.MessageProcessor, error)
		// DeleteProcessor 删除处理器
		DeleteProcessor(ctx context.Context, id int64) error
		// UpdateProcessorStatus 更新处理器状态
		UpdateProcessorStatus(ctx context.Context, req UpdateProcessorStatusRequest) error
		// CreateRoutingRule 创建路由规则
		CreateRoutingRule(ctx context.Context, req CreateRoutingRuleRequest) (*entity.MessageRoutingRule, error)
		// ListRoutingRules 获取路由规则列表
		ListRoutingRules(ctx context.Context, processorID int64, params ListRoutingRulesParams) ([]*entity.MessageRoutingRule, error)
		// UpdateRoutingRule 更新路由规则
		UpdateRoutingRule(ctx context.Context, req UpdateRoutingRuleRequest) (*entity.MessageRoutingRule, error)
		// DeleteRoutingRule 删除路由规则
		DeleteRoutingRule(ctx context.Context, ruleID int64) error
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
type ChannelListResult = ListResult[entity.Channel]

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
type MessageHistoryResult = ListResult[entity.Message]

// 系统配置相关

// CreateSystemConfigRequest 创建系统配置请求
type CreateSystemConfigRequest struct {
	Key         string `json:"key" validate:"required"`
	Value       string `json:"value" validate:"required"`
	Type        string `json:"type" validate:"required"`
	Group       string `json:"group"`
	Description string `json:"description"`
	IsEncrypted bool   `json:"is_encrypted"`
	IsSystem    bool   `json:"is_system"`
}

// UpdateSystemConfigRequest 更新系统配置请求
type UpdateSystemConfigRequest struct {
	Key         string  `json:"key" validate:"required"`
	Value       *string `json:"value,omitempty"`
	Type        *string `json:"type,omitempty"`
	Group       *string `json:"group,omitempty"`
	Description *string `json:"description,omitempty"`
	IsEncrypted *bool   `json:"is_encrypted,omitempty"`
}

// ListSystemConfigsParams 获取系统配置列表参数
type ListSystemConfigsParams struct {
	Group    *string `json:"group,omitempty"`
	Key      *string `json:"key,omitempty"`
	IsSystem *bool   `json:"is_system,omitempty"`
	Page     int     `json:"page" validate:"min=1"`
	PageSize int     `json:"page_size" validate:"min=1,max=100"`
}

// SystemConfigListResult 系统配置列表结果
type SystemConfigListResult = ListResult[entity.SystemConfig]

// 平台管理相关

// PlatformInfo 平台信息
type PlatformInfo struct {
	Type              string                `json:"type"`
	Name              string                `json:"name"`
	Description       string                `json:"description"`
	SupportedFeatures []string              `json:"supported_features"`
	ConfigFields      []PlatformConfigField `json:"config_fields"`
	WebhookConfig     PlatformWebhookConfig `json:"webhook_config"`
	Status            string                `json:"status"`
	IconURL           string                `json:"icon_url"`
}

// PlatformConfigField 平台配置字段
type PlatformConfigField struct {
	Field    string `json:"field"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// PlatformWebhookConfig 平台Webhook配置
type PlatformWebhookConfig struct {
	URLPattern  string `json:"url_pattern"`
	Method      string `json:"method"`
	ContentType string `json:"content_type"`
}

// ValidatePlatformConfigRequest 验证平台配置请求
type ValidatePlatformConfigRequest struct {
	Config map[string]interface{} `json:"config" validate:"required"`
}

// PlatformConfigValidationResult 平台配置验证结果
type PlatformConfigValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// PlatformStatusResult 平台状态结果
type PlatformStatusResult struct {
	PlatformType string `json:"platform_type"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

// 监控相关

// SystemOverviewResult 系统概览结果
type SystemOverviewResult struct {
	TotalBots      int64 `json:"total_bots"`
	TotalChannels  int64 `json:"total_channels"`
	ActiveChannels int64 `json:"active_channels"`
	TotalMessages  int64 `json:"total_messages"`
	PendingJobs    int64 `json:"pending_jobs"`
	FailedJobs     int64 `json:"failed_jobs"`
}

// GetSystemMetricsParams 获取系统指标参数
type GetSystemMetricsParams struct {
	MetricType *string `json:"metric_type,omitempty"`
	TimeRange  *string `json:"time_range,omitempty"`
}

// SystemMetricsResult 系统指标结果
type SystemMetricsResult struct {
	CPUUsage    float64            `json:"cpu_usage"`
	MemoryUsage float64            `json:"memory_usage"`
	Metrics     map[string]float64 `json:"metrics"`
}

// DetailedHealthResult 详细健康检查结果
type DetailedHealthResult struct {
	Status     string                     `json:"status"`
	Components map[string]ComponentHealth `json:"components"`
}

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// 日志相关

// ListConnectionLogsParams 获取连接日志列表参数
type ListConnectionLogsParams struct {
	ChannelID *int64  `json:"channel_id,omitempty"`
	LogLevel  *string `json:"log_level,omitempty"`
	Page      int     `json:"page" validate:"min=1"`
	PageSize  int     `json:"page_size" validate:"min=1,max=100"`
}

// ConnectionLogListResult 连接日志列表结果
type ConnectionLogListResult = ListResult[entity.ConnectionLog]

// ListAPICallLogsParams 获取API调用日志列表参数
type ListAPICallLogsParams struct {
	ChannelID   *int64 `json:"channel_id,omitempty"`
	ProcessorID *int64 `json:"processor_id,omitempty"`
	Page        int    `json:"page" validate:"min=1"`
	PageSize    int    `json:"page_size" validate:"min=1,max=100"`
}

// APICallLogListResult API调用日志列表结果
type APICallLogListResult = ListResult[entity.APICallLog]

// GetAPICallStatsParams 获取API调用统计参数
type GetAPICallStatsParams struct {
	ChannelID   *int64 `json:"channel_id,omitempty"`
	ProcessorID *int64 `json:"processor_id,omitempty"`
}

// APICallStatsResult API调用统计结果
type APICallStatsResult struct {
	TotalCalls   int64   `json:"total_calls"`
	SuccessCalls int64   `json:"success_calls"`
	FailedCalls  int64   `json:"failed_calls"`
	SuccessRate  float64 `json:"success_rate"`
	AvgDuration  float64 `json:"avg_duration_ms"`
}

// 队列相关

// ListQueueMessagesParams 获取队列消息列表参数
type ListQueueMessagesParams struct {
	QueueName *string `json:"queue_name,omitempty"`
	Status    *string `json:"status,omitempty"`
	Page      int     `json:"page" validate:"min=1"`
	PageSize  int     `json:"page_size" validate:"min=1,max=100"`
}

// QueueMessageListResult 队列消息列表结果
type QueueMessageListResult = ListResult[entity.MessageQueue]

// GetQueueStatsParams 获取队列统计参数
type GetQueueStatsParams struct {
	QueueName *string `json:"queue_name,omitempty"`
}

// QueueStatsResult 队列统计结果
type QueueStatsResult struct {
	TotalJobs     int64 `json:"total_jobs"`
	PendingJobs   int64 `json:"pending_jobs"`
	RunningJobs   int64 `json:"running_jobs"`
	CompletedJobs int64 `json:"completed_jobs"`
	FailedJobs    int64 `json:"failed_jobs"`
}

// CleanCompletedMessagesParams 清理已完成消息参数
type CleanCompletedMessagesParams struct {
	BeforeDays int `json:"before_days" validate:"min=1"`
}

// 处理器相关

// CreateProcessorRequest 创建处理器请求
type CreateProcessorRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Type        string                 `json:"type" validate:"required"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Priority    int                    `json:"priority"`
	IsEnabled   bool                   `json:"is_enabled"`
}

// ListProcessorsParams 获取处理器列表参数
type ListProcessorsParams struct {
	Type      *string        `json:"type,omitempty"`
	Status    *entity.Status `json:"status,omitempty"`
	IsEnabled *bool          `json:"is_enabled,omitempty"`
	Page      int            `json:"page" validate:"min=1"`
	PageSize  int            `json:"page_size" validate:"min=1,max=100"`
}

// ProcessorListResult 处理器列表结果
type ProcessorListResult = ListResult[entity.MessageProcessor]

// UpdateProcessorRequest 更新处理器请求
type UpdateProcessorRequest struct {
	ID          int64                  `json:"id" validate:"required"`
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Priority    *int                   `json:"priority,omitempty"`
	IsEnabled   *bool                  `json:"is_enabled,omitempty"`
	Status      *entity.Status         `json:"status,omitempty"`
}

// UpdateProcessorStatusRequest 更新处理器状态请求
type UpdateProcessorStatusRequest struct {
	Status entity.Status `json:"status" validate:"required"`
}

// CreateRoutingRuleRequest 创建路由规则请求
type CreateRoutingRuleRequest struct {
	ProcessorID  int64                  `json:"processor_id" validate:"required"`
	RuleName     string                 `json:"rule_name" validate:"required"`
	PlatformType string                 `json:"platform_type"`
	MessageType  string                 `json:"message_type"`
	Conditions   map[string]interface{} `json:"conditions"`
	Priority     int                    `json:"priority"`
	IsFallback   bool                   `json:"is_fallback"`
}

// ListRoutingRulesParams 获取路由规则列表参数
type ListRoutingRulesParams struct {
	PlatformType *string        `json:"platform_type,omitempty"`
	MessageType  *string        `json:"message_type,omitempty"`
	Status       *entity.Status `json:"status,omitempty"`
	IsFallback   *bool          `json:"is_fallback,omitempty"`
	Page         int            `json:"page" validate:"min=1"`
	PageSize     int            `json:"page_size" validate:"min=1,max=100"`
}

// UpdateRoutingRuleRequest 更新路由规则请求
type UpdateRoutingRuleRequest struct {
	ID           int64                  `json:"id" validate:"required"`
	RuleName     *string                `json:"rule_name,omitempty"`
	PlatformType *string                `json:"platform_type,omitempty"`
	MessageType  *string                `json:"message_type,omitempty"`
	Conditions   map[string]interface{} `json:"conditions,omitempty"`
	Priority     *int                   `json:"priority,omitempty"`
	IsFallback   *bool                  `json:"is_fallback,omitempty"`
	Status       *entity.Status         `json:"status,omitempty"`
}
