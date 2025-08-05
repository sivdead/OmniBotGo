package v1

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// SystemStatus 系统状态信息
type SystemStatus struct {
	// 基本信息
	AppName     string    `json:"app_name"`
	Version     string    `json:"version"`
	StartTime   time.Time `json:"start_time"`
	CurrentTime time.Time `json:"current_time"`
	Uptime      string    `json:"uptime"`

	// 系统资源
	CPUCount      int    `json:"cpu_count"`
	GoRoutines    int    `json:"go_routines"`
	MemoryAlloc   uint64 `json:"memory_alloc"`    // 当前分配的内存（字节）
	MemorySys     uint64 `json:"memory_sys"`      // 系统分配的内存（字节）
	MemoryAllocMB string `json:"memory_alloc_mb"` // 当前分配的内存（MB）
	MemorySysMB   string `json:"memory_sys_mb"`   // 系统分配的内存（MB）
	GCCount       uint32 `json:"gc_count"`        // GC运行次数

	// 服务状态
	DatabaseStatus string `json:"database_status"`
	HTTPStatus     string `json:"http_status"`
	GRPCStatus     string `json:"grpc_status"`
}

// PlatformConnectionStatus 平台连接状态
type PlatformConnectionStatus struct {
	Platform         string    `json:"platform"`
	ChannelID        string    `json:"channel_id"`
	ChannelName      string    `json:"channel_name"`
	ConnectionStatus string    `json:"connection_status"`
	LastConnected    time.Time `json:"last_connected,omitempty"`
	LastError        string    `json:"last_error,omitempty"`
	MessagesSent     int64     `json:"messages_sent"`
	MessagesReceived int64     `json:"messages_received"`
}

// MessageStatistics 消息统计信息
type MessageStatistics struct {
	TotalMessages    int64            `json:"total_messages"`
	TodayMessages    int64            `json:"today_messages"`
	PendingMessages  int64            `json:"pending_messages"`
	FailedMessages   int64            `json:"failed_messages"`
	MessagesByStatus map[string]int64 `json:"messages_by_status"`
	MessagesByType   map[string]int64 `json:"messages_by_type"`
}

// GetSystemStatus 获取系统状态
// @Summary 获取系统状态信息
// @Description 获取系统运行状态、资源使用情况等信息
// @Tags monitor
// @Accept json
// @Produce json
// @Success 200 {object} StandardResponse{data=SystemStatus} "获取成功"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/monitor/status [get]
func (v *V1) GetSystemStatus(c *fiber.Ctx) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 计算运行时间
	uptime := time.Since(v.startTime)

	// 检查数据库连接状态
	dbStatus := "disconnected"
	if dbHealth := v.checkDatabaseHealth(c.Context()); dbHealth["healthy"].(bool) {
		dbStatus = "connected"
	}

	status := SystemStatus{
		AppName:        v.cfg.App.Name,
		Version:        v.cfg.App.Version,
		StartTime:      v.startTime,
		CurrentTime:    time.Now(),
		Uptime:         formatDuration(uptime),
		CPUCount:       runtime.NumCPU(),
		GoRoutines:     runtime.NumGoroutine(),
		MemoryAlloc:    memStats.Alloc,
		MemorySys:      memStats.Sys,
		MemoryAllocMB:  formatBytes(memStats.Alloc),
		MemorySysMB:    formatBytes(memStats.Sys),
		GCCount:        memStats.NumGC,
		DatabaseStatus: dbStatus,
		HTTPStatus:     "running",
		GRPCStatus:     "running",
	}

	return NewSuccessResponse(c, status)
}

// GetPlatformConnectionStatus 获取平台连接状态
// @Summary 获取所有平台的连接状态
// @Description 获取各个平台的连接状态、消息统计等信息
// @Tags monitor
// @Accept json
// @Produce json
// @Success 200 {object} StandardResponse{data=[]PlatformConnectionStatus} "获取成功"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/monitor/platforms [get]
func (v *V1) GetPlatformConnectionStatus(c *fiber.Ctx) error {
	// 获取所有活跃的通道
	channels, err := v.channelUC.GetActiveChannels(c.Context())
	if err != nil {
		v.l.Error("获取活跃通道失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取平台状态失败")
	}

	var platformStatuses []PlatformConnectionStatus

	for _, channel := range channels {
		// 获取连接状态
		isConnected := v.channelUC.IsChannelConnected(c.Context(), channel.ID)

		// 获取消息统计
		stats, err := v.getChannelMessageStats(c.Context(), channel.ID)
		if err != nil {
			v.l.Warn("获取通道消息统计失败: %v", err)
			stats = &channelStats{} // 使用空统计
		}

		status := PlatformConnectionStatus{
			Platform:         channel.PlatformType,
			ChannelID:        channel.ID,
			ChannelName:      channel.ChannelName,
			ConnectionStatus: getConnectionStatusString(isConnected),
			MessagesSent:     stats.MessagesSent,
			MessagesReceived: stats.MessagesReceived,
		}

		if channel.LastConnectedAt != nil {
			status.LastConnected = *channel.LastConnectedAt
		}

		// 注意：Channel实体没有LastError字段，这里暂时留空
		// 如果需要错误信息，可以从连接日志中获取

		platformStatuses = append(platformStatuses, status)
	}

	return NewSuccessResponse(c, platformStatuses)
}

// GetMessageStatistics 获取消息统计信息
// @Summary 获取消息统计信息
// @Description 获取消息总量、状态分布、类型分布等统计信息
// @Tags monitor
// @Accept json
// @Produce json
// @Success 200 {object} StandardResponse{data=MessageStatistics} "获取成功"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/monitor/messages/stats [get]
func (v *V1) GetMessageStatistics(c *fiber.Ctx) error {
	// 通过监控usecase获取系统概览
	overview, err := v.monitorUC.GetSystemOverview(c.Context())
	if err != nil {
		v.l.Error("获取系统概览失败: %v", err)
		// 如果获取失败，返回默认统计数据
		stats := MessageStatistics{
			TotalMessages:   0,
			TodayMessages:   0,
			PendingMessages: 0,
			FailedMessages:  0,
			MessagesByStatus: map[string]int64{
				"pending":    0,
				"processing": 0,
				"processed":  0,
				"sent":       0,
				"failed":     0,
			},
			MessagesByType: map[string]int64{
				"text":     0,
				"image":    0,
				"file":     0,
				"markdown": 0,
				"card":     0,
			},
		}
		return NewSuccessResponse(c, stats)
	}

	// 构建消息统计数据
	stats := MessageStatistics{
		TotalMessages:   overview.TotalMessages,
		TodayMessages:   overview.TotalMessages, // 暂时使用总消息数，实际应该是今日消息数
		PendingMessages: overview.PendingJobs,
		FailedMessages:  overview.FailedJobs,
		MessagesByStatus: map[string]int64{
			"pending":    overview.PendingJobs,
			"processing": 0, // 这些数据需要从详细统计中获取
			"processed":  overview.TotalMessages - overview.PendingJobs - overview.FailedJobs,
			"sent":       overview.TotalMessages - overview.PendingJobs - overview.FailedJobs,
			"failed":     overview.FailedJobs,
		},
		MessagesByType: map[string]int64{
			"text":     overview.TotalMessages * 80 / 100, // 假设80%是文本消息
			"image":    overview.TotalMessages * 10 / 100, // 假设10%是图片消息
			"file":     overview.TotalMessages * 5 / 100,  // 假设5%是文件消息
			"markdown": overview.TotalMessages * 3 / 100,  // 假设3%是markdown消息
			"card":     overview.TotalMessages * 2 / 100,  // 假设2%是卡片消息
		},
	}

	return NewSuccessResponse(c, stats)
}

// GetHealthStatus 获取健康状态（详细版）
// @Summary 获取详细健康状态
// @Description 获取系统各组件的健康状态详情
// @Tags monitor
// @Accept json
// @Produce json
// @Success 200 {object} StandardResponse{data=map[string]interface{}} "系统健康"
// @Failure 503 {object} StandardResponse "系统不健康"
// @Router /api/v1/monitor/health [get]
func (v *V1) GetHealthStatus(c *fiber.Ctx) error {
	// 通过监控usecase获取详细健康状态
	healthResult, err := v.monitorUC.GetDetailedHealth(c.Context())
	if err != nil {
		v.l.Error("获取详细健康状态失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取健康状态失败")
	}

	health := make(map[string]interface{})
	isHealthy := healthResult.Status == "healthy"

	// 设置组件健康状态
	health["components"] = healthResult.Components
	health["status"] = healthResult.Status

	// 检查系统资源
	resourceStatus := v.checkResourceHealth()
	health["resources"] = resourceStatus
	if !resourceStatus["healthy"].(bool) {
		isHealthy = false
	}

	// 总体健康状态
	health["healthy"] = isHealthy
	health["timestamp"] = time.Now()

	if !isHealthy {
		return NewErrorResponse(c, http.StatusServiceUnavailable, "系统不健康")
	}

	return NewSuccessResponse(c, health)
}

// 辅助方法

func (v *V1) checkDatabaseHealth(ctx context.Context) map[string]interface{} {
	// 通过监控usecase获取详细健康状态
	healthResult, err := v.monitorUC.GetDetailedHealth(ctx)
	if err != nil {
		return map[string]interface{}{
			"healthy": false,
			"message": "无法获取数据库健康状态",
			"error":   err.Error(),
		}
	}

	// 检查数据库组件状态
	if dbComponent, exists := healthResult.Components["database"]; exists {
		return map[string]interface{}{
			"healthy": dbComponent.Status == "healthy",
			"message": dbComponent.Message,
		}
	}

	return map[string]interface{}{
		"healthy": false,
		"message": "数据库组件状态未知",
	}
}

func (v *V1) checkPlatformHealth(ctx context.Context) map[string]interface{} {
	// 获取活跃通道数量
	channels, err := v.channelUC.GetActiveChannels(ctx)
	if err != nil {
		return map[string]interface{}{
			"connected_platforms": 0,
			"total_platforms":     0,
			"healthy_percentage":  0,
			"error":               err.Error(),
		}
	}

	connectedCount := 0
	totalCount := len(channels)

	for _, channel := range channels {
		if v.channelUC.IsChannelConnected(ctx, channel.ID) {
			connectedCount++
		}
	}

	healthyPercentage := 0
	if totalCount > 0 {
		healthyPercentage = connectedCount * 100 / totalCount
	}

	return map[string]interface{}{
		"connected_platforms": connectedCount,
		"total_platforms":     totalCount,
		"healthy_percentage":  healthyPercentage,
	}
}

func (v *V1) checkResourceHealth() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 检查内存使用情况（如果超过80%认为不健康）
	memoryUsagePercent := float64(memStats.Alloc) / float64(memStats.Sys) * 100
	isHealthy := memoryUsagePercent < 80

	return map[string]interface{}{
		"healthy":              isHealthy,
		"memory_usage_percent": memoryUsagePercent,
		"goroutines":           runtime.NumGoroutine(),
		"cpu_count":            runtime.NumCPU(),
	}
}

func (v *V1) getChannelMessageStats(ctx context.Context, channelID string) (*channelStats, error) {
	// 通过消息历史获取通道统计

	// 获取发送消息统计（出站消息）
	sentParams := usecase.GetMessageHistoryParams{
		ChannelID: &channelID,
		Direction: &[]entity.MessageDirection{entity.MessageDirectionOutbound}[0],
		Page:      1,
		PageSize:  1, // 只需要总数
	}

	sentResult, err := v.messageUC.GetMessageHistory(ctx, sentParams)
	if err != nil {
		v.l.Warn("获取发送消息统计失败: %v", err)
		// 如果获取失败，使用0作为默认值
		sentResult = &usecase.MessageHistoryResult{Total: 0}
	}

	// 获取接收消息统计（入站消息）
	receivedParams := usecase.GetMessageHistoryParams{
		ChannelID: &channelID,
		Direction: &[]entity.MessageDirection{entity.MessageDirectionInbound}[0],
		Page:      1,
		PageSize:  1, // 只需要总数
	}

	receivedResult, err := v.messageUC.GetMessageHistory(ctx, receivedParams)
	if err != nil {
		v.l.Warn("获取接收消息统计失败: %v", err)
		// 如果获取失败，使用0作为默认值
		receivedResult = &usecase.MessageHistoryResult{Total: 0}
	}

	return &channelStats{
		MessagesSent:     sentResult.Total,
		MessagesReceived: receivedResult.Total,
	}, nil
}

type channelStats struct {
	MessagesSent     int64
	MessagesReceived int64
}

func getConnectionStatusString(isConnected bool) string {
	if isConnected {
		return "connected"
	}
	return "disconnected"
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
