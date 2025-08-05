package usecase

import (
	"context"
	"fmt"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// logUC 日志用例实现
type logUC struct {
	connectionLogRepo port.ConnectionLogRepository
	apiCallLogRepo    port.APICallLogRepository
	logger            logger.Interface
}

// NewLogUseCase 创建日志用例
func NewLogUseCase(
	connectionLogRepo port.ConnectionLogRepository,
	apiCallLogRepo port.APICallLogRepository,
	logger logger.Interface,
) LogUseCase {
	return &logUC{
		connectionLogRepo: connectionLogRepo,
		apiCallLogRepo:    apiCallLogRepo,
		logger:            logger,
	}
}

// ListConnectionLogs 获取连接日志列表
func (uc *logUC) ListConnectionLogs(ctx context.Context, params ListConnectionLogsParams) (*ConnectionLogListResult, error) {
	// 构建查询参数
	filters := make(map[string]interface{})
	if params.ChannelID != nil {
		filters["channel_id"] = *params.ChannelID
	}
	if params.LogLevel != nil {
		filters["log_level"] = *params.LogLevel
	}

	listParams := port.ListParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		Filters:  filters,
		OrderBy:  "created_at DESC",
	}

	// 查询数据
	result, err := uc.connectionLogRepo.List(ctx, listParams)
	if err != nil {
		uc.logger.Error("failed to list connection logs", "error", err)
		return nil, fmt.Errorf("failed to list connection logs: %w", err)
	}

	// 转换结果
	items := make([]entity.ConnectionLog, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &ConnectionLogListResult{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}, nil
}

// GetConnectionLog 获取连接日志详情
func (uc *logUC) GetConnectionLog(ctx context.Context, id int64) (*entity.ConnectionLog, error) {
	// 由于ConnectionLogRepository没有GetByID方法，我们使用List方法来查询
	filters := map[string]interface{}{
		"id": id,
	}
	listParams := port.ListParams{
		Page:     1,
		PageSize: 1,
		Filters:  filters,
	}

	result, err := uc.connectionLogRepo.List(ctx, listParams)
	if err != nil {
		uc.logger.Error("failed to get connection log", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get connection log: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("connection log not found")
	}

	return result.Items[0], nil
}

// ListAPICallLogs 获取API调用日志列表
func (uc *logUC) ListAPICallLogs(ctx context.Context, params ListAPICallLogsParams) (*APICallLogListResult, error) {
	// 构建查询参数
	filters := make(map[string]interface{})
	if params.ChannelID != nil {
		filters["channel_id"] = *params.ChannelID
	}
	if params.ProcessorID != nil {
		filters["processor_id"] = *params.ProcessorID
	}

	listParams := port.ListParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		Filters:  filters,
		OrderBy:  "created_at DESC",
	}

	// 查询数据
	result, err := uc.apiCallLogRepo.List(ctx, listParams)
	if err != nil {
		uc.logger.Error("failed to list API call logs", "error", err)
		return nil, fmt.Errorf("failed to list API call logs: %w", err)
	}

	// 转换结果
	items := make([]entity.APICallLog, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &APICallLogListResult{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}, nil
}

// GetAPICallLog 获取API调用日志详情
func (uc *logUC) GetAPICallLog(ctx context.Context, id int64) (*entity.APICallLog, error) {
	// 由于APICallLogRepository没有GetByID方法，我们使用List方法来查询
	filters := map[string]interface{}{
		"id": id,
	}
	listParams := port.ListParams{
		Page:     1,
		PageSize: 1,
		Filters:  filters,
	}

	result, err := uc.apiCallLogRepo.List(ctx, listParams)
	if err != nil {
		uc.logger.Error("failed to get API call log", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get API call log: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("API call log not found")
	}

	return result.Items[0], nil
}

// GetAPICallStats 获取API调用统计
func (uc *logUC) GetAPICallStats(ctx context.Context, params GetAPICallStatsParams) (*APICallStatsResult, error) {
	stats, err := uc.apiCallLogRepo.GetStatistics(ctx, params.ChannelID, params.ProcessorID)
	if err != nil {
		uc.logger.Error("failed to get API call statistics", "error", err)
		return nil, fmt.Errorf("failed to get API call statistics: %w", err)
	}

	// 计算成功率
	successRate := float64(0)
	if stats.TotalCalls > 0 {
		successRate = float64(stats.SuccessCalls) / float64(stats.TotalCalls) * 100
	}

	return &APICallStatsResult{
		TotalCalls:   stats.TotalCalls,
		SuccessCalls: stats.SuccessCalls,
		FailedCalls:  stats.FailedCalls,
		SuccessRate:  successRate,
		AvgDuration:  stats.AvgDuration,
	}, nil
}
