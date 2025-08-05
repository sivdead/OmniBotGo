package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// CreateChannel 创建通道
// @Summary 创建新的通道
// @Description 为指定机器人创建新的平台通道
// @Tags channels
// @Accept json
// @Produce json
// @Param request body usecase.CreateChannelRequest true "创建通道请求"
// @Success 201 {object} StandardResponse{data=entity.Channel} "创建成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 409 {object} StandardResponse "通道已存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/channels [post]
func (v *V1) CreateChannel(c *fiber.Ctx) error {
	var req usecase.CreateChannelRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	result, err := v.channelUC.CreateChannel(c.Context(), req)
	if err != nil {
		v.l.Error("创建通道失败: %v", err)
		if err.Error() == "通道已存在" {
			return NewErrorResponse(c, http.StatusConflict, err.Error())
		}
		return NewErrorResponse(c, http.StatusInternalServerError, "创建通道失败")
	}

	v.l.Info("通道创建成功: %d", result.ID)

	return NewSuccessResponse(c, http.StatusCreated, "通道创建成功", result)
}

// GetChannels 获取通道列表
// @Summary 获取通道列表
// @Description 根据条件查询通道列表，支持分页和过滤
// @Tags channels
// @Accept json
// @Produce json
// @Param bot_id query int false "机器人ID"
// @Param platform_type query string false "平台类型"
// @Param status query string false "状态"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认20" default(20)
// @Success 200 {object} StandardResponse{data=usecase.ChannelListResult} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/channels [get]
func (v *V1) GetChannels(c *fiber.Ctx) error {
	var req usecase.ListChannelsParams
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
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("查询参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	result, err := v.channelUC.ListChannels(c.Context(), req)
	if err != nil {
		v.l.Error("获取通道列表失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取通道列表失败")
	}

	v.l.Debug("获取通道列表成功，共 %d 条记录", result.Total)

	return NewSuccessResponse(c, http.StatusOK, "获取成功", result)
}

// GetChannelByID 获取通道详情
// @Summary 获取通道详情
// @Description 根据ID获取指定通道的详细信息
// @Tags channels
// @Accept json
// @Produce json
// @Param id path string true "通道ID"
// @Success 200 {object} StandardResponse{data=entity.Channel} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "通道不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/channels/{id} [get]
func (v *V1) GetChannelByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "通道ID不能为空")
	}

	result, err := v.channelUC.GetChannel(c.Context(), id)
	if err != nil {
		v.l.Error("获取通道详情失败: %v", err)
		if err.Error() == "通道不存在" {
			return NewErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return NewErrorResponse(c, http.StatusInternalServerError, "获取通道详情失败")
	}

	v.l.Debug("获取通道详情成功: %s", id)

	return NewSuccessResponse(c, http.StatusOK, "获取成功", result)
}

// UpdateChannel 更新通道
// @Summary 更新通道配置
// @Description 更新指定通道的配置信息
// @Tags channels
// @Accept json
// @Produce json
// @Param id path string true "通道ID"
// @Param request body usecase.UpdateChannelRequest true "更新通道请求"
// @Success 200 {object} StandardResponse{data=entity.Channel} "更新成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "通道不存在"
// @Failure 409 {object} StandardResponse "通道名称已存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/channels/{id} [put]
func (v *V1) UpdateChannel(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "通道ID不能为空")
	}

	var req usecase.UpdateChannelRequest
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

	result, err := v.channelUC.UpdateChannel(c.Context(), req)
	if err != nil {
		v.l.Error("更新通道失败: %v", err)
		switch err.Error() {
		case "通道不存在":
			return NewErrorResponse(c, http.StatusNotFound, err.Error())
		case "通道名称已存在":
			return NewErrorResponse(c, http.StatusConflict, err.Error())
		default:
			return NewErrorResponse(c, http.StatusInternalServerError, "更新通道失败")
		}
	}

	v.l.Info("通道更新成功: %s", id)

	return NewSuccessResponse(c, http.StatusOK, "更新成功", result)
}

// DeleteChannel 删除通道
// @Summary 删除通道
// @Description 删除指定的通道配置
// @Tags channels
// @Accept json
// @Produce json
// @Param id path string true "通道ID"
// @Success 200 {object} StandardResponse "删除成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "通道不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/channels/{id} [delete]
func (v *V1) DeleteChannel(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "通道ID不能为空")
	}

	err := v.channelUC.DeleteChannel(c.Context(), id)
	if err != nil {
		v.l.Error("删除通道失败: %v", err)
		if err.Error() == "通道不存在" {
			return NewErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return NewErrorResponse(c, http.StatusInternalServerError, "删除通道失败")
	}

	v.l.Info("通道删除成功: %s", id)

	return NewSuccessResponse(c, http.StatusOK, "删除成功", nil)
}

// UpdateChannelStatusRequest 更新通道状态请求
type UpdateChannelStatusRequest struct {
	Status entity.ConnectionStatus `json:"status" validate:"required"`
}

// UpdateChannelStatus 更新通道状态
// @Summary 更新通道连接状态
// @Description 更新指定通道的连接状态
// @Tags channels
// @Accept json
// @Produce json
// @Param id path string true "通道ID"
// @Param request body UpdateChannelStatusRequest true "更新状态请求"
// @Success 200 {object} StandardResponse "更新成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "通道不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/channels/{id}/status [patch]
func (v *V1) UpdateChannelStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "通道ID不能为空")
	}

	var req UpdateChannelStatusRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	err := v.channelUC.UpdateChannelStatus(c.Context(), id, req.Status)
	if err != nil {
		v.l.Error("更新通道状态失败: %v", err)
		if err.Error() == "通道不存在" {
			return NewErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return NewErrorResponse(c, http.StatusInternalServerError, "更新通道状态失败")
	}

	v.l.Info("通道状态更新成功: %s", id)

	return NewSuccessResponse(c, http.StatusOK, "状态更新成功", nil)
}

// RefreshChannelToken 刷新通道令牌
// @Summary 刷新通道访问令牌
// @Description 刷新指定通道的访问令牌
// @Tags channels
// @Accept json
// @Produce json
// @Param id path string true "通道ID"
// @Success 200 {object} StandardResponse "刷新成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "通道不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/channels/{id}/refresh-token [post]
func (v *V1) RefreshChannelToken(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "通道ID不能为空")
	}

	err := v.channelUC.RefreshChannelToken(c.Context(), id)
	if err != nil {
		v.l.Error("刷新通道令牌失败: %v", err)
		if err.Error() == "通道不存在" {
			return NewErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return NewErrorResponse(c, http.StatusInternalServerError, "刷新通道令牌失败")
	}

	v.l.Info("通道令牌刷新成功: %s", id)

	return NewSuccessResponse(c, http.StatusOK, "令牌刷新成功", nil)
}
