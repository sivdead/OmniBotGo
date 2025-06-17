package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/adapter"
)

// HealthController 健康检查控制器
type HealthController struct {
	*V1
	adapterManager *adapter.Manager
}

// NewHealthController 创建健康检查控制器实例
func NewHealthController(v1 *V1, adapterManager *adapter.Manager) *HealthController {
	return &HealthController{
		V1:             v1,
		adapterManager: adapterManager,
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status      string                 `json:"status"`
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      string                 `json:"uptime"`
	Database    DatabaseHealth         `json:"database"`
	Adapters    []AdapterHealth        `json:"adapters"`
	Channels    ChannelHealthSummary   `json:"channels"`
	Messages    MessageHealthSummary   `json:"messages"`
	Environment map[string]interface{} `json:"environment"`
}

// DatabaseHealth 数据库健康状态
type DatabaseHealth struct {
	Status      string `json:"status"`
	Ping        string `json:"ping"`
	Connections struct {
		Active int `json:"active"`
		Idle   int `json:"idle"`
		Total  int `json:"total"`
	} `json:"connections"`
}

// AdapterHealth 适配器健康状态
type AdapterHealth struct {
	Platform     string `json:"platform"`
	Status       string `json:"status"`
	LastError    string `json:"last_error,omitempty"`
	ChannelCount int    `json:"channel_count"`
}

// ChannelHealthSummary 通道健康状态摘要
type ChannelHealthSummary struct {
	Total      int `json:"total"`
	Connected  int `json:"connected"`
	Connecting int `json:"connecting"`
	Failed     int `json:"failed"`
}

// MessageHealthSummary 消息健康状态摘要
type MessageHealthSummary struct {
	Total24h    int64   `json:"total_24h"`
	Pending     int64   `json:"pending"`
	Processing  int64   `json:"processing"`
	Processed   int64   `json:"processed"`
	Failed      int64   `json:"failed"`
	SuccessRate float64 `json:"success_rate"`
}

// StartTime 应用启动时间
var StartTime = time.Now()

// GetHealth 获取系统健康状态
// @Summary 获取系统健康状态
// @Description 返回系统各组件的健康状态信息
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} StandardResponse{data=HealthResponse} "系统健康"
// @Failure 503 {object} StandardResponse{data=HealthResponse} "系统异常"
// @Router /healthz [get]
func (h *HealthController) GetHealth(c *fiber.Ctx) error {
	ctx := c.Context()

	response := HealthResponse{
		Status:    "healthy",
		Version:   "1.0.0", // TODO: 从配置或构建信息获取
		Timestamp: time.Now(),
		Uptime:    time.Since(StartTime).String(),
		Environment: map[string]interface{}{
			"go_version": "1.21",        // TODO: 从runtime获取
			"platform":   "linux/amd64", // TODO: 从runtime获取
		},
	}

	// 检查数据库健康状态
	dbHealth := h.checkDatabaseHealth(ctx)
	response.Database = dbHealth

	// 检查适配器健康状态
	adapterHealths := h.checkAdaptersHealth(ctx)
	response.Adapters = adapterHealths

	// 检查通道健康状态
	channelHealth := h.checkChannelsHealth(ctx)
	response.Channels = channelHealth

	// 检查消息健康状态
	messageHealth := h.checkMessagesHealth(ctx)
	response.Messages = messageHealth

	// 确定整体健康状态
	if dbHealth.Status != "healthy" {
		response.Status = "unhealthy"
	}

	for _, adapter := range adapterHealths {
		if adapter.Status == "unhealthy" {
			response.Status = "degraded"
			break
		}
	}

	// 根据健康状态返回适当的HTTP状态码
	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	return NewSuccessResponse(c, statusCode, "健康检查完成", response)
}

// GetHealthSimple 获取简单健康状态
// @Summary 获取简单健康状态
// @Description 返回简化的系统健康状态，用于负载均衡器健康检查
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "系统健康"
// @Failure 503 {object} map[string]string "系统异常"
// @Router /health [get]
func (h *HealthController) GetHealthSimple(c *fiber.Ctx) error {
	ctx := c.Context()

	// 只检查关键组件
	dbHealth := h.checkDatabaseHealth(ctx)

	if dbHealth.Status == "healthy" {
		return c.JSON(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	}

	c.Status(http.StatusServiceUnavailable)
	return c.JSON(map[string]string{
		"status": "unhealthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// GetReadiness 获取就绪状态
// @Summary 获取应用就绪状态
// @Description 检查应用是否已经准备好接收请求
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "应用就绪"
// @Failure 503 {object} map[string]string "应用未就绪"
// @Router /ready [get]
func (h *HealthController) GetReadiness(c *fiber.Ctx) error {
	ctx := c.Context()

	// 检查数据库连接
	dbHealth := h.checkDatabaseHealth(ctx)
	if dbHealth.Status != "healthy" {
		c.Status(http.StatusServiceUnavailable)
		return c.JSON(map[string]string{
			"status": "not ready",
			"reason": "database not available",
			"time":   time.Now().Format(time.RFC3339),
		})
	}

	// TODO: 检查其他必需的依赖服务

	return c.JSON(map[string]string{
		"status": "ready",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// GetLiveness 获取存活状态
// @Summary 获取应用存活状态
// @Description 检查应用是否仍在运行
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "应用存活"
// @Router /live [get]
func (h *HealthController) GetLiveness(c *fiber.Ctx) error {
	return c.JSON(map[string]string{
		"status": "alive",
		"time":   time.Now().Format(time.RFC3339),
		"uptime": time.Since(StartTime).String(),
	})
}

// checkDatabaseHealth 检查数据库健康状态
func (h *HealthController) checkDatabaseHealth(ctx context.Context) DatabaseHealth {
	health := DatabaseHealth{
		Status: "healthy",
		Ping:   "ok",
	}

	// TODO: 实际检查数据库连接
	// 这里需要添加实际的数据库ping检查
	// 可以通过依赖注入获取数据库连接池

	return health
}

// checkAdaptersHealth 检查适配器健康状态
func (h *HealthController) checkAdaptersHealth(ctx context.Context) []AdapterHealth {
	var healths []AdapterHealth

	supportedPlatforms := h.adapterManager.GetSupportedPlatforms()

	for _, platform := range supportedPlatforms {
		health := AdapterHealth{
			Platform: string(platform),
			Status:   "healthy",
		}

		// 检查该平台的通道数量
		// TODO: 通过channelUC获取通道统计信息
		health.ChannelCount = 0

		healths = append(healths, health)
	}

	return healths
}

// checkChannelsHealth 检查通道健康状态
func (h *HealthController) checkChannelsHealth(ctx context.Context) ChannelHealthSummary {
	summary := ChannelHealthSummary{}

	// TODO: 通过channelUC获取通道统计信息
	// channels, err := h.channelUC.GetChannelStatistics(ctx)
	// if err != nil {
	//     h.l.Error("获取通道统计失败", "error", err)
	//     return summary
	// }

	return summary
}

// checkMessagesHealth 检查消息健康状态
func (h *HealthController) checkMessagesHealth(ctx context.Context) MessageHealthSummary {
	summary := MessageHealthSummary{}

	// TODO: 通过messageUC获取消息统计信息
	// stats, err := h.messageUC.GetMessageStatistics(ctx, time.Now().Add(-24*time.Hour))
	// if err != nil {
	//     h.l.Error("获取消息统计失败", "error", err)
	//     return summary
	// }

	return summary
}
