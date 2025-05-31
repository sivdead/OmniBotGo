package v1

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// ProcessWebhook 处理webhook请求
// @Summary 处理平台webhook请求
// @Description 接收并处理来自各个平台的webhook消息
// @Tags webhooks
// @Accept json
// @Produce json
// @Param platform path string true "平台类型" Enums(wecom,dingtalk,wechat,feishu)
// @Param channel_id path string true "通道ID"
// @Param signature header string false "签名"
// @Param timestamp header string false "时间戳"
// @Param nonce header string false "随机数"
// @Param body body object true "Webhook请求体"
// @Success 200 {object} StandardResponse{data=usecase.ProcessWebhookResponse} "处理成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 401 {object} StandardResponse "签名验证失败"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /webhook/{platform}/{channel_id} [post]
func (v *V1) ProcessWebhook(c *fiber.Ctx) error {
	platform := c.Params("platform")
	channelID := c.Params("channel_id")

	if platform == "" || channelID == "" {
		return errorResponse(c, http.StatusBadRequest, "平台类型和通道ID不能为空")
	}

	// 获取请求头
	signature := c.Get("X-Signature", "")
	timestamp := c.Get("X-Timestamp", "")
	nonce := c.Get("X-Nonce", "")

	// 获取原始请求体
	body := c.Body()

	req := &usecase.ProcessWebhookRequest{
		Platform:  platform,
		ChannelID: channelID,
		Signature: signature,
		Timestamp: timestamp,
		Nonce:     nonce,
		Body:      body,
	}

	result, err := v.webhookUC.HandleWebhook(c.Context(), req)
	if err != nil {
		v.l.Error("处理webhook失败: %v", err)
		if err.Error() == "签名验证失败" {
			return errorResponse(c, http.StatusUnauthorized, err.Error())
		}
		return errorResponse(c, http.StatusInternalServerError, "处理webhook失败")
	}

	v.l.Info("webhook处理成功: platform=%s, channel=%s, message=%s",
		platform, channelID, result.MessageID)

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"success": true,
		"message": "处理成功",
		"data":    result,
	})
}

// VerifyWebhook 验证webhook配置
// @Summary 验证webhook配置
// @Description 用于平台验证webhook URL的有效性
// @Tags webhooks
// @Accept json
// @Produce json
// @Param platform path string true "平台类型" Enums(wecom,dingtalk,wechat,feishu)
// @Param channel_id path string true "通道ID"
// @Param echostr query string false "回显字符串（微信公众号）"
// @Param challenge query string false "验证字符串（其他平台）"
// @Success 200 {string} string "验证成功"
// @Failure 400 {object} StandardResponse "请求参数错误"
// @Failure 500 {object} StandardResponse "内部服务器错误"
// @Router /webhook/{platform}/{channel_id} [get]
func (v *V1) VerifyWebhook(c *fiber.Ctx) error {
	platform := c.Params("platform")
	channelID := c.Params("channel_id")

	if platform == "" || channelID == "" {
		return errorResponse(c, http.StatusBadRequest, "平台类型和通道ID不能为空")
	}

	// 获取验证参数
	echostr := c.Query("echostr")
	challenge := c.Query("challenge")

	v.l.Info("webhook验证请求: platform=%s, channel=%s", platform, channelID)

	// 根据平台返回相应的验证响应
	switch platform {
	case "wechat":
		// 微信公众号返回echostr
		if echostr != "" {
			return c.SendString(echostr)
		}
	case "dingtalk", "feishu":
		// 钉钉、飞书返回challenge
		if challenge != "" {
			return c.JSON(map[string]string{
				"challenge": challenge,
			})
		}
	case "wecom":
		// 企业微信需要解密返回
		// 这里简化处理，实际需要按照企业微信文档解密
		if echostr != "" {
			return c.SendString(echostr)
		}
	}

	return errorResponse(c, http.StatusBadRequest, "缺少验证参数")
}
