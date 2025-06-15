package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// 类型别名，用于简化引用
type (
	MessageUseCase = usecase.MessageUseCase
	ChannelUseCase = usecase.ChannelUseCase
	BotUseCase     = usecase.BotUseCase
)

// V1 HTTP控制器结构体
type V1 struct {
	l logger.Interface
	v *validator.Validate

	// 业务逻辑接口
	messageUC MessageUseCase
	channelUC ChannelUseCase
	botUC     BotUseCase
}

// NewV1Controller 创建V1控制器实例
func NewV1Controller(
	messageUC MessageUseCase,
	channelUC ChannelUseCase,
	botUC BotUseCase,
	l logger.Interface,
) *V1 {
	return &V1{
		messageUC: messageUC,
		channelUC: channelUC,
		botUC:     botUC,
		l:         l,
		v:         validator.New(),
	}
}

// 消息相关的请求和响应结构体

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ChannelID      int64  `json:"channel_id" validate:"required"`
	MessageType    string `json:"message_type" validate:"required"`
	ReceiverID     string `json:"receiver_id" validate:"required"`
	ReceiverName   string `json:"receiver_name,omitempty"`
	ReceiverType   string `json:"receiver_type,omitempty"`
	Content        string `json:"content" validate:"required"`
	MediaURL       string `json:"media_url,omitempty"`
	MediaType      string `json:"media_type,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// GetMessageHistoryRequest 获取消息历史请求
type GetMessageHistoryRequest struct {
	ChannelID     *int64  `query:"channel_id"`
	SenderID      *string `query:"sender_id"`
	ReceiverID    *string `query:"receiver_id"`
	MessageType   *string `query:"message_type"`
	MessageStatus *string `query:"message_status"`
	Direction     *string `query:"direction"`
	StartTime     *string `query:"start_time"`
	EndTime       *string `query:"end_time"`
	Page          int     `query:"page" validate:"min=1"`
	PageSize      int     `query:"page_size" validate:"min=1,max=100"`
}

// parseTime 解析时间字符串
func parseTime(timeStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, timeStr)
		if err == nil {
			return t, nil
		}
	}
	return t, err
}

// SendMessage 发送消息
func (v1 *V1) SendMessage(c *fiber.Ctx) error {
	var req SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		v1.l.Error("解析请求体失败", "error", err)
		return NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	if err := v1.v.Struct(&req); err != nil {
		v1.l.Error("请求验证失败", "error", err)
		return NewValidationErrorResponse(c, err)
	}

	// 构建消息实体
	message := &entity.Message{
		ChannelID:      req.ChannelID,
		MessageType:    req.MessageType,
		SenderID:       "system",
		SenderName:     "System",
		SenderType:     "system",
		ReceiverID:     req.ReceiverID,
		ReceiverName:   req.ReceiverName,
		ReceiverType:   req.ReceiverType,
		Content:        req.Content,
		MediaURL:       req.MediaURL,
		MediaType:      req.MediaType,
		ConversationID: req.ConversationID,
	}

	// 发送消息
	if err := v1.messageUC.SendMessage(c.Context(), message); err != nil {
		v1.l.Error("发送消息失败", "error", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "failed to send message")
	}

	v1.l.Info("消息发送成功", "message_id", message.MessageID, "channel_id", message.ChannelID)

	return NewSuccessResponse(c, map[string]interface{}{
		"message_id": message.MessageID,
		"status":     message.MessageStatus.String(),
	})
}

// GetMessageHistory 获取消息历史
func (v1 *V1) GetMessageHistory(c *fiber.Ctx) error {
	var req GetMessageHistoryRequest
	if err := c.QueryParser(&req); err != nil {
		v1.l.Error("解析查询参数失败", "error", err)
		return NewErrorResponse(c, http.StatusBadRequest, "invalid query parameters")
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	if err := v1.v.Struct(&req); err != nil {
		v1.l.Error("请求验证失败", "error", err)
		return NewValidationErrorResponse(c, err)
	}

	// 构建查询参数
	params := usecase.GetMessageHistoryParams{
		ChannelID:   req.ChannelID,
		SenderID:    req.SenderID,
		ReceiverID:  req.ReceiverID,
		MessageType: req.MessageType,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}

	// 处理消息状态
	if req.MessageStatus != nil && *req.MessageStatus != "" {
		var status entity.MessageStatus
		switch *req.MessageStatus {
		case "pending":
			status = entity.MessageStatusPending
		case "processing":
			status = entity.MessageStatusProcessing
		case "processed":
			status = entity.MessageStatusProcessed
		case "sent":
			status = entity.MessageStatusSent
		case "failed":
			status = entity.MessageStatusFailed
		case "expired":
			status = entity.MessageStatusExpired
		default:
			v1.l.Error("无效的消息状态", "status", *req.MessageStatus)
			return NewErrorResponse(c, http.StatusBadRequest, "invalid message status")
		}
		params.MessageStatus = &status
	}

	// 处理消息方向
	if req.Direction != nil && *req.Direction != "" {
		var direction entity.MessageDirection
		switch *req.Direction {
		case "inbound":
			direction = entity.MessageDirectionInbound
		case "outbound":
			direction = entity.MessageDirectionOutbound
		default:
			v1.l.Error("无效的消息方向", "direction", *req.Direction)
			return NewErrorResponse(c, http.StatusBadRequest, "invalid message direction")
		}
		params.Direction = &direction
	}

	// 处理时间参数
	if req.StartTime != nil && *req.StartTime != "" {
		if _, err := parseTime(*req.StartTime); err != nil {
			v1.l.Error("解析开始时间失败", "error", err, "start_time", *req.StartTime)
			return NewErrorResponse(c, http.StatusBadRequest, "invalid start_time format")
		}
		params.StartTime = req.StartTime
	}

	if req.EndTime != nil && *req.EndTime != "" {
		if _, err := parseTime(*req.EndTime); err != nil {
			v1.l.Error("解析结束时间失败", "error", err, "end_time", *req.EndTime)
			return NewErrorResponse(c, http.StatusBadRequest, "invalid end_time format")
		}
		params.EndTime = req.EndTime
	}

	// 查询消息历史
	result, err := v1.messageUC.GetMessageHistory(c.Context(), params)
	if err != nil {
		v1.l.Error("获取消息历史失败", "error", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "failed to get message history")
	}

	v1.l.Info("消息历史查询成功", "total", result.Total, "page", result.Page, "page_size", result.PageSize)

	return NewSuccessResponse(c, result)
}

// GetMessage 获取消息详情
func (v1 *V1) GetMessage(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		v1.l.Error("无效的消息ID", "error", err, "id", idStr)
		return NewErrorResponse(c, http.StatusBadRequest, "invalid message id")
	}

	message, err := v1.messageUC.GetMessage(c.Context(), id)
	if err != nil {
		v1.l.Error("获取消息失败", "error", err, "id", id)
		return NewErrorResponse(c, http.StatusNotFound, "message not found")
	}

	v1.l.Info("消息获取成功", "message_id", message.MessageID, "id", id)

	return NewSuccessResponse(c, message)
}

// RetryFailedMessage 重试失败的消息
func (v1 *V1) RetryFailedMessage(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		v1.l.Error("无效的消息ID", "error", err, "id", idStr)
		return NewErrorResponse(c, http.StatusBadRequest, "invalid message id")
	}

	if err := v1.messageUC.RetryFailedMessage(c.Context(), id); err != nil {
		v1.l.Error("重试消息失败", "error", err, "id", id)
		return NewErrorResponse(c, http.StatusInternalServerError, "failed to retry message")
	}

	v1.l.Info("消息重试成功", "id", id)

	return NewSuccessResponse(c, map[string]interface{}{
		"message": "Message retry initiated",
		"id":      id,
	})
}
