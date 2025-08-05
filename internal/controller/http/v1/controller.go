package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// 类型别名，用于简化引用
type (
	MessageUseCase      = usecase.MessageUseCase
	ChannelUseCase      = usecase.ChannelUseCase
	BotUseCase          = usecase.BotUseCase
	SystemConfigUseCase = usecase.SystemConfigUseCase
	PlatformUseCase     = usecase.PlatformUseCase
	MonitorUseCase      = usecase.MonitorUseCase
	LogUseCase          = usecase.LogUseCase
	QueueUseCase        = usecase.QueueUseCase
	ProcessorUseCase    = usecase.ProcessorUseCase
)

// V1 HTTP控制器结构体
type V1 struct {
	l         logger.Interface
	v         *validator.Validate
	startTime time.Time
	cfg       *config.Config

	// 业务逻辑接口
	messageUC      MessageUseCase
	channelUC      ChannelUseCase
	botUC          BotUseCase
	systemConfigUC SystemConfigUseCase
	platformUC     PlatformUseCase
	monitorUC      MonitorUseCase
	logUC          LogUseCase
	queueUC        QueueUseCase
	processorUC    ProcessorUseCase
}

// NewV1Controller 创建V1控制器实例
func NewV1Controller(
	messageUC MessageUseCase,
	channelUC ChannelUseCase,
	botUC BotUseCase,
	systemConfigUC SystemConfigUseCase,
	platformUC PlatformUseCase,
	monitorUC MonitorUseCase,
	logUC LogUseCase,
	queueUC QueueUseCase,
	processorUC ProcessorUseCase,
	l logger.Interface,
	cfg *config.Config,
) *V1 {
	return &V1{
		messageUC:      messageUC,
		channelUC:      channelUC,
		botUC:          botUC,
		systemConfigUC: systemConfigUC,
		platformUC:     platformUC,
		monitorUC:      monitorUC,
		logUC:          logUC,
		queueUC:        queueUC,
		processorUC:    processorUC,
		l:              l,
		v:              validator.New(),
		startTime:      time.Now(),
		cfg:            cfg,
	}
}

// 消息相关的请求和响应结构体

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ChannelID      string `json:"channel_id" validate:"required"`
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
	ChannelID     *string                  `json:"channel_id,omitempty"`
	SenderID      *string                  `json:"sender_id,omitempty"`
	ReceiverID    *string                  `json:"receiver_id,omitempty"`
	MessageType   *string                  `json:"message_type,omitempty"`
	MessageStatus *entity.MessageStatus    `json:"message_status,omitempty"`
	Direction     *entity.MessageDirection `json:"direction,omitempty"`
	StartTime     *string                  `json:"start_time,omitempty"`
	EndTime       *string                  `json:"end_time,omitempty"`
	Page          int                      `json:"page" validate:"min=1"`
	PageSize      int                      `json:"page_size" validate:"min=1,max=100"`
}

// 工具函数

// parseInt64 解析int64参数
func (v *V1) parseInt64(value string, fieldName string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, &entity.ValidationError{Field: fieldName, Message: "必须是有效的数字"}
	}
	return id, nil
}

// setDefaultPagination 设置默认分页参数
func setDefaultPagination(page, pageSize *int) {
	if *page <= 0 {
		*page = 1
	}
	if *pageSize <= 0 {
		*pageSize = 20
	}
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
// @Summary 发送消息
// @Description 通过指定通道发送消息到平台
// @Tags messages
// @Accept json
// @Produce json
// @Param request body SendMessageRequest true "发送消息请求"
// @Success 200 {object} StandardResponse{data=entity.Message} "发送成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "通道不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/messages/send [post]
func (v *V1) SendMessage(c *fiber.Ctx) error {
	var req SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		v.l.Error("解析请求体失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		v.l.Error("请求参数验证失败: %v", err)
		return NewValidationErrorResponse(c, err)
	}

	// 创建消息实体
	message := &entity.Message{
		ChannelID:      req.ChannelID,
		MessageType:    req.MessageType,
		ReceiverID:     req.ReceiverID,
		ReceiverName:   req.ReceiverName,
		ReceiverType:   req.ReceiverType,
		Content:        req.Content,
		MediaURL:       req.MediaURL,
		MediaType:      req.MediaType,
		ConversationID: req.ConversationID,
		Direction:      entity.MessageDirectionOutbound,
		MessageStatus:  entity.MessageStatusPending,
	}

	// 调用usecase发送消息
	if err := v.messageUC.SendMessage(c.Context(), message); err != nil {
		v.l.Error("发送消息失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "发送消息失败")
	}

	return NewSuccessResponse(c, message)
}

// GetMessageHistory 获取消息历史
// @Summary 获取消息历史
// @Description 根据条件获取消息历史记录，支持分页
// @Tags messages
// @Accept json
// @Produce json
// @Param channel_id query int false "通道ID"
// @Param sender_id query string false "发送者ID"
// @Param receiver_id query string false "接收者ID"
// @Param message_type query string false "消息类型"
// @Param message_status query string false "消息状态"
// @Param direction query string false "消息方向"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认20" default(20)
// @Success 200 {object} StandardResponse{data=usecase.MessageHistoryResult} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/messages/history [get]
func (v *V1) GetMessageHistory(c *fiber.Ctx) error {
	var req GetMessageHistoryRequest
	if err := c.QueryParser(&req); err != nil {
		v.l.Error("解析查询参数失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "查询参数格式错误")
	}

	// 设置默认值
	setDefaultPagination(&req.Page, &req.PageSize)

	// 构建usecase参数
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

	// 调用usecase获取消息历史
	result, err := v.messageUC.GetMessageHistory(c.Context(), params)
	if err != nil {
		v.l.Error("获取消息历史失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "获取消息历史失败")
	}

	return NewSuccessResponse(c, result)
}

// GetMessage 获取消息详情
// @Summary 获取消息详情
// @Description 根据ID获取单个消息的详细信息
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "消息ID"
// @Success 200 {object} StandardResponse{data=entity.Message} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "消息不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/messages/{id} [get]
func (v *V1) GetMessage(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "消息ID不能为空")
	}

	// 调用usecase获取消息
	message, err := v.messageUC.GetMessage(c.Context(), id)
	if err != nil {
		v.l.Error("获取消息失败: %v", err)
		return NewErrorResponse(c, http.StatusNotFound, "消息不存在")
	}

	return NewSuccessResponse(c, message)
}

// RetryMessage 重试失败的消息
// @Summary 重试失败的消息
// @Description 重新发送失败的消息
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "消息ID"
// @Success 200 {object} StandardResponse "重试成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "消息不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/messages/{id}/retry [post]
func (v *V1) RetryMessage(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "消息ID不能为空")
	}

	// 调用usecase重试消息
	if err := v.messageUC.RetryFailedMessage(c.Context(), id); err != nil {
		v.l.Error("重试消息失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "重试消息失败")
	}

	return NewSuccessResponse(c, "重试成功")
}
