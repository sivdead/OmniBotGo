package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// CreateProcessor 创建处理器
// @Summary 创建新的消息处理器
// @Description 创建新的消息处理器配置
// @Tags processors
// @Accept json
// @Produce json
// @Param request body usecase.CreateProcessorRequest true "创建处理器请求"
// @Success 201 {object} StandardResponse{data=entity.MessageProcessor} "创建成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 409 {object} StandardResponse "处理器名称已存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/processors [post]
func (v *V1) CreateProcessor(c *fiber.Ctx) error {
	var req usecase.CreateProcessorRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	result, err := v.processorUC.CreateProcessor(c.Context(), req)
	if err != nil {
		v.l.Error("创建处理器失败: %v", err)
		if err.Error() == "处理器名称已存在" {
			return NewErrorResponse(c, http.StatusConflict, err.Error())
		}
		return NewErrorResponse(c, http.StatusInternalServerError, "创建处理器失败")
	}

	return NewSuccessResponse(c, http.StatusCreated, "处理器创建成功", result)
}

// GetProcessors 获取处理器列表
// @Summary 获取处理器列表
// @Description 根据条件查询处理器列表，支持分页和过滤
// @Tags processors
// @Accept json
// @Produce json
// @Param status query string false "状态"
// @Param type query string false "处理器类型"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认20" default(20)
// @Success 200 {object} StandardResponse{data=usecase.ProcessorListResult} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/processors [get]
func (v *V1) GetProcessors(c *fiber.Ctx) error {
	var req usecase.ListProcessorsParams
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

	result, err := v.processorUC.ListProcessors(c.Context(), req)
	if err != nil {
		v.l.Error("获取处理器列表失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取处理器列表失败")
	}

	return NewSuccessResponse(c, result)
}

// GetProcessorByID 获取处理器详情
// @Summary 获取处理器详情
// @Description 根据ID获取指定处理器的详细信息
// @Tags processors
// @Accept json
// @Produce json
// @Param id path string true "处理器ID"
// @Success 200 {object} StandardResponse{data=entity.MessageProcessor} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "处理器不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/processors/{id} [get]
func (v *V1) GetProcessorByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "处理器ID不能为空")
	}

	result, err := v.processorUC.GetProcessor(c.Context(), id)
	if err != nil {
		v.l.Error("获取处理器详情失败: %v", err)
		return NewErrorResponse(c, http.StatusNotFound, "处理器不存在")
	}

	return NewSuccessResponse(c, result)
}

// UpdateProcessor 更新处理器
// @Summary 更新处理器配置
// @Description 更新指定处理器的配置信息
// @Tags processors
// @Accept json
// @Produce json
// @Param id path string true "处理器ID"
// @Param request body usecase.UpdateProcessorRequest true "更新处理器请求"
// @Success 200 {object} StandardResponse{data=entity.MessageProcessor} "更新成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "处理器不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/processors/{id} [put]
func (v *V1) UpdateProcessor(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "处理器ID不能为空")
	}

	var req usecase.UpdateProcessorRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	// 设置ID
	req.ID = id

	result, err := v.processorUC.UpdateProcessor(c.Context(), req)
	if err != nil {
		v.l.Error("更新处理器失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "更新处理器失败")
	}

	return NewSuccessResponse(c, result)
}

// DeleteProcessor 删除处理器
// @Summary 删除处理器
// @Description 删除指定的处理器配置
// @Tags processors
// @Accept json
// @Produce json
// @Param id path string true "处理器ID"
// @Success 200 {object} StandardResponse "删除成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "处理器不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/processors/{id} [delete]
func (v *V1) DeleteProcessor(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "处理器ID不能为空")
	}

	err := v.processorUC.DeleteProcessor(c.Context(), id)
	if err != nil {
		v.l.Error("删除处理器失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "删除处理器失败")
	}

	return NewSuccessResponse(c, map[string]interface{}{
		"message": "处理器删除成功",
		"id":      id,
	})
}

// UpdateProcessorStatus 更新处理器状态
// @Summary 更新处理器状态
// @Description 启用或禁用指定的处理器
// @Tags processors
// @Accept json
// @Produce json
// @Param id path string true "处理器ID"
// @Param request body usecase.UpdateProcessorStatusRequest true "更新处理器状态请求"
// @Success 200 {object} StandardResponse "更新成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "处理器不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/processors/{id}/status [patch]
func (v *V1) UpdateProcessorStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "处理器ID不能为空")
	}

	var req usecase.UpdateProcessorStatusRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	// 设置ID
	req.ID = id

	err := v.processorUC.UpdateProcessorStatus(c.Context(), req)
	if err != nil {
		v.l.Error("更新处理器状态失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "更新处理器状态失败")
	}

	return NewSuccessResponse(c, map[string]interface{}{
		"message": "处理器状态更新成功",
		"id":      id,
	})
}

// CreateRoutingRule 创建路由规则
// @Summary 创建路由规则
// @Description 为指定处理器创建路由规则
// @Tags processors
// @Accept json
// @Produce json
// @Param id path string true "处理器ID"
// @Param request body usecase.CreateRoutingRuleRequest true "创建路由规则请求"
// @Success 201 {object} StandardResponse{data=entity.MessageRoutingRule} "创建成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "处理器不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/processors/{id}/rules [post]
func (v *V1) CreateRoutingRule(c *fiber.Ctx) error {
	processorID := c.Params("id")
	if processorID == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "处理器ID不能为空")
	}

	var req usecase.CreateRoutingRuleRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	// 设置处理器ID
	req.ProcessorID = processorID

	result, err := v.processorUC.CreateRoutingRule(c.Context(), req)
	if err != nil {
		v.l.Error("创建路由规则失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "创建路由规则失败")
	}

	return NewSuccessResponse(c, http.StatusCreated, "路由规则创建成功", result)
}

// GetRoutingRules 获取路由规则列表
// @Summary 获取路由规则列表
// @Description 获取指定处理器的路由规则列表
// @Tags processors
// @Accept json
// @Produce json
// @Param id path string true "处理器ID"
// @Success 200 {object} StandardResponse{data=[]entity.MessageRoutingRule} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "处理器不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/processors/{id}/rules [get]
func (v *V1) GetRoutingRules(c *fiber.Ctx) error {
	processorID := c.Params("id")
	if processorID == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "处理器ID不能为空")
	}

	var params usecase.ListRoutingRulesParams
	if err := c.QueryParser(&params); err != nil {
		v.l.Error("解析查询参数失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "查询参数格式错误")
	}

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	result, err := v.processorUC.ListRoutingRules(c.Context(), processorID, params)
	if err != nil {
		v.l.Error("获取路由规则列表失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取路由规则列表失败")
	}

	return NewSuccessResponse(c, result)
}

// UpdateRoutingRule 更新路由规则
// @Summary 更新路由规则
// @Description 更新指定的路由规则配置
// @Tags processors
// @Accept json
// @Produce json
// @Param id path string true "处理器ID"
// @Param rule_id path string true "规则ID"
// @Param request body usecase.UpdateRoutingRuleRequest true "更新路由规则请求"
// @Success 200 {object} StandardResponse{data=entity.MessageRoutingRule} "更新成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "规则不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/processors/{id}/rules/{rule_id} [put]
func (v *V1) UpdateRoutingRule(c *fiber.Ctx) error {
	processorID := c.Params("id")
	ruleID := c.Params("rule_id")

	if processorID == "" || ruleID == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "处理器ID和规则ID不能为空")
	}

	var req usecase.UpdateRoutingRuleRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	// 设置ID
	req.ID = ruleID
	req.ProcessorID = &processorID

	result, err := v.processorUC.UpdateRoutingRule(c.Context(), req)
	if err != nil {
		v.l.Error("更新路由规则失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "更新路由规则失败")
	}

	return NewSuccessResponse(c, result)
}

// DeleteRoutingRule 删除路由规则
// @Summary 删除路由规则
// @Description 删除指定的路由规则
// @Tags processors
// @Accept json
// @Produce json
// @Param id path string true "处理器ID"
// @Param rule_id path string true "规则ID"
// @Success 200 {object} StandardResponse "删除成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "规则不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/processors/{id}/rules/{rule_id} [delete]
func (v *V1) DeleteRoutingRule(c *fiber.Ctx) error {
	processorID := c.Params("id")
	ruleID := c.Params("rule_id")

	if processorID == "" || ruleID == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "处理器ID和规则ID不能为空")
	}

	err := v.processorUC.DeleteRoutingRule(c.Context(), ruleID)
	if err != nil {
		v.l.Error("删除路由规则失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "删除路由规则失败")
	}

	return NewSuccessResponse(c, map[string]interface{}{
		"message":      "路由规则删除成功",
		"processor_id": processorID,
		"rule_id":      ruleID,
	})
}
