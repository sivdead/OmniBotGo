package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// CreateBot 创建机器人
// @Summary 创建新的机器人
// @Description 创建新的机器人配置
// @Tags bots
// @Accept json
// @Produce json
// @Param request body usecase.CreateBotRequest true "创建机器人请求"
// @Success 201 {object} StandardResponse{data=usecase.CreateBotResponse} "创建成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 409 {object} StandardResponse "机器人名称已存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/bots [post]
func (v *V1) CreateBot(c *fiber.Ctx) error {
	var req usecase.CreateBotRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return errorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return errorResponse(c, http.StatusBadRequest, "请求参数验证失败")
	}

	result, err := v.botUC.CreateBot(c.Context(), &req)
	if err != nil {
		v.l.Error("创建机器人失败: %v", err)
		if err.Error() == "机器人名称已存在" {
			return errorResponse(c, http.StatusConflict, err.Error())
		}
		return errorResponse(c, http.StatusInternalServerError, "创建机器人失败")
	}

	v.l.Info("机器人创建成功: %s", result.BotID)

	return c.Status(http.StatusCreated).JSON(map[string]interface{}{
		"success": true,
		"message": "机器人创建成功",
		"data":    result,
	})
}

// GetBots 获取机器人列表
// @Summary 获取机器人列表
// @Description 根据条件查询机器人列表，支持分页和过滤
// @Tags bots
// @Accept json
// @Produce json
// @Param status query string false "状态" Enums(active,inactive)
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认20" default(20)
// @Success 200 {object} StandardResponse{data=usecase.ListBotsResponse} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/bots [get]
func (v *V1) GetBots(c *fiber.Ctx) error {
	var req usecase.ListBotsRequest
	if err := c.QueryParser(&req); err != nil {
		v.l.Error("解析查询参数失败: %v", err)
		return errorResponse(c, http.StatusBadRequest, "查询参数格式错误")
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
		return errorResponse(c, http.StatusBadRequest, "查询参数验证失败")
	}

	result, err := v.botUC.ListBots(c.Context(), &req)
	if err != nil {
		v.l.Error("获取机器人列表失败: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "获取机器人列表失败")
	}

	v.l.Debug("获取机器人列表成功，共 %d 条记录", result.Total)

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"success": true,
		"message": "获取成功",
		"data":    result,
	})
}

// GetBotByID 获取机器人详情
// @Summary 获取机器人详情
// @Description 根据ID获取指定机器人的详细信息
// @Tags bots
// @Accept json
// @Produce json
// @Param id path string true "机器人ID"
// @Success 200 {object} StandardResponse{data=usecase.GetBotByIDResponse} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "机器人不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/bots/{id} [get]
func (v *V1) GetBotByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errorResponse(c, http.StatusBadRequest, "机器人ID不能为空")
	}

	result, err := v.botUC.GetBotByID(c.Context(), &usecase.GetBotByIDRequest{
		ID: id,
	})
	if err != nil {
		v.l.Error("获取机器人详情失败: %v", err)
		if err.Error() == "机器人不存在" {
			return errorResponse(c, http.StatusNotFound, err.Error())
		}
		return errorResponse(c, http.StatusInternalServerError, "获取机器人详情失败")
	}

	v.l.Debug("获取机器人详情成功: %s", id)

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"success": true,
		"message": "获取成功",
		"data":    result,
	})
}

// UpdateBot 更新机器人
// @Summary 更新机器人配置
// @Description 更新指定机器人的配置信息
// @Tags bots
// @Accept json
// @Produce json
// @Param id path string true "机器人ID"
// @Param request body usecase.UpdateBotRequest true "更新机器人请求"
// @Success 200 {object} StandardResponse{data=usecase.UpdateBotResponse} "更新成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "机器人不存在"
// @Failure 409 {object} StandardResponse "机器人名称已存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/bots/{id} [put]
func (v *V1) UpdateBot(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errorResponse(c, http.StatusBadRequest, "机器人ID不能为空")
	}

	var req usecase.UpdateBotRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return errorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return errorResponse(c, http.StatusBadRequest, "请求参数验证失败")
	}

	// 设置ID
	req.ID = id

	result, err := v.botUC.UpdateBot(c.Context(), &req)
	if err != nil {
		v.l.Error("更新机器人失败: %v", err)
		switch err.Error() {
		case "机器人不存在":
			return errorResponse(c, http.StatusNotFound, err.Error())
		case "机器人名称已存在":
			return errorResponse(c, http.StatusConflict, err.Error())
		default:
			return errorResponse(c, http.StatusInternalServerError, "更新机器人失败")
		}
	}

	v.l.Info("机器人更新成功: %s", id)

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"success": true,
		"message": "更新成功",
		"data":    result,
	})
}

// DeleteBot 删除机器人
// @Summary 删除机器人
// @Description 删除指定的机器人配置
// @Tags bots
// @Accept json
// @Produce json
// @Param id path string true "机器人ID"
// @Success 200 {object} StandardResponse "删除成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "机器人不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/bots/{id} [delete]
func (v *V1) DeleteBot(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errorResponse(c, http.StatusBadRequest, "机器人ID不能为空")
	}

	err := v.botUC.DeleteBot(c.Context(), &usecase.DeleteBotRequest{
		ID: id,
	})
	if err != nil {
		v.l.Error("删除机器人失败: %v", err)
		if err.Error() == "机器人不存在" {
			return errorResponse(c, http.StatusNotFound, err.Error())
		}
		return errorResponse(c, http.StatusInternalServerError, "删除机器人失败")
	}

	v.l.Info("机器人删除成功: %s", id)

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"success": true,
		"message": "删除成功",
	})
}
