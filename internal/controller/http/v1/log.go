package v1

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// GetConnectionLogs 获取连接日志列表
// @Summary 获取连接日志列表
// @Description 获取平台连接日志记录，支持分页和过滤
// @Tags logs
// @Accept json
// @Produce json
// @Param channel_id query int false "通道ID"
// @Param status query string false "连接状态"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认20" default(20)
// @Success 200 {object} StandardResponse{data=usecase.ConnectionLogListResult} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/logs/connections [get]
func (v *V1) GetConnectionLogs(c *fiber.Ctx) error {
	var req usecase.ListConnectionLogsParams
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

	result, err := v.logUC.ListConnectionLogs(c.Context(), req)
	if err != nil {
		v.l.Error("获取连接日志失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取连接日志失败")
	}

	return NewSuccessResponse(c, result)
}

// GetConnectionLogByID 获取连接日志详情
// @Summary 获取连接日志详情
// @Description 根据ID获取指定连接日志的详细信息
// @Tags logs
// @Accept json
// @Produce json
// @Param id path string true "日志ID"
// @Success 200 {object} StandardResponse{data=entity.ConnectionLog} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "日志不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/logs/connections/{id} [get]
func (v *V1) GetConnectionLogByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "日志ID不能为空")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return NewErrorResponse(c, http.StatusBadRequest, "日志ID格式错误")
	}

	result, err := v.logUC.GetConnectionLog(c.Context(), id)
	if err != nil {
		v.l.Error("获取连接日志详情失败: %v", err)
		return NewErrorResponse(c, http.StatusNotFound, "日志不存在")
	}

	return NewSuccessResponse(c, result)
}

// GetAPICallLogs 获取API调用日志列表
// @Summary 获取API调用日志列表
// @Description 获取API调用日志记录，支持分页和过滤
// @Tags logs
// @Accept json
// @Produce json
// @Param method query string false "HTTP方法"
// @Param path query string false "请求路径"
// @Param status_code query int false "状态码"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认20" default(20)
// @Success 200 {object} StandardResponse{data=usecase.APICallLogListResult} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/logs/api-calls [get]
func (v *V1) GetAPICallLogs(c *fiber.Ctx) error {
	var req usecase.ListAPICallLogsParams
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

	result, err := v.logUC.ListAPICallLogs(c.Context(), req)
	if err != nil {
		v.l.Error("获取API调用日志失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取API调用日志失败")
	}

	return NewSuccessResponse(c, result)
}

// GetAPICallLogByID 获取API调用日志详情
// @Summary 获取API调用日志详情
// @Description 根据ID获取指定API调用日志的详细信息
// @Tags logs
// @Accept json
// @Produce json
// @Param id path string true "日志ID"
// @Success 200 {object} StandardResponse{data=entity.APICallLog} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "日志不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/logs/api-calls/{id} [get]
func (v *V1) GetAPICallLogByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "日志ID不能为空")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return NewErrorResponse(c, http.StatusBadRequest, "日志ID格式错误")
	}

	result, err := v.logUC.GetAPICallLog(c.Context(), id)
	if err != nil {
		v.l.Error("获取API调用日志详情失败: %v", err)
		return NewErrorResponse(c, http.StatusNotFound, "日志不存在")
	}

	return NewSuccessResponse(c, result)
}

// GetAPICallStats 获取API调用统计
// @Summary 获取API调用统计
// @Description 获取API调用的统计信息，包括成功率、响应时间等
// @Tags logs
// @Accept json
// @Produce json
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param group_by query string false "分组方式" Enums(hour,day,week,month)
// @Success 200 {object} StandardResponse{data=usecase.APICallStatsResult} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/logs/api-calls/stats [get]
func (v *V1) GetAPICallStats(c *fiber.Ctx) error {
	var req usecase.GetAPICallStatsParams
	if err := c.QueryParser(&req); err != nil {
		v.l.Error("解析查询参数失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "查询参数格式错误")
	}

	result, err := v.logUC.GetAPICallStats(c.Context(), req)
	if err != nil {
		v.l.Error("获取API调用统计失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取API调用统计失败")
	}

	return NewSuccessResponse(c, result)
}
