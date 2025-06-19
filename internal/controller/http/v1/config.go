package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// GetSystemConfigs 获取系统配置列表
// @Summary 获取系统配置列表
// @Description 获取系统配置项列表，支持分页和过滤
// @Tags configs
// @Accept json
// @Produce json
// @Param group query string false "配置组"
// @Param key query string false "配置键"
// @Param is_system query bool false "是否为系统配置"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认20" default(20)
// @Success 200 {object} StandardResponse{data=usecase.SystemConfigListResult} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/configs [get]
func (v *V1) GetSystemConfigs(c *fiber.Ctx) error {
	var req usecase.ListSystemConfigsParams
	if err := c.QueryParser(&req); err != nil {
		v.l.Error("解析查询参数失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "查询参数格式错误")
	}

	// 设置默认值
	setDefaultPagination(&req.Page, &req.PageSize)

	// 调用usecase处理系统配置查询
	result, err := v.systemConfigUC.ListSystemConfigs(c.Context(), req)
	if err != nil {
		v.l.Error("获取系统配置失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取系统配置失败")
	}

	return NewSuccessResponse(c, result)
}

// GetSystemConfigByKey 获取系统配置详情
// @Summary 获取系统配置详情
// @Description 根据键获取指定系统配置的值
// @Tags configs
// @Accept json
// @Produce json
// @Param key path string true "配置键"
// @Success 200 {object} StandardResponse{data=entity.SystemConfig} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "配置不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/configs/{key} [get]
func (v *V1) GetSystemConfigByKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "配置键不能为空")
	}

	// 调用usecase获取系统配置详情
	result, err := v.systemConfigUC.GetSystemConfig(c.Context(), key)
	if err != nil {
		v.l.Error("获取系统配置详情失败: %v", err)
		return NewErrorResponse(c, http.StatusNotFound, "配置不存在")
	}

	return NewSuccessResponse(c, result)
}

// CreateSystemConfig 创建系统配置
// @Summary 创建系统配置
// @Description 创建新的系统配置项
// @Tags configs
// @Accept json
// @Produce json
// @Param request body usecase.CreateSystemConfigRequest true "创建系统配置请求"
// @Success 201 {object} StandardResponse{data=entity.SystemConfig} "创建成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 409 {object} StandardResponse "配置已存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/configs [post]
func (v *V1) CreateSystemConfig(c *fiber.Ctx) error {
	var req usecase.CreateSystemConfigRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	// 调用usecase创建系统配置
	result, err := v.systemConfigUC.CreateSystemConfig(c.Context(), req)
	if err != nil {
		v.l.Error("创建系统配置失败: %v", err)
		if err.Error() == "配置已存在" {
			return NewErrorResponse(c, http.StatusConflict, err.Error())
		}
		return NewErrorResponse(c, http.StatusInternalServerError, "创建系统配置失败")
	}

	return NewSuccessResponse(c, http.StatusCreated, "配置创建成功", result)
}

// UpdateSystemConfig 更新系统配置
// @Summary 更新系统配置
// @Description 更新指定系统配置的值
// @Tags configs
// @Accept json
// @Produce json
// @Param key path string true "配置键"
// @Param request body usecase.UpdateSystemConfigRequest true "更新系统配置请求"
// @Success 200 {object} StandardResponse{data=entity.SystemConfig} "更新成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "配置不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/configs/{key} [put]
func (v *V1) UpdateSystemConfig(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "配置键不能为空")
	}

	var req usecase.UpdateSystemConfigRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	// 设置键
	req.Key = key

	// 调用usecase更新系统配置
	result, err := v.systemConfigUC.UpdateSystemConfig(c.Context(), req)
	if err != nil {
		v.l.Error("更新系统配置失败: %v", err)
		if err.Error() == "配置不存在" {
			return NewErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return NewErrorResponse(c, http.StatusInternalServerError, "更新系统配置失败")
	}

	return NewSuccessResponse(c, result)
}

// DeleteSystemConfig 删除系统配置
// @Summary 删除系统配置
// @Description 删除指定的系统配置项
// @Tags configs
// @Accept json
// @Produce json
// @Param key path string true "配置键"
// @Success 200 {object} StandardResponse "删除成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "配置不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/configs/{key} [delete]
func (v *V1) DeleteSystemConfig(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "配置键不能为空")
	}

	// 调用usecase删除系统配置
	err := v.systemConfigUC.DeleteSystemConfig(c.Context(), key)
	if err != nil {
		v.l.Error("删除系统配置失败: %v", err)
		if err.Error() == "配置不存在" {
			return NewErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return NewErrorResponse(c, http.StatusInternalServerError, "删除系统配置失败")
	}

	return NewSuccessResponse(c, "删除成功")
}

// GetSystemConfigsByGroup 根据组获取系统配置
// @Summary 根据组获取系统配置
// @Description 获取指定组的所有系统配置项
// @Tags configs
// @Accept json
// @Produce json
// @Param group path string true "配置组"
// @Success 200 {object} StandardResponse{data=[]entity.SystemConfig} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/configs/groups/{group} [get]
func (v *V1) GetSystemConfigsByGroup(c *fiber.Ctx) error {
	group := c.Params("group")
	if group == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "配置组不能为空")
	}

	// 调用usecase根据组获取系统配置
	result, err := v.systemConfigUC.GetSystemConfigsByGroup(c.Context(), group)
	if err != nil {
		v.l.Error("根据组获取系统配置失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取系统配置失败")
	}

	return NewSuccessResponse(c, result)
}
