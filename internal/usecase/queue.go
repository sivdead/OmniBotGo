package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// queueUC 队列用例实现
type queueUC struct {
	queueRepo port.MessageQueueRepository
	logger    logger.Interface
}

// NewQueueUseCase 创建队列用例
func NewQueueUseCase(
	queueRepo port.MessageQueueRepository,
	logger logger.Interface,
) QueueUseCase {
	return &queueUC{
		queueRepo: queueRepo,
		logger:    logger,
	}
}

// ListQueueMessages 获取队列消息列表
func (uc *queueUC) ListQueueMessages(ctx context.Context, params ListQueueMessagesParams) (*QueueMessageListResult, error) {
	// 构建查询参数
	filters := make(map[string]interface{})
	if params.QueueName != nil {
		filters["queue_name"] = *params.QueueName
	}
	if params.Status != nil {
		filters["status"] = *params.Status
	}

	listParams := port.ListParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		Filters:  filters,
		OrderBy:  "created_at DESC",
	}

	// 查询数据
	result, err := uc.queueRepo.List(ctx, listParams)
	if err != nil {
		uc.logger.Error("failed to list queue messages", "error", err)
		return nil, fmt.Errorf("failed to list queue messages: %w", err)
	}

	// 转换结果
	items := make([]entity.MessageQueue, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &QueueMessageListResult{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}, nil
}

// GetQueueMessage 获取队列消息详情
func (uc *queueUC) GetQueueMessage(ctx context.Context, id int64) (*entity.MessageQueue, error) {
	message, err := uc.queueRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("failed to get queue message", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get queue message: %w", err)
	}
	return message, nil
}

// RetryQueueMessage 重试队列消息
func (uc *queueUC) RetryQueueMessage(ctx context.Context, id int64) error {
	// 获取消息
	message, err := uc.queueRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("failed to get queue message for retry", "error", err, "id", id)
		return fmt.Errorf("failed to get queue message: %w", err)
	}

	// 检查状态是否允许重试
	if message.Status != entity.QueueStatusFailed {
		return fmt.Errorf("only failed messages can be retried")
	}

	// 标记为重试状态
	nextScheduleTime := time.Now().Unix()
	if err := uc.queueRepo.MarkForRetry(ctx, id, nextScheduleTime); err != nil {
		uc.logger.Error("failed to mark queue message for retry", "error", err, "id", id)
		return fmt.Errorf("failed to mark queue message for retry: %w", err)
	}

	uc.logger.Info("queue message marked for retry", "id", id)
	return nil
}

// CancelQueueMessage 取消队列消息
func (uc *queueUC) CancelQueueMessage(ctx context.Context, id int64) error {
	// 获取消息
	message, err := uc.queueRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("failed to get queue message for cancellation", "error", err, "id", id)
		return fmt.Errorf("failed to get queue message: %w", err)
	}

	// 检查状态是否允许取消
	if message.Status == entity.QueueStatusCompleted || message.Status == entity.QueueStatusCancelled {
		return fmt.Errorf("cannot cancel completed or already cancelled messages")
	}

	// 标记为取消状态
	if err := uc.queueRepo.MarkAsCancelled(ctx, id); err != nil {
		uc.logger.Error("failed to cancel queue message", "error", err, "id", id)
		return fmt.Errorf("failed to cancel queue message: %w", err)
	}

	uc.logger.Info("queue message cancelled", "id", id)
	return nil
}

// GetQueueStats 获取队列统计
func (uc *queueUC) GetQueueStats(ctx context.Context, params GetQueueStatsParams) (*QueueStatsResult, error) {
	// 构建基础过滤条件
	baseFilters := make(map[string]interface{})
	if params.QueueName != nil {
		baseFilters["queue_name"] = *params.QueueName
	}

	// 获取各状态的统计
	stats := &QueueStatsResult{}

	// 总任务数
	totalParams := port.ListParams{
		Page:     1,
		PageSize: 1,
		Filters:  baseFilters,
	}
	totalResult, err := uc.queueRepo.List(ctx, totalParams)
	if err != nil {
		uc.logger.Error("failed to get total job count", "error", err)
		return nil, fmt.Errorf("failed to get total job count: %w", err)
	}
	stats.TotalJobs = totalResult.Total

	// 待处理任务数
	pendingFilters := make(map[string]interface{})
	for k, v := range baseFilters {
		pendingFilters[k] = v
	}
	pendingFilters["status"] = entity.QueueStatusPending
	pendingParams := port.ListParams{
		Page:     1,
		PageSize: 1,
		Filters:  pendingFilters,
	}
	pendingResult, err := uc.queueRepo.List(ctx, pendingParams)
	if err != nil {
		uc.logger.Error("failed to get pending job count", "error", err)
		return nil, fmt.Errorf("failed to get pending job count: %w", err)
	}
	stats.PendingJobs = pendingResult.Total

	// 运行中任务数
	runningFilters := make(map[string]interface{})
	for k, v := range baseFilters {
		runningFilters[k] = v
	}
	runningFilters["status"] = entity.QueueStatusRunning
	runningParams := port.ListParams{
		Page:     1,
		PageSize: 1,
		Filters:  runningFilters,
	}
	runningResult, err := uc.queueRepo.List(ctx, runningParams)
	if err != nil {
		uc.logger.Error("failed to get running job count", "error", err)
		return nil, fmt.Errorf("failed to get running job count: %w", err)
	}
	stats.RunningJobs = runningResult.Total

	// 已完成任务数
	completedFilters := make(map[string]interface{})
	for k, v := range baseFilters {
		completedFilters[k] = v
	}
	completedFilters["status"] = entity.QueueStatusCompleted
	completedParams := port.ListParams{
		Page:     1,
		PageSize: 1,
		Filters:  completedFilters,
	}
	completedResult, err := uc.queueRepo.List(ctx, completedParams)
	if err != nil {
		uc.logger.Error("failed to get completed job count", "error", err)
		return nil, fmt.Errorf("failed to get completed job count: %w", err)
	}
	stats.CompletedJobs = completedResult.Total

	// 失败任务数
	failedFilters := make(map[string]interface{})
	for k, v := range baseFilters {
		failedFilters[k] = v
	}
	failedFilters["status"] = entity.QueueStatusFailed
	failedParams := port.ListParams{
		Page:     1,
		PageSize: 1,
		Filters:  failedFilters,
	}
	failedResult, err := uc.queueRepo.List(ctx, failedParams)
	if err != nil {
		uc.logger.Error("failed to get failed job count", "error", err)
		return nil, fmt.Errorf("failed to get failed job count: %w", err)
	}
	stats.FailedJobs = failedResult.Total

	return stats, nil
}

// CleanCompletedMessages 清理已完成消息
func (uc *queueUC) CleanCompletedMessages(ctx context.Context, params CleanCompletedMessagesParams) error {
	// 计算截止时间
	cutoffTime := time.Now().AddDate(0, 0, -params.BeforeDays)

	// 获取需要清理的消息
	filters := map[string]interface{}{
		"status":     entity.QueueStatusCompleted,
		"created_at": cutoffTime,
	}
	listParams := port.ListParams{
		Page:     1,
		PageSize: 100, // 批量处理
		Filters:  filters,
		OrderBy:  "created_at ASC",
	}

	// 循环删除
	for {
		result, err := uc.queueRepo.List(ctx, listParams)
		if err != nil {
			uc.logger.Error("failed to list completed messages for cleanup", "error", err)
			return fmt.Errorf("failed to list completed messages: %w", err)
		}

		if len(result.Items) == 0 {
			break
		}

		// 删除消息
		for _, item := range result.Items {
			if err := uc.queueRepo.Delete(ctx, item.ID); err != nil {
				uc.logger.Error("failed to delete completed message", "error", err, "id", item.ID)
				// 继续处理其他消息
				continue
			}
		}

		uc.logger.Info("cleaned completed messages", "count", len(result.Items))

		// 如果已经处理完所有消息，退出循环
		if int64(len(result.Items)) < result.Total {
			listParams.Page++
		} else {
			break
		}
	}

	return nil
}
