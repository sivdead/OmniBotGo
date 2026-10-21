package v1

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// PreviewRoutingRequest is a synthetic inbound-shaped message for dry-run routing (T05).
type PreviewRoutingRequest struct {
	ChannelID   string `json:"channel_id" validate:"required"`
	MessageType string `json:"message_type" validate:"required"`
	Content     string `json:"content"`
	SenderID    string `json:"sender_id"`
	MessageID   string `json:"message_id,omitempty"` // optional: load stored inbound instead
}

// PreviewRoutingResponse is the safe preview payload (no processor execution).
type PreviewRoutingResponse struct {
	MatchedRules []PreviewRuleSummary      `json:"matched_rules"`
	Selected     []PreviewProcessorSummary `json:"selected_processors"`
	UsedFallback bool                      `json:"used_fallback"`
}

// PreviewRuleSummary is a redacted rule view for preview.
type PreviewRuleSummary struct {
	ID          string           `json:"id"`
	RuleName    string           `json:"rule_name"`
	Priority    int              `json:"priority"`
	IsFallback  bool             `json:"is_fallback"`
	RouteType   entity.RouteType `json:"route_type"`
	ProcessorID string           `json:"processor_id"`
}

// PreviewProcessorSummary is a redacted processor view (config secrets omitted).
type PreviewProcessorSummary struct {
	ID            string `json:"id"`
	ProcessorName string `json:"processor_name"`
	ProcessorType string `json:"processor_type"`
	Priority      int    `json:"priority"`
}

// PreviewRouting 预览路由匹配结果（不执行处理器、不发送）（T05）
// @Summary 预览消息路由
// @Description 返回匹配规则、选中处理器与是否走兜底；不调用处理器、不向平台发送
// @Tags routing
// @Accept json
// @Produce json
// @Param request body PreviewRoutingRequest true "预览请求"
// @Success 200 {object} StandardResponse{data=PreviewRoutingResponse} "预览成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /api/v1/routing/preview [post]
func (v *V1) PreviewRouting(c *fiber.Ctx) error {
	var req PreviewRoutingRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, http.StatusBadRequest, "请求参数格式错误")
	}

	if err := v.v.Struct(&req); err != nil {
		return NewValidationErrorResponse(c, err)
	}

	msg := &entity.Message{
		ChannelID:   req.ChannelID,
		MessageType: req.MessageType,
		Content:     req.Content,
		SenderID:    req.SenderID,
		Direction:   entity.MessageDirectionInbound,
		MessageID:   req.MessageID,
	}

	if req.MessageID != "" {
		stored, err := v.messageUC.GetMessage(c.Context(), req.MessageID)
		if err == nil && stored != nil {
			msg = stored
		}
	}

	preview, err := v.routingUC.PreviewRoute(c.Context(), msg)
	if err != nil {
		if errors.Is(err, usecase.ErrNoMatchingRoutingRules) {
			return NewErrorResponse(c, http.StatusNotFound, err.Error())
		}

		v.l.Error("路由预览失败: %v", err)

		return NewErrorResponse(c, http.StatusInternalServerError, "路由预览失败")
	}

	resp := PreviewRoutingResponse{
		UsedFallback: preview.UsedFallback,
		MatchedRules: make([]PreviewRuleSummary, 0, len(preview.MatchedRules)),
		Selected:     make([]PreviewProcessorSummary, 0, len(preview.SelectedProcessors)),
	}

	for _, rule := range preview.MatchedRules {
		resp.MatchedRules = append(resp.MatchedRules, PreviewRuleSummary{
			ID:          rule.ID,
			RuleName:    rule.RuleName,
			Priority:    rule.Priority,
			IsFallback:  rule.IsFallback,
			RouteType:   rule.RouteType,
			ProcessorID: rule.ProcessorID,
		})
	}

	for _, p := range preview.SelectedProcessors {
		resp.Selected = append(resp.Selected, PreviewProcessorSummary{
			ID:            p.ID,
			ProcessorName: p.ProcessorName,
			ProcessorType: p.ProcessorType,
			Priority:      p.Priority,
		})
	}

	return NewSuccessResponse(c, resp)
}
