package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// GetPlatforms 获取支持的平台列表
// @Summary 获取支持的平台列表
// @Description 获取系统支持的所有消息平台列表
// @Tags platforms
// @Accept json
// @Produce json
// @Success 200 {object} StandardResponse{data=[]usecase.PlatformInfo} "获取成功"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/platforms [get]
func (v *V1) GetPlatforms(c *fiber.Ctx) error {
	// 调用usecase获取平台列表
	result, err := v.platformUC.GetPlatforms(c.Context())
	if err != nil {
		v.l.Error("获取平台列表失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取平台列表失败")
	}

	return NewSuccessResponse(c, result)
}

// GetPlatformByType 获取平台详情
// @Summary 获取平台详情
// @Description 根据平台类型获取平台的详细信息
// @Tags platforms
// @Accept json
// @Produce json
// @Param type path string true "平台类型" Enums(wecom,dingtalk,wechat_official,feishu)
// @Success 200 {object} StandardResponse{data=usecase.PlatformInfo} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "平台不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/platforms/{type} [get]
func (v *V1) GetPlatformByType(c *fiber.Ctx) error {
	platformType := c.Params("type")
	if platformType == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "平台类型不能为空")
	}

	// 验证平台类型
	var validPlatformTypes = map[string]bool{
		string(entity.PlatformTypeWecom):          true,
		string(entity.PlatformTypeDingtalk):       true,
		string(entity.PlatformTypeWechatOfficial): true,
		string(entity.PlatformTypeFeishu):         true,
	}

	if !validPlatformTypes[platformType] {
		return NewErrorResponse(c, http.StatusBadRequest, "不支持的平台类型")
	}

	// 调用usecase获取平台详情
	result, err := v.platformUC.GetPlatformByType(c.Context(), platformType)
	if err != nil {
		v.l.Error("获取平台详情失败: %v", err)
		return NewErrorResponse(c, http.StatusNotFound, "平台不存在")
	}

	return NewSuccessResponse(c, result)
}

// ValidatePlatformConfig 验证平台配置
// @Summary 验证平台配置
// @Description 验证指定平台的配置是否正确
// @Tags platforms
// @Accept json
// @Produce json
// @Param type path string true "平台类型" Enums(wecom,dingtalk,wechat_official,feishu)
// @Param request body usecase.ValidatePlatformConfigRequest true "验证平台配置请求"
// @Success 200 {object} StandardResponse{data=usecase.PlatformConfigValidationResult} "验证成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/platforms/{type}/validate [post]
func (v *V1) ValidatePlatformConfig(c *fiber.Ctx) error {
	platformType := c.Params("type")
	if platformType == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "平台类型不能为空")
	}

	var req usecase.ValidatePlatformConfigRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	// 调用usecase验证平台配置
	result, err := v.platformUC.ValidatePlatformConfig(c.Context(), req)
	if err != nil {
		v.l.Error("验证平台配置失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "验证平台配置失败")
	}

	return NewSuccessResponse(c, result)
}

// GetPlatformStatus 获取平台状态
// @Summary 获取平台状态
// @Description 获取指定平台的运行状态
// @Tags platforms
// @Accept json
// @Produce json
// @Param type path string true "平台类型" Enums(wecom,dingtalk,wechat_official,feishu)
// @Success 200 {object} StandardResponse{data=usecase.PlatformStatusResult} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/platforms/{type}/status [get]
func (v *V1) GetPlatformStatus(c *fiber.Ctx) error {
	platformType := c.Params("type")
	if platformType == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "平台类型不能为空")
	}

	// 调用usecase获取平台状态
	result, err := v.platformUC.GetPlatformStatus(c.Context(), platformType)
	if err != nil {
		v.l.Error("获取平台状态失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取平台状态失败")
	}

	return NewSuccessResponse(c, result)
}
