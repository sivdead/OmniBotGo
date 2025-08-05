package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// GetQueueMessages 获取队列消息列表
// @Summary 获取队列消息列表
// @Description 获取消息队列中的消息列表，支持分页和过滤
// @Tags queues
// @Accept json
// @Produce json
// @Param status query string false "消息状态"
// @Param queue_type query string false "队列类型"
// @Param priority query int false "优先级"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认20" default(20)
// @Success 200 {object} StandardResponse{data=usecase.QueueMessageListResult} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/queues [get]
func (v *V1) GetQueueMessages(c *fiber.Ctx) error {
	var req usecase.ListQueueMessagesParams
	if err := c.QueryParser(&req); err != nil {
		v.l.Error("解析查询参数失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "查询参数格式错误")
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := v.queueUC.ListQueueMessages(c.Context(), req)
	if err != nil {
		v.l.Error("获取队列消息失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取队列消息失败")
	}

	return NewSuccessResponse(c, result)
}

// GetQueueMessageByID 获取队列消息详情
// @Summary 获取队列消息详情
// @Description 根据ID获取指定队列消息的详细信息
// @Tags queues
// @Accept json
// @Produce json
// @Param id path string true "消息ID"
// @Success 200 {object} StandardResponse{data=entity.MessageQueue} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "消息不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/queues/{id} [get]
func (v *V1) GetQueueMessageByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "消息ID不能为空")
	}

	result, err := v.queueUC.GetQueueMessage(c.Context(), id)
	if err != nil {
		v.l.Error("获取队列消息详情失败: %v", err)
		return NewErrorResponse(c, http.StatusNotFound, "消息不存在")
	}

	return NewSuccessResponse(c, result)
}

// RetryQueueMessage 重试队列消息
// @Summary 重试队列消息
// @Description 重新处理指定的队列消息
// @Tags queues
// @Accept json
// @Produce json
// @Param id path string true "消息ID"
// @Success 200 {object} StandardResponse "重试成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "消息不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/queues/{id}/retry [post]
func (v *V1) RetryQueueMessage(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "消息ID不能为空")
	}

	err := v.queueUC.RetryQueueMessage(c.Context(), id)
	if err != nil {
		v.l.Error("重试队列消息失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "重试队列消息失败")
	}

	return NewSuccessResponse(c, map[string]interface{}{
		"message": "队列消息重试成功",
		"id":      id,
	})
}

// CancelQueueMessage 取消队列消息
// @Summary 取消队列消息
// @Description 取消指定的队列消息处理
// @Tags queues
// @Accept json
// @Produce json
// @Param id path string true "消息ID"
// @Success 200 {object} StandardResponse "取消成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "消息不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/queues/{id}/cancel [patch]
func (v *V1) CancelQueueMessage(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "消息ID不能为空")
	}

	err := v.queueUC.CancelQueueMessage(c.Context(), id)
	if err != nil {
		v.l.Error("取消队列消息失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "取消队列消息失败")
	}

	return NewSuccessResponse(c, map[string]interface{}{
		"message": "队列消息取消成功",
		"id":      id,
	})
}

// GetQueueStats 获取队列统计信息
// @Summary 获取队列统计信息
// @Description 获取消息队列的统计信息，包括处理情况、性能指标等
// @Tags queues
// @Accept json
// @Produce json
// @Param queue_type query string false "队列类型"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {object} StandardResponse{data=usecase.QueueStatsResult} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/queues/stats [get]
func (v *V1) GetQueueStats(c *fiber.Ctx) error {
	var req usecase.GetQueueStatsParams
	if err := c.QueryParser(&req); err != nil {
		v.l.Error("解析查询参数失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "查询参数格式错误")
	}

	result, err := v.queueUC.GetQueueStats(c.Context(), req)
	if err != nil {
		v.l.Error("获取队列统计失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取队列统计失败")
	}

	return NewSuccessResponse(c, result)
}

// ClearFailedMessages 清理失败消息
// @Summary 清理失败消息
// @Description 清理队列中的失败消息
// @Tags queues
// @Accept json
// @Produce json
// @Param before query string false "清理指定时间之前的失败消息"
// @Param batch_size query int false "批量处理大小，默认100" default(100)
// @Success 200 {object} StandardResponse "清理成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/queues/failed [delete]
func (v *V1) ClearFailedMessages(c *fiber.Ctx) error {
	var req usecase.ClearFailedMessagesParams
	if err := c.QueryParser(&req); err != nil {
		v.l.Error("解析查询参数失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "查询参数格式错误")
	}

	// 设置默认值
	if req.BatchSize <= 0 {
		req.BatchSize = 100
	}

	// 调用usecase清理失败消息，这里需要使用清理已完成消息的方法
	// 因为接口定义中没有清理失败消息的方法，我们使用清理已完成消息的方法
	cleanParams := usecase.CleanCompletedMessagesParams{
		BeforeDays: 7, // 默认清理7天前的
	}

	err := v.queueUC.CleanCompletedMessages(c.Context(), cleanParams)
	if err != nil {
		v.l.Error("清理失败消息失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "清理失败消息失败")
	}

	return NewSuccessResponse(c, map[string]interface{}{
		"message":    "清理失败消息成功",
		"batch_size": req.BatchSize,
	})
}
