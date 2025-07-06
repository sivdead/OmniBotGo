package usecase

import (
	"context"
	"fmt"
	"runtime"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// monitorUC 监控用例实现
type monitorUC struct {
	botRepo     port.BotRepository
	channelRepo port.ChannelRepository
	messageRepo port.MessageRepository
	queueRepo   port.MessageQueueRepository
	logger      logger.Interface
}

// NewMonitorUseCase 创建监控用例
func NewMonitorUseCase(
	botRepo port.BotRepository,
	channelRepo port.ChannelRepository,
	messageRepo port.MessageRepository,
	queueRepo port.MessageQueueRepository,
	logger logger.Interface,
) MonitorUseCase {
	return &monitorUC{
		botRepo:     botRepo,
		channelRepo: channelRepo,
		messageRepo: messageRepo,
		queueRepo:   queueRepo,
		logger:      logger,
	}
}

// GetSystemOverview 获取系统概览
func (uc *monitorUC) GetSystemOverview(ctx context.Context) (*SystemOverviewResult, error) {
	// 获取机器人总数
	botParams := port.ListParams{
		Page:     1,
		PageSize: 1, // 只需要总数
	}
	botResult, err := uc.botRepo.List(ctx, botParams)
	if err != nil {
		uc.logger.Error("failed to get bot count", "error", err)
		return nil, fmt.Errorf("failed to get bot count: %w", err)
	}

	// 获取通道总数和活跃通道数
	channelParams := port.ListParams{
		Page:     1,
		PageSize: 1,
	}
	channelResult, err := uc.channelRepo.List(ctx, channelParams)
	if err != nil {
		uc.logger.Error("failed to get channel count", "error", err)
		return nil, fmt.Errorf("failed to get channel count: %w", err)
	}

	// 获取活跃通道
	activeChannels, err := uc.channelRepo.ListActive(ctx)
	if err != nil {
		uc.logger.Error("failed to get active channels", "error", err)
		return nil, fmt.Errorf("failed to get active channels: %w", err)
	}

	// 获取消息总数
	messageParams := port.ListParams{
		Page:     1,
		PageSize: 1,
	}
	messageResult, err := uc.messageRepo.List(ctx, messageParams)
	if err != nil {
		uc.logger.Error("failed to get message count", "error", err)
		return nil, fmt.Errorf("failed to get message count: %w", err)
	}

	// 获取队列状态
	queueParams := port.ListParams{
		Page:     1,
		PageSize: 1,
		Filters: map[string]interface{}{
			"status": entity.QueueStatusPending,
		},
	}
	pendingResult, err := uc.queueRepo.List(ctx, queueParams)
	if err != nil {
		uc.logger.Error("failed to get pending jobs", "error", err)
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}

	queueParams.Filters["status"] = entity.QueueStatusFailed
	failedResult, err := uc.queueRepo.List(ctx, queueParams)
	if err != nil {
		uc.logger.Error("failed to get failed jobs", "error", err)
		return nil, fmt.Errorf("failed to get failed jobs: %w", err)
	}

	return &SystemOverviewResult{
		TotalBots:      botResult.Total,
		TotalChannels:  channelResult.Total,
		ActiveChannels: int64(len(activeChannels)),
		TotalMessages:  messageResult.Total,
		PendingJobs:    pendingResult.Total,
		FailedJobs:     failedResult.Total,
	}, nil
}

// GetSystemMetrics 获取系统指标
func (uc *monitorUC) GetSystemMetrics(ctx context.Context, params GetSystemMetricsParams) (*SystemMetricsResult, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 计算CPU使用率（这是一个简化的实现）
	cpuUsage := float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 10 // 简化计算

	// 计算内存使用率
	memoryUsage := float64(m.Alloc) / float64(m.Sys) * 100

	metrics := map[string]float64{
		"goroutines":     float64(runtime.NumGoroutine()),
		"num_cpu":        float64(runtime.NumCPU()),
		"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
		"sys_mb":         float64(m.Sys) / 1024 / 1024,
		"gc_count":       float64(m.NumGC),
		"heap_alloc_mb":  float64(m.HeapAlloc) / 1024 / 1024,
		"heap_sys_mb":    float64(m.HeapSys) / 1024 / 1024,
		"heap_idle_mb":   float64(m.HeapIdle) / 1024 / 1024,
		"heap_inuse_mb":  float64(m.HeapInuse) / 1024 / 1024,
		"stack_inuse_mb": float64(m.StackInuse) / 1024 / 1024,
	}

	// 根据请求的指标类型过滤
	if params.MetricType != nil {
		filteredMetrics := make(map[string]float64)
		switch *params.MetricType {
		case "memory":
			for k, v := range metrics {
				if k == "alloc_mb" || k == "sys_mb" || k == "heap_alloc_mb" || k == "heap_sys_mb" {
					filteredMetrics[k] = v
				}
			}
			metrics = filteredMetrics
		case "goroutine":
			filteredMetrics["goroutines"] = metrics["goroutines"]
			metrics = filteredMetrics
		case "gc":
			filteredMetrics["gc_count"] = metrics["gc_count"]
			metrics = filteredMetrics
		}
	}

	return &SystemMetricsResult{
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		Metrics:     metrics,
	}, nil
}

// GetDetailedHealth 获取详细健康检查
func (uc *monitorUC) GetDetailedHealth(ctx context.Context) (*DetailedHealthResult, error) {
	components := make(map[string]ComponentHealth)

	// 检查数据库连接
	dbHealth := ComponentHealth{
		Status:  "healthy",
		Message: "Database connection is healthy",
	}
	// 尝试查询一条记录来验证数据库连接
	_, err := uc.botRepo.List(ctx, port.ListParams{Page: 1, PageSize: 1})
	if err != nil {
		dbHealth.Status = "unhealthy"
		dbHealth.Message = fmt.Sprintf("Database error: %v", err)
	}
	components["database"] = dbHealth

	// 检查通道连接状态
	activeChannels, err := uc.channelRepo.ListActive(ctx)
	channelHealth := ComponentHealth{
		Status:  "healthy",
		Message: fmt.Sprintf("%d active channels", len(activeChannels)),
	}
	if err != nil {
		channelHealth.Status = "unhealthy"
		channelHealth.Message = fmt.Sprintf("Failed to get active channels: %v", err)
	} else if len(activeChannels) == 0 {
		channelHealth.Status = "warning"
		channelHealth.Message = "No active channels"
	}
	components["channels"] = channelHealth

	// 检查队列状态
	pendingJobs, err := uc.queueRepo.GetPendingJobs(ctx, "", 1)
	queueHealth := ComponentHealth{
		Status:  "healthy",
		Message: "Queue is healthy",
	}
	if err != nil {
		queueHealth.Status = "unhealthy"
		queueHealth.Message = fmt.Sprintf("Queue error: %v", err)
	} else if len(pendingJobs) > 100 {
		queueHealth.Status = "warning"
		queueHealth.Message = fmt.Sprintf("High number of pending jobs: %d", len(pendingJobs))
	}
	components["queue"] = queueHealth

	// 检查内存使用
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryHealth := ComponentHealth{
		Status:  "healthy",
		Message: fmt.Sprintf("Memory usage: %.2f MB", float64(m.Alloc)/1024/1024),
	}
	if float64(m.Alloc)/float64(m.Sys) > 0.9 {
		memoryHealth.Status = "warning"
		memoryHealth.Message = fmt.Sprintf("High memory usage: %.2f%%", float64(m.Alloc)/float64(m.Sys)*100)
	}
	components["memory"] = memoryHealth

	// 整体状态
	overallStatus := "healthy"
	for _, component := range components {
		if component.Status == "unhealthy" {
			overallStatus = "unhealthy"
			break
		} else if component.Status == "warning" && overallStatus == "healthy" {
			overallStatus = "warning"
		}
	}

	return &DetailedHealthResult{
		Status:     overallStatus,
		Components: components,
	}, nil
}
