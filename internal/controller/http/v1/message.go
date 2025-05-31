package v1

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ChannelID       int64                  `json:"channel_id" validate:"required"`
	MessageType     string                 `json:"message_type" validate:"required"`
	ContentType     string                 `json:"content_type,omitempty"`
	ReceiverID      string                 `json:"receiver_id" validate:"required"`
	ReceiverName    string                 `json:"receiver_name,omitempty"`
	ReceiverType    string                 `json:"receiver_type,omitempty"`
	Content         string                 `json:"content" validate:"required"`
	MediaURL        string                 `json:"media_url,omitempty"`
	MediaType       string                 `json:"media_type,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ParentMessageID *int64                 `json:"parent_message_id,omitempty"`
}

// GetMessageHistoryRequest 获取消息历史请求
type GetMessageHistoryRequest struct {
	ChannelID     *int64                   `query:"channel_id"`
	SenderID      *string                  `query:"sender_id"`
	ReceiverID    *string                  `query:"receiver_id"`
	MessageType   *string                  `query:"message_type"`
	MessageStatus *entity.MessageStatus    `query:"message_status"`
	Direction     *entity.MessageDirection `query:"direction"`
	StartTime     *string                  `query:"start_time"`
	EndTime       *string                  `query:"end_time"`
	Page          int                      `query:"page" validate:"min=1"`
	PageSize      int                      `query:"page_size" validate:"min=1,max=100"`
}

// SendMessage 发送消息
// @Summary 发送消息
// @Description 向指定平台发送消息
// @Tags messages
// @Accept json
// @Produce json
// @Param message body SendMessageRequest true "消息内容"
// @Success 200 {object} APIResponse{data=entity.Message}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/messages/send [post]
func (h *V1) SendMessage(c *fiber.Ctx) error {
	var req SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		h.l.Error("解析请求参数失败: %v", err)
		return NewErrorResponse(c, fiber.StatusBadRequest, "请求参数格式错误")
	}

	// 验证请求参数
	if err := h.v.Struct(&req); err != nil {
		h.l.Error("请求参数验证失败: %v", err)
		return NewErrorResponse(c, fiber.StatusBadRequest, "请求参数验证失败: "+err.Error())
	}

	// 创建消息实体
	message := &entity.Message{
		ChannelID:       req.ChannelID,
		MessageType:     req.MessageType,
		ContentType:     req.ContentType,
		ReceiverID:      req.ReceiverID,
		ReceiverName:    req.ReceiverName,
		ReceiverType:    req.ReceiverType,
		Content:         req.Content,
		MediaURL:        req.MediaURL,
		MediaType:       req.MediaType,
		ParentMessageID: req.ParentMessageID,
	}

	// 设置元数据
	if req.Metadata != nil {
		message.UnifiedContent = entity.JSONField(req.Metadata)
	}

	// 调用业务逻辑发送消息
	if err := h.messageUC.SendMessage(c.Context(), message); err != nil {
		h.l.Error("发送消息失败: %v", err)
		return NewErrorResponse(c, fiber.StatusInternalServerError, "发送消息失败: "+err.Error())
	}

	return NewSuccessResponse(c, message, "消息发送成功")
}

// GetMessageHistory 获取消息历史
// @Summary 获取消息历史
// @Description 根据条件查询消息历史记录
// @Tags messages
// @Accept json
// @Produce json
// @Param channel_id query int false "通道ID"
// @Param sender_id query string false "发送者ID"
// @Param receiver_id query string false "接收者ID"
// @Param message_type query string false "消息类型"
// @Param message_status query int false "消息状态"
// @Param direction query int false "消息方向"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param page query int true "页码" default(1)
// @Param page_size query int true "每页大小" default(20)
// @Success 200 {object} APIResponse{data=usecase.MessageHistoryResult}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/messages/history [get]
func (h *V1) GetMessageHistory(c *fiber.Ctx) error {
	var req GetMessageHistoryRequest
	if err := c.QueryParser(&req); err != nil {
		h.l.Error("解析查询参数失败: %v", err)
		return NewErrorResponse(c, fiber.StatusBadRequest, "查询参数格式错误")
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 验证请求参数
	if err := h.v.Struct(&req); err != nil {
		h.l.Error("请求参数验证失败: %v", err)
		return NewErrorResponse(c, fiber.StatusBadRequest, "请求参数验证失败: "+err.Error())
	}

	// 构建查询参数
	params := usecase.GetMessageHistoryParams{
		ChannelID:     req.ChannelID,
		SenderID:      req.SenderID,
		ReceiverID:    req.ReceiverID,
		MessageType:   req.MessageType,
		MessageStatus: req.MessageStatus,
		Direction:     req.Direction,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Page:          req.Page,
		PageSize:      req.PageSize,
	}

	// 调用业务逻辑获取消息历史
	result, err := h.messageUC.GetMessageHistory(c.Context(), params)
	if err != nil {
		h.l.Error("获取消息历史失败: %v", err)
		return NewErrorResponse(c, fiber.StatusInternalServerError, "获取消息历史失败: "+err.Error())
	}

	return NewSuccessResponse(c, result, "获取消息历史成功")
}

// GetMessage 获取单个消息
// @Summary 获取消息详情
// @Description 根据消息ID获取消息详情
// @Tags messages
// @Accept json
// @Produce json
// @Param id path int true "消息ID"
// @Success 200 {object} APIResponse{data=entity.Message}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/messages/{id} [get]
func (h *V1) GetMessage(c *fiber.Ctx) error {
	// 解析消息ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.l.Error("消息ID格式错误: %v", err)
		return NewErrorResponse(c, fiber.StatusBadRequest, "消息ID格式错误")
	}

	// 调用业务逻辑获取消息
	message, err := h.messageUC.GetMessage(c.Context(), id)
	if err != nil {
		h.l.Error("获取消息失败: %v", err)
		return NewErrorResponse(c, fiber.StatusInternalServerError, "获取消息失败: "+err.Error())
	}

	return NewSuccessResponse(c, message, "获取消息成功")
}

// RetryFailedMessage 重试失败消息
// @Summary 重试失败消息
// @Description 重新处理失败的消息
// @Tags messages
// @Accept json
// @Produce json
// @Param id path int true "消息ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/v1/messages/{id}/retry [post]
func (h *V1) RetryFailedMessage(c *fiber.Ctx) error {
	// 解析消息ID
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.l.Error("消息ID格式错误: %v", err)
		return NewErrorResponse(c, fiber.StatusBadRequest, "消息ID格式错误")
	}

	// 调用业务逻辑重试消息
	if err := h.messageUC.RetryFailedMessage(c.Context(), id); err != nil {
		h.l.Error("重试消息失败: %v", err)
		return NewErrorResponse(c, fiber.StatusInternalServerError, "重试消息失败: "+err.Error())
	}

	return NewSuccessResponse(c, nil, "消息重试成功")
}
