package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/entity"
)

type webhookUseCase struct {
	messageUC MessageUseCase
	channelUC ChannelUseCase
	logger    zerolog.Logger
}

func NewWebhookUseCase(
	messageUC MessageUseCase,
	channelUC ChannelUseCase,
	logger zerolog.Logger,
) WebhookUseCase {
	return &webhookUseCase{
		messageUC: messageUC,
		channelUC: channelUC,
		logger:    logger.With().Str("component", "webhook_usecase").Logger(),
	}
}

func (uc *webhookUseCase) HandleWebhook(ctx context.Context, req *ProcessWebhookRequest) (*ProcessWebhookResponse, error) {
	logger := uc.logger.With().
		Str("platform", req.Platform).
		Str("channel_id", req.ChannelID).
		Logger()

	logger.Info().Msg("开始处理webhook请求")

	// 获取通道信息
	channel, err := uc.channelUC.GetChannelByID(ctx, &GetChannelByIDRequest{
		ID: req.ChannelID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("获取通道信息失败")
		return nil, fmt.Errorf("获取通道信息失败: %w", err)
	}

	if channel.Channel.Status != entity.StatusActive {
		return nil, errors.New("通道未激活")
	}

	// 验证签名
	if err := uc.verifySignature(req.Platform, req.Signature, req.Timestamp, req.Nonce, req.Body, channel.Channel.Config.Secret); err != nil {
		logger.Error().Err(err).Msg("签名验证失败")
		return nil, fmt.Errorf("签名验证失败: %w", err)
	}

	// 解析消息
	unifiedMessage, err := uc.parseMessage(req.Platform, req.Body, channel.Channel.Config)
	if err != nil {
		logger.Error().Err(err).Msg("消息解析失败")
		return nil, fmt.Errorf("消息解析失败: %w", err)
	}

	if unifiedMessage == nil {
		logger.Debug().Msg("收到非消息事件，跳过处理")
		return &ProcessWebhookResponse{
			Success: true,
			Message: "事件已忽略",
		}, nil
	}

	// 处理入站消息
	processReq := &ProcessInboundMessageRequest{
		PlatformID:     req.Platform,
		ChannelID:      req.ChannelID,
		FromUserID:     unifiedMessage.FromUserID,
		ToUserID:       unifiedMessage.ToUserID,
		MessageType:    unifiedMessage.MessageType,
		Content:        unifiedMessage.Content,
		OriginalData:   req.Body,
		ReceivedAt:     time.Now(),
		ConversationID: unifiedMessage.ConversationID,
		ReplyToMsgID:   unifiedMessage.ReplyToMsgID,
	}

	result, err := uc.messageUC.ProcessInboundMessage(ctx, processReq)
	if err != nil {
		logger.Error().Err(err).Msg("处理入站消息失败")
		return nil, fmt.Errorf("处理入站消息失败: %w", err)
	}

	logger.Info().
		Str("message_id", result.MessageID).
		Msg("webhook处理完成")

	return &ProcessWebhookResponse{
		Success:   true,
		MessageID: result.MessageID,
		Message:   "消息处理成功",
	}, nil
}

func (uc *webhookUseCase) verifySignature(platform, signature, timestamp, nonce, body, secret string) error {
	switch platform {
	case "wecom":
		return uc.verifyWeComSignature(signature, timestamp, nonce, body, secret)
	case "dingtalk":
		return uc.verifyDingTalkSignature(signature, timestamp, secret)
	case "wechat":
		return uc.verifyWeChatSignature(signature, timestamp, nonce, body, secret)
	case "feishu":
		return uc.verifyFeishuSignature(signature, timestamp, nonce, body, secret)
	default:
		return errors.New("不支持的平台")
	}
}

func (uc *webhookUseCase) verifyWeComSignature(signature, timestamp, nonce, body, token string) error {
	// 企业微信签名验证算法
	// signature = sha1(sort(token, timestamp, nonce, encrypt))
	if token == "" {
		return errors.New("企业微信token未配置")
	}

	// 这里简化处理，实际应该按照企业微信文档进行完整验证
	// 包括消息体解密等步骤
	expected := fmt.Sprintf("%s%s%s%s", token, timestamp, nonce, body)
	hash := sha256.Sum256([]byte(expected))
	expectedSig := hex.EncodeToString(hash[:])

	if signature != expectedSig {
		return errors.New("企业微信签名验证失败")
	}

	return nil
}

func (uc *webhookUseCase) verifyDingTalkSignature(signature, timestamp, secret string) error {
	// 钉钉签名验证算法
	// signature = base64(hmac_sha256(timestamp + "\n" + secret, secret))
	if secret == "" {
		return errors.New("钉钉secret未配置")
	}

	// 检查时间戳（防重放攻击）
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的时间戳: %w", err)
	}

	now := time.Now().Unix() * 1000
	if now-ts > 3600000 { // 1小时过期
		return errors.New("请求已过期")
	}

	// 这里简化处理，实际应该使用HMAC-SHA256验证
	expected := timestamp + secret
	hash := sha256.Sum256([]byte(expected))
	expectedSig := hex.EncodeToString(hash[:])

	if signature != expectedSig {
		return errors.New("钉钉签名验证失败")
	}

	return nil
}

func (uc *webhookUseCase) verifyWeChatSignature(signature, timestamp, nonce, body, token string) error {
	// 微信公众号签名验证算法
	// signature = sha1(sort(token, timestamp, nonce))
	if token == "" {
		return errors.New("微信token未配置")
	}

	expected := fmt.Sprintf("%s%s%s%s", token, timestamp, nonce, body)
	hash := sha256.Sum256([]byte(expected))
	expectedSig := hex.EncodeToString(hash[:])

	if signature != expectedSig {
		return errors.New("微信签名验证失败")
	}

	return nil
}

func (uc *webhookUseCase) verifyFeishuSignature(signature, timestamp, nonce, body, secret string) error {
	// 飞书签名验证算法
	if secret == "" {
		return errors.New("飞书secret未配置")
	}

	expected := timestamp + nonce + secret + body
	hash := sha256.Sum256([]byte(expected))
	expectedSig := hex.EncodeToString(hash[:])

	if signature != expectedSig {
		return errors.New("飞书签名验证失败")
	}

	return nil
}

type UnifiedMessage struct {
	FromUserID     string
	ToUserID       string
	MessageType    string
	Content        string
	ConversationID string
	ReplyToMsgID   string
}

func (uc *webhookUseCase) parseMessage(platform string, body []byte, config entity.ChannelConfig) (*UnifiedMessage, error) {
	switch platform {
	case "wecom":
		return uc.parseWeComMessage(body, config)
	case "dingtalk":
		return uc.parseDingTalkMessage(body, config)
	case "wechat":
		return uc.parseWeChatMessage(body, config)
	case "feishu":
		return uc.parseFeishuMessage(body, config)
	default:
		return nil, errors.New("不支持的平台")
	}
}

func (uc *webhookUseCase) parseWeComMessage(body []byte, config entity.ChannelConfig) (*UnifiedMessage, error) {
	// 解析企业微信消息格式
	// 这里简化处理，实际需要解析XML格式的企业微信消息

	// 示例解析逻辑（实际应该根据企业微信文档解析XML）
	return &UnifiedMessage{
		FromUserID:     "sample_user",
		ToUserID:       config.AppID,
		MessageType:    "text",
		Content:        string(body),
		ConversationID: "conv_" + strconv.FormatInt(time.Now().Unix(), 10),
	}, nil
}

func (uc *webhookUseCase) parseDingTalkMessage(body []byte, config entity.ChannelConfig) (*UnifiedMessage, error) {
	// 解析钉钉消息格式
	// 这里简化处理，实际需要解析JSON格式的钉钉消息

	return &UnifiedMessage{
		FromUserID:     "sample_user",
		ToUserID:       config.AppID,
		MessageType:    "text",
		Content:        string(body),
		ConversationID: "conv_" + strconv.FormatInt(time.Now().Unix(), 10),
	}, nil
}

func (uc *webhookUseCase) parseWeChatMessage(body []byte, config entity.ChannelConfig) (*UnifiedMessage, error) {
	// 解析微信公众号消息格式
	// 这里简化处理，实际需要解析XML格式的微信消息

	return &UnifiedMessage{
		FromUserID:     "sample_user",
		ToUserID:       config.AppID,
		MessageType:    "text",
		Content:        string(body),
		ConversationID: "conv_" + strconv.FormatInt(time.Now().Unix(), 10),
	}, nil
}

func (uc *webhookUseCase) parseFeishuMessage(body []byte, config entity.ChannelConfig) (*UnifiedMessage, error) {
	// 解析飞书消息格式
	// 这里简化处理，实际需要解析JSON格式的飞书消息

	return &UnifiedMessage{
		FromUserID:     "sample_user",
		ToUserID:       config.AppID,
		MessageType:    "text",
		Content:        string(body),
		ConversationID: "conv_" + strconv.FormatInt(time.Now().Unix(), 10),
	}, nil
}
