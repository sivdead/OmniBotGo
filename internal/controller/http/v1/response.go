package v1

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// StandardResponse 标准API响应结构
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}

// APIResponse 通用API响应结构（用于Swagger文档，与StandardResponse相同）
type APIResponse = StandardResponse

// NewErrorResponse 创建错误响应
func NewErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(StandardResponse{
		Success: false,
		Error:   message,
		Code:    statusCode,
	})
}

// NewSuccessResponse 创建成功响应（兼容不同参数格式）
func NewSuccessResponse(c *fiber.Ctx, args ...interface{}) error {
	response := StandardResponse{
		Success: true,
	}

	// 处理不同的参数格式
	switch len(args) {
	case 1:
		// NewSuccessResponse(c, data)
		response.Data = args[0]
	case 2:
		// NewSuccessResponse(c, data, message) 或 NewSuccessResponse(c, statusCode, message)
		if statusCode, ok := args[0].(int); ok {
			response.Message = args[1].(string)
			return c.Status(statusCode).JSON(response)
		} else {
			response.Data = args[0]
			response.Message = args[1].(string)
		}
	case 3:
		// NewSuccessResponse(c, statusCode, message, data)
		statusCode := args[0].(int)
		response.Message = args[1].(string)
		response.Data = args[2]
		return c.Status(statusCode).JSON(response)
	}

	return c.Status(StatusOK).JSON(response)
}

// NewValidationErrorResponse 创建验证错误响应
func NewValidationErrorResponse(c *fiber.Ctx, err error) error {
	var message string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		message = "参数验证失败："
		for _, ve := range validationErrors {
			message += ve.Field() + "字段验证失败 "
		}
	} else {
		message = "参数验证失败: " + err.Error()
	}

	return NewErrorResponse(c, StatusBadRequest, message)
}
