package persistent

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/database"
)

// ConnectionLogRepo ConnectionLog相关的数据访问层实现
type ConnectionLogRepo struct {
	*BaseRepo
}

// NewConnectionLogRepository 创建ConnectionLog Repository实例
func NewConnectionLogRepository(db database.CommonDB) port.ConnectionLogRepository {
	return &ConnectionLogRepo{
		BaseRepo: NewBaseRepo(db),
	}
}

// Create 创建新的ConnectionLog
func (r *ConnectionLogRepo) Create(ctx context.Context, log *entity.ConnectionLog) error {
	if err := log.Validate(); err != nil {
		return fmt.Errorf("connection log validation failed: %w", err)
	}

	if err := r.db.GetGORM().WithContext(ctx).Create(log).Error; err != nil {
		return r.handleError(err, "create connection log")
	}
	return nil
}

// GetByChannelID 根据通道ID获取ConnectionLog列表（分页）
func (r *ConnectionLogRepo) GetByChannelID(ctx context.Context, channelID string, params port.ListParams) (*port.PaginatedResult[*entity.ConnectionLog], error) {
	internalParams := convertToInternalParams(params)
	internalParams = r.validateParams(internalParams)

	var logs []*entity.ConnectionLog
	query := r.db.GetGORM().WithContext(ctx).
		Model(&entity.ConnectionLog{}).
		Where("channel_id = ?", channelID)

	query = r.buildQuery(query, internalParams)

	result, err := PaginateTypedForPort(r.db.GetGORM(), ctx, query, internalParams, &logs)
	if err != nil {
		return nil, r.handleError(err, "get connection logs by channel id")
	}

	return result, nil
}

// GetRecentLogs 获取最近的连接日志
func (r *ConnectionLogRepo) GetRecentLogs(ctx context.Context, limit int) ([]*entity.ConnectionLog, error) {
	var logs []*entity.ConnectionLog
	query := r.db.GetGORM().WithContext(ctx).
		Preload("Channel").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	if err != nil {
		return nil, r.handleError(err, "get recent connection logs")
	}
	return logs, nil
}

// GetErrorLogs 获取错误连接日志
func (r *ConnectionLogRepo) GetErrorLogs(ctx context.Context, channelID string, limit int) ([]*entity.ConnectionLog, error) {
	var logs []*entity.ConnectionLog
	query := r.db.GetGORM().WithContext(ctx).
		Where("status = ?", entity.StatusInactive).
		Order("created_at DESC")

	if channelID != "" {
		query = query.Where("channel_id = ?", channelID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	if err != nil {
		return nil, r.handleError(err, "get error connection logs")
	}
	return logs, nil
}

// List 获取ConnectionLog列表（分页）
func (r *ConnectionLogRepo) List(ctx context.Context, params port.ListParams) (*port.PaginatedResult[*entity.ConnectionLog], error) {
	internalParams := convertToInternalParams(params)
	internalParams = r.validateParams(internalParams)

	var logs []*entity.ConnectionLog
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.ConnectionLog{}), internalParams)

	result, err := PaginateTypedForPort(r.db.GetGORM(), ctx, query, internalParams, &logs)
	if err != nil {
		return nil, r.handleError(err, "list connection logs")
	}

	return result, nil
}

// Delete 删除ConnectionLog（软删除）
func (r *ConnectionLogRepo) Delete(ctx context.Context, id string) error {
	return r.softDelete(ctx, &entity.ConnectionLog{}, id)
}

// DeleteOldLogs 删除指定天数前的日志
func (r *ConnectionLogRepo) DeleteOldLogs(ctx context.Context, beforeDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -beforeDays)
	return r.hardDelete(ctx, &entity.ConnectionLog{}, "created_at < ?", cutoffTime)
}

// 确保 ConnectionLogRepo 实现了 port.ConnectionLogRepository 接口
var _ port.ConnectionLogRepository = (*ConnectionLogRepo)(nil)

// APICallLogRepo APICallLog相关的数据访问层实现
type APICallLogRepo struct {
	*BaseRepo
}

// NewAPICallLogRepository 创建APICallLog Repository实例
func NewAPICallLogRepository(db database.CommonDB) port.APICallLogRepository {
	return &APICallLogRepo{
		BaseRepo: NewBaseRepo(db),
	}
}

// Create 创建新的APICallLog
func (r *APICallLogRepo) Create(ctx context.Context, log *entity.APICallLog) error {
	if err := log.Validate(); err != nil {
		return fmt.Errorf("api call log validation failed: %w", err)
	}

	if err := r.db.GetGORM().WithContext(ctx).Create(log).Error; err != nil {
		return r.handleError(err, "create api call log")
	}
	return nil
}

// GetByRequestID 根据请求ID获取APICallLog
func (r *APICallLogRepo) GetByRequestID(ctx context.Context, requestID string) (*entity.APICallLog, error) {
	var log entity.APICallLog
	err := r.db.GetGORM().WithContext(ctx).
		Preload("Channel").
		Preload("Processor").
		Where("request_id = ?", requestID).
		First(&log).Error
	if err != nil {
		return nil, r.handleError(err, "get api call log by request id")
	}
	return &log, nil
}

// GetByChannelID 根据通道ID获取APICallLog列表（分页）
func (r *APICallLogRepo) GetByChannelID(ctx context.Context, channelID string, params port.ListParams) (*port.PaginatedResult[*entity.APICallLog], error) {
	internalParams := convertToInternalParams(params)
	internalParams = r.validateParams(internalParams)

	var logs []*entity.APICallLog
	query := r.db.GetGORM().WithContext(ctx).
		Model(&entity.APICallLog{}).
		Where("channel_id = ?", channelID)

	query = r.buildQuery(query, internalParams)

	result, err := PaginateTypedForPort(r.db.GetGORM(), ctx, query, internalParams, &logs)
	if err != nil {
		return nil, r.handleError(err, "get api call logs by channel id")
	}

	return result, nil
}

// GetByProcessorID 根据处理器ID获取APICallLog列表（分页）
func (r *APICallLogRepo) GetByProcessorID(ctx context.Context, processorID string, params port.ListParams) (*port.PaginatedResult[*entity.APICallLog], error) {
	internalParams := convertToInternalParams(params)
	internalParams = r.validateParams(internalParams)

	var logs []*entity.APICallLog
	query := r.db.GetGORM().WithContext(ctx).
		Model(&entity.APICallLog{}).
		Where("processor_id = ?", processorID)

	query = r.buildQuery(query, internalParams)

	result, err := PaginateTypedForPort(r.db.GetGORM(), ctx, query, internalParams, &logs)
	if err != nil {
		return nil, r.handleError(err, "get api call logs by processor id")
	}

	return result, nil
}

// GetSlowCalls 获取慢调用记录
func (r *APICallLogRepo) GetSlowCalls(ctx context.Context, thresholdMs int, limit int) ([]*entity.APICallLog, error) {
	var logs []*entity.APICallLog
	query := r.db.GetGORM().WithContext(ctx).
		Preload("Channel").
		Preload("Processor").
		Where("duration_ms > ?", thresholdMs).
		Order("duration_ms DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	if err != nil {
		return nil, r.handleError(err, "get slow api calls")
	}
	return logs, nil
}

// GetFailedCalls 获取失败的调用记录
func (r *APICallLogRepo) GetFailedCalls(ctx context.Context, limit int) ([]*entity.APICallLog, error) {
	var logs []*entity.APICallLog
	query := r.db.GetGORM().WithContext(ctx).
		Preload("Channel").
		Preload("Processor").
		Where("success = ?", false).
		Order("start_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	if err != nil {
		return nil, r.handleError(err, "get failed api calls")
	}
	return logs, nil
}

// GetRecentCalls 获取最近的调用记录
func (r *APICallLogRepo) GetRecentCalls(ctx context.Context, limit int) ([]*entity.APICallLog, error) {
	var logs []*entity.APICallLog
	query := r.db.GetGORM().WithContext(ctx).
		Preload("Channel").
		Preload("Processor").
		Order("start_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	if err != nil {
		return nil, r.handleError(err, "get recent api calls")
	}
	return logs, nil
}

// List 获取APICallLog列表（分页）
func (r *APICallLogRepo) List(ctx context.Context, params port.ListParams) (*port.PaginatedResult[*entity.APICallLog], error) {
	internalParams := convertToInternalParams(params)
	internalParams = r.validateParams(internalParams)

	var logs []*entity.APICallLog
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.APICallLog{}), internalParams)

	result, err := PaginateTypedForPort(r.db.GetGORM(), ctx, query, internalParams, &logs)
	if err != nil {
		return nil, r.handleError(err, "list api call logs")
	}

	return result, nil
}

// Delete 删除APICallLog（软删除）
func (r *APICallLogRepo) Delete(ctx context.Context, id string) error {
	return r.softDelete(ctx, &entity.APICallLog{}, id)
}

// DeleteOldLogs 删除指定天数前的日志
func (r *APICallLogRepo) DeleteOldLogs(ctx context.Context, beforeDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -beforeDays)
	return r.hardDelete(ctx, &entity.APICallLog{}, "start_time < ?", cutoffTime)
}

// GetStatistics 获取API调用统计信息
func (r *APICallLogRepo) GetStatistics(ctx context.Context, channelID *string, processorID *string) (*port.CallStatistics, error) {
	var stats port.CallStatistics

	query := r.db.GetGORM().WithContext(ctx).Model(&entity.APICallLog{})

	if channelID != nil {
		query = query.Where("channel_id = ?", *channelID)
	}
	if processorID != nil {
		query = query.Where("processor_id = ?", *processorID)
	}

	// 总调用次数
	if err := query.Count(&stats.TotalCalls).Error; err != nil {
		return nil, r.handleError(err, "count total calls")
	}

	// 成功调用次数
	if err := query.Where("success = ?", true).Count(&stats.SuccessCalls).Error; err != nil {
		return nil, r.handleError(err, "count success calls")
	}

	// 失败调用次数
	stats.FailedCalls = stats.TotalCalls - stats.SuccessCalls

	// 成功率
	if stats.TotalCalls > 0 {
		stats.SuccessRate = float64(stats.SuccessCalls) / float64(stats.TotalCalls) * 100
	}

	// 平均耗时、最大耗时、最小耗时
	var avgDuration, maxDuration, minDuration sql.NullFloat64
	row := query.Select("AVG(duration_ms) as avg_duration, MAX(duration_ms) as max_duration, MIN(duration_ms) as min_duration").Row()
	if err := row.Scan(&avgDuration, &maxDuration, &minDuration); err != nil {
		return nil, r.handleError(err, "calculate duration statistics")
	}

	if avgDuration.Valid {
		stats.AvgDuration = avgDuration.Float64
	}
	if maxDuration.Valid {
		stats.MaxDuration = int(maxDuration.Float64)
	}
	if minDuration.Valid {
		stats.MinDuration = int(minDuration.Float64)
	}

	return &stats, nil
}

// 确保 APICallLogRepo 实现了 port.APICallLogRepository 接口
var _ port.APICallLogRepository = (*APICallLogRepo)(nil)
