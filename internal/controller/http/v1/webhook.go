package v1

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/adapter"
	"github.com/sivdead/OmniBotGo/internal/entity"
)

// WebhookController Webhook控制器
type WebhookController struct {
	*V1
	adapterManager *adapter.Manager
}

// NewWebhookController 创建Webhook控制器实例
func NewWebhookController(v1 *V1, adapterManager *adapter.Manager) *WebhookController {
	return &WebhookController{
		V1:             v1,
		adapterManager: adapterManager,
	}
}

// HandleWebhook 处理Webhook回调
// @Summary 处理平台Webhook回调
// @Description 接收并处理来自各个平台的消息回调
// @Tags webhook
// @Accept json
// @Produce json
// @Param platform path string true "平台类型" Enums(wecom,dingtalk,wechat_official,feishu)
// @Param channel_id path int true "通道ID"
// @Param signature query string false "签名"
// @Param timestamp query string false "时间戳"
// @Param nonce query string false "随机数"
// @Param echostr query string false "验证字符串"
// @Success 200 {object} StandardResponse "处理成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "通道不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /webhook/{platform}/{channel_id} [post]
func (w *WebhookController) HandleWebhook(c *fiber.Ctx) error {
	platformStr := c.Params("platform")
	channelIDStr := c.Params("channel_id")

	// 验证平台类型
	platformType := entity.PlatformType(platformStr)
	_, err := w.adapterManager.GetAdapter(platformType)
	if err != nil {
		w.l.Error("不支持的平台类型: %s", platformStr)
		return NewErrorResponse(c, http.StatusBadRequest, "不支持的平台类型")
	}

	// 验证通道ID
	channelID := channelIDStr
	if channelID == "" {
		w.l.Error("通道ID不能为空")
		return NewErrorResponse(c, http.StatusBadRequest, "通道ID不能为空")
	}

	// 获取通道信息
	channel, err := w.channelUC.GetChannel(c.Context(), channelID)
	if err != nil {
		w.l.Error("获取通道信息失败: %v", err)
		return NewErrorResponse(c, http.StatusNotFound, "通道不存在")
	}

	// 验证平台类型匹配
	if channel.PlatformType != platformStr {
		w.l.Error("通道平台类型不匹配: 期望 %s, 实际 %s", channel.PlatformType, platformStr)
		return NewErrorResponse(c, http.StatusBadRequest, "通道平台类型不匹配")
	}

	// 获取查询参数
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echoStr := c.Query("echostr")

	// 企业微信特殊处理
	if platformType == entity.PlatformTypeWecom {
		signature = c.Query("msg_signature")
	}

	// 读取请求体
	body := c.Body()

	// 如果是验证URL的请求（GET请求或有echostr参数）
	if c.Method() == "GET" || echoStr != "" {
		return w.handleURLVerification(c, platformType, signature, timestamp, nonce, echoStr, channel.Config)
	}

	// 验证Webhook签名
	err = w.adapterManager.VerifyWebhook(c.Context(), platformType, signature, timestamp, nonce, body, channel.Config)
	if err != nil {
		w.l.Error("Webhook签名验证失败: %v", err)
		return NewErrorResponse(c, http.StatusUnauthorized, "签名验证失败")
	}

	// 解析入站消息
	unifiedMessage, err := w.adapterManager.ParseInboundMessage(c.Context(), platformType, body, channel.Config)
	if err != nil {
		w.l.Error("解析入站消息失败: %v", err)
		return NewErrorResponse(c, http.StatusBadRequest, "消息解析失败")
	}

	// 转换为内部消息格式
	message := &entity.Message{
		MessageID:         unifiedMessage.MessageID,
		ChannelID:         channelID,
		PlatformMessageID: unifiedMessage.PlatformMessageID,
		Direction:         entity.MessageDirectionInbound,
		MessageType:       unifiedMessage.MessageType,
		SenderID:          unifiedMessage.SenderID,
		SenderName:        unifiedMessage.SenderName,
		SenderType:        unifiedMessage.SenderType,
		ReceiverID:        unifiedMessage.ReceiverID,
		ReceiverName:      unifiedMessage.ReceiverName,
		ReceiverType:      unifiedMessage.ReceiverType,
		Content:           unifiedMessage.Content,
		RawContent:        entity.JSONField(unifiedMessage.RawContent),
		UnifiedContent:    entity.JSONField(unifiedMessage.RawContent),
		MediaURL:          unifiedMessage.MediaURL,
		MediaType:         unifiedMessage.MediaType,
		FileSize:          unifiedMessage.FileSize,
		ConversationID:    unifiedMessage.ConversationID,
		PlatformTimestamp: unifiedMessage.PlatformTimestamp,
	}

	// 处理入站消息
	err = w.messageUC.ProcessInboundMessage(c.Context(), message)
	if err != nil {
		w.l.Error("处理入站消息失败: %v", err)
		return NewErrorResponse(c, http.StatusInternalServerError, "消息处理失败")
	}

	w.l.Info("成功处理Webhook消息: 平台=%s, 通道=%d, 消息ID=%s", platformStr, channelID, unifiedMessage.MessageID)

	return NewSuccessResponse(c, http.StatusOK, "消息处理成功", nil)
}

// handleURLVerification 处理URL验证
func (w *WebhookController) handleURLVerification(c *fiber.Ctx, platformType entity.PlatformType, signature, timestamp, nonce, echoStr string, config entity.JSONField) error {
	// 验证签名
	err := w.adapterManager.VerifyWebhook(c.Context(), platformType, signature, timestamp, nonce, []byte{}, config)
	if err != nil {
		w.l.Error("URL验证失败: %v", err)
		return NewErrorResponse(c, http.StatusUnauthorized, "URL验证失败")
	}

	// 返回echostr
	return c.SendString(echoStr)
}

// GetWebhookInfo 获取Webhook信息
// @Summary 获取Webhook配置信息
// @Description 获取指定通道的Webhook配置信息
// @Tags webhook
// @Accept json
// @Produce json
// @Param platform path string true "平台类型"
// @Param channel_id path int true "通道ID"
// @Success 200 {object} StandardResponse{data=WebhookInfo} "获取成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 404 {object} StandardResponse "通道不存在"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /webhook/{platform}/{channel_id} [get]
func (w *WebhookController) GetWebhookInfo(c *fiber.Ctx) error {
	platformStr := c.Params("platform")
	channelIDStr := c.Params("channel_id")

	// 验证平台类型
	platformType := entity.PlatformType(platformStr)
	_, err := w.adapterManager.GetAdapter(platformType)
	if err != nil {
		return NewErrorResponse(c, http.StatusBadRequest, "不支持的平台类型")
	}

	// 验证通道ID
	channelID := channelIDStr
	if channelID == "" {
		return NewErrorResponse(c, http.StatusBadRequest, "通道ID不能为空")
	}

	// 获取通道信息
	channel, err := w.channelUC.GetChannel(c.Context(), channelID)
	if err != nil {
		return NewErrorResponse(c, http.StatusNotFound, "通道不存在")
	}

	// 构建Webhook路径 (临时转换，待后续重构接口)
	// TODO: 更新 BuildWebhookPath 接口以支持 string channelID
	channelIDInt, _ := strconv.ParseInt(channelID, 10, 64)
	webhookPath, err := w.adapterManager.BuildWebhookPath(platformType, channelIDInt)
	if err != nil {
		return NewErrorResponse(c, http.StatusInternalServerError, "构建Webhook路径失败")
	}

	webhookInfo := WebhookInfo{
		ChannelID:        channelID,
		Platform:         platformStr,
		WebhookPath:      webhookPath,
		WebhookURL:       c.Protocol() + "://" + c.Get("Host") + webhookPath,
		ChannelName:      channel.ChannelName,
		Status:           channel.Status.String(),
		ConnectionStatus: channel.ConnectionStatus.String(),
	}

	return NewSuccessResponse(c, http.StatusOK, "获取成功", webhookInfo)
}

// WebhookInfo Webhook信息
type WebhookInfo struct {
	ChannelID        string `json:"channel_id"`
	Platform         string `json:"platform"`
	WebhookPath      string `json:"webhook_path"`
	WebhookURL       string `json:"webhook_url"`
	ChannelName      string `json:"channel_name"`
	Status           string `json:"status"`
	ConnectionStatus string `json:"connection_status"`
}
