package persistent

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/database"
)

// MessageQueueRepo MessageQueue相关的数据访问层实现
type MessageQueueRepo struct {
	*BaseRepo
}

// NewMessageQueueRepository 创建MessageQueue Repository实例
func NewMessageQueueRepository(db database.CommonDB) port.MessageQueueRepository {
	return &MessageQueueRepo{
		BaseRepo: NewBaseRepo(db),
	}
}

// Create 创建新的MessageQueue
func (r *MessageQueueRepo) Create(ctx context.Context, queue *entity.MessageQueue) error {
	if err := queue.Validate(); err != nil {
		return fmt.Errorf("queue validation failed: %w", err)
	}

	if err := r.db.GetGORM().WithContext(ctx).Create(queue).Error; err != nil {
		return r.handleError(err, "create message queue")
	}
	return nil
}

// GetByID 根据ID获取MessageQueue
func (r *MessageQueueRepo) GetByID(ctx context.Context, id int64) (*entity.MessageQueue, error) {
	var queue entity.MessageQueue
	err := r.db.GetGORM().WithContext(ctx).First(&queue, id).Error
	if err != nil {
		return nil, r.handleError(err, "get message queue by id")
	}
	return &queue, nil
}

// GetPendingJobs 获取待处理的任务
func (r *MessageQueueRepo) GetPendingJobs(ctx context.Context, queueName string, limit int) ([]*entity.MessageQueue, error) {
	var queues []*entity.MessageQueue
	query := r.db.GetGORM().WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", entity.QueueStatusPending, time.Now()).
		Order("priority ASC, scheduled_at ASC")

	if queueName != "" {
		query = query.Where("queue_name = ?", queueName)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&queues).Error
	if err != nil {
		return nil, r.handleError(err, "get pending jobs")
	}
	return queues, nil
}

// GetRetryableJobs 获取可重试的任务
func (r *MessageQueueRepo) GetRetryableJobs(ctx context.Context, limit int) ([]*entity.MessageQueue, error) {
	var queues []*entity.MessageQueue
	query := r.db.GetGORM().WithContext(ctx).
		Where("status = ? AND retry_count < max_retries AND scheduled_at <= ?",
			entity.QueueStatusRetrying, time.Now()).
		Order("priority ASC, scheduled_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&queues).Error
	if err != nil {
		return nil, r.handleError(err, "get retryable jobs")
	}
	return queues, nil
}

// GetExpiredJobs 获取超时的任务
func (r *MessageQueueRepo) GetExpiredJobs(ctx context.Context, timeout int64) ([]*entity.MessageQueue, error) {
	var queues []*entity.MessageQueue
	expiredTime := time.Now().Add(-time.Duration(timeout) * time.Second)

	err := r.db.GetGORM().WithContext(ctx).
		Where("status = ? AND started_at < ?", entity.QueueStatusRunning, expiredTime).
		Find(&queues).Error
	if err != nil {
		return nil, r.handleError(err, "get expired jobs")
	}
	return queues, nil
}

// Update 更新MessageQueue
func (r *MessageQueueRepo) Update(ctx context.Context, queue *entity.MessageQueue) error {
	if err := queue.Validate(); err != nil {
		return fmt.Errorf("queue validation failed: %w", err)
	}

	result := r.db.GetGORM().WithContext(ctx).Save(queue)
	if result.Error != nil {
		return r.handleError(result.Error, "update message queue")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除MessageQueue（软删除）
func (r *MessageQueueRepo) Delete(ctx context.Context, id int64) error {
	return r.softDelete(ctx, &entity.MessageQueue{}, id)
}

// List 获取MessageQueue列表（分页）
func (r *MessageQueueRepo) List(ctx context.Context, params port.ListParams) (*port.PaginatedResult[*entity.MessageQueue], error) {
	internalParams := convertToInternalParams(params)
	internalParams = r.validateParams(internalParams)

	var queues []*entity.MessageQueue
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.MessageQueue{}), internalParams)

	result, err := PaginateTypedForPort(r.db.GetGORM(), ctx, query, internalParams, &queues)
	if err != nil {
		return nil, r.handleError(err, "list message queues")
	}

	return result, nil
}

// MarkAsRunning 标记任务为运行中
func (r *MessageQueueRepo) MarkAsRunning(ctx context.Context, id int64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     entity.QueueStatusRunning,
		"started_at": &now,
	}

	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.MessageQueue{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return r.handleError(result.Error, "mark job as running")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MarkAsCompleted 标记任务为已完成
func (r *MessageQueueRepo) MarkAsCompleted(ctx context.Context, id int64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       entity.QueueStatusCompleted,
		"completed_at": &now,
	}

	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.MessageQueue{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return r.handleError(result.Error, "mark job as completed")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MarkAsFailed 标记任务为失败
func (r *MessageQueueRepo) MarkAsFailed(ctx context.Context, id int64, errorMsg string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":        entity.QueueStatusFailed,
		"error_message": errorMsg,
		"retry_count":   gorm.Expr("retry_count + 1"),
		"completed_at":  &now,
	}

	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.MessageQueue{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return r.handleError(result.Error, "mark job as failed")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MarkForRetry 标记任务为重试状态
func (r *MessageQueueRepo) MarkForRetry(ctx context.Context, id int64, nextScheduleTime int64) error {
	nextTime := time.Unix(nextScheduleTime, 0)
	updates := map[string]interface{}{
		"status":       entity.QueueStatusRetrying,
		"scheduled_at": nextTime,
	}

	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.MessageQueue{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return r.handleError(result.Error, "mark job for retry")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MarkAsCancelled 标记任务为已取消
func (r *MessageQueueRepo) MarkAsCancelled(ctx context.Context, id int64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       entity.QueueStatusCancelled,
		"completed_at": &now,
	}

	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.MessageQueue{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return r.handleError(result.Error, "mark job as cancelled")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Exists 检查MessageQueue是否存在
func (r *MessageQueueRepo) Exists(ctx context.Context, id int64) (bool, error) {
	return r.exists(ctx, &entity.MessageQueue{}, "id = ?", id)
}

// 确保 MessageQueueRepo 实现了 port.MessageQueueRepository 接口
var _ port.MessageQueueRepository = (*MessageQueueRepo)(nil)
