package entity

import (
	"time"
)

// MessageQueue 消息队列实体，用于异步处理消息
type MessageQueue struct {
	BaseEntity
	QueueName    string      `json:"queue_name" gorm:"column:queue_name;type:varchar(100);not null;index;comment:队列名称"`
	MessageID    string      `json:"message_id" gorm:"column:message_id;type:varchar(100);not null;index;comment:消息ID"`
	Payload      JSONField   `json:"payload" gorm:"column:payload;type:json;comment:消息负载"`
	Priority     int         `json:"priority" gorm:"column:priority;default:100;comment:优先级"`
	MaxRetries   int         `json:"max_retries" gorm:"column:max_retries;default:3;comment:最大重试次数"`
	RetryCount   int         `json:"retry_count" gorm:"column:retry_count;default:0;comment:当前重试次数"`
	Status       QueueStatus `json:"status" gorm:"column:status;type:tinyint;default:0;comment:队列状态:0-待处理,1-运行中,2-已完成,3-失败,4-重试中,5-已取消"`
	ScheduledAt  time.Time   `json:"scheduled_at" gorm:"column:scheduled_at;comment:调度时间"`
	StartedAt    *time.Time  `json:"started_at" gorm:"column:started_at;comment:开始处理时间"`
	CompletedAt  *time.Time  `json:"completed_at" gorm:"column:completed_at;comment:完成处理时间"`
	ErrorMessage string      `json:"error_message" gorm:"column:error_message;type:text;comment:错误消息"`
}

// TableName 指定表名
func (MessageQueue) TableName() string {
	return "message_queue"
}

// IsPending 检查是否为待处理状态
func (mq *MessageQueue) IsPending() bool {
	return mq.Status == QueueStatusPending
}

// IsRunning 检查是否正在运行
func (mq *MessageQueue) IsRunning() bool {
	return mq.Status == QueueStatusRunning
}

// IsCompleted 检查是否已完成
func (mq *MessageQueue) IsCompleted() bool {
	return mq.Status == QueueStatusCompleted
}

// IsFailed 检查是否已失败
func (mq *MessageQueue) IsFailed() bool {
	return mq.Status == QueueStatusFailed
}

// CanRetry 检查是否可以重试
func (mq *MessageQueue) CanRetry() bool {
	return mq.Status == QueueStatusFailed && mq.RetryCount < mq.MaxRetries
}

// IsExpired 检查是否已过期（超过调度时间且未完成）
func (mq *MessageQueue) IsExpired(timeout time.Duration) bool {
	if mq.IsCompleted() {
		return false
	}
	return time.Now().After(mq.ScheduledAt.Add(timeout))
}

// MarkAsRunning 标记为运行中
func (mq *MessageQueue) MarkAsRunning() {
	mq.Status = QueueStatusRunning
	now := time.Now()
	mq.StartedAt = &now
}

// MarkAsCompleted 标记为已完成
func (mq *MessageQueue) MarkAsCompleted() {
	mq.Status = QueueStatusCompleted
	now := time.Now()
	mq.CompletedAt = &now
}

// MarkAsFailed 标记为失败并记录错误
func (mq *MessageQueue) MarkAsFailed(errorMsg string) {
	mq.Status = QueueStatusFailed
	mq.ErrorMessage = errorMsg
	mq.RetryCount++
	now := time.Now()
	mq.CompletedAt = &now
}

// MarkForRetry 标记为重试状态
func (mq *MessageQueue) MarkForRetry(delay time.Duration) {
	if mq.CanRetry() {
		mq.Status = QueueStatusRetrying
		mq.ScheduledAt = time.Now().Add(delay)
	}
}

// MarkAsCancelled 标记为已取消
func (mq *MessageQueue) MarkAsCancelled() {
	mq.Status = QueueStatusCancelled
	now := time.Now()
	mq.CompletedAt = &now
}

// GetPayloadValue 获取负载中的特定值
func (mq *MessageQueue) GetPayloadValue(key string) interface{} {
	if mq.Payload == nil {
		return nil
	}
	return mq.Payload.Get(key)
}

// SetPayloadValue 设置负载中的特定值
func (mq *MessageQueue) SetPayloadValue(key string, value interface{}) {
	if mq.Payload == nil {
		mq.Payload = make(JSONField)
	}
	mq.Payload.Set(key, value)
}

// GetDuration 获取处理耗时
func (mq *MessageQueue) GetDuration() time.Duration {
	if mq.StartedAt == nil || mq.CompletedAt == nil {
		return 0
	}
	return mq.CompletedAt.Sub(*mq.StartedAt)
}

// Validate 验证MessageQueue实体数据
func (mq *MessageQueue) Validate() error {
	if mq.QueueName == "" {
		return NewValidationError("queue_name", "队列名称不能为空")
	}
	if mq.MessageID == "" {
		return NewValidationError("message_id", "消息ID不能为空")
	}
	if mq.MaxRetries < 0 {
		return NewValidationError("max_retries", "最大重试次数不能小于0")
	}
	return nil
}
