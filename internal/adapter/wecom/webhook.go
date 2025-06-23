package wecom

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

// VerifyWebhook 验证Webhook签名
func (w *WecomAdapter) VerifyWebhook(ctx context.Context, signature string, timestamp string, nonce string, body []byte, config map[string]interface{}) error {
	// 实现企业微信的签名验证逻辑
	// 获取配置中的token
	token, ok := config["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("token is required for webhook verification")
	}

	// 企业微信的签名验证算法
	// 1. 将token、timestamp、nonce三个参数进行字典序排序
	// 2. 将三个参数字符串拼接成一个字符串进行sha1哈希
	// 3. 开发者获得加密后的字符串可与signature对比，标识该请求来源于微信
	
	params := []string{token, timestamp, nonce}
	sort.Strings(params)
	
	// 拼接成字符串并进行sha1哈希
	data := strings.Join(params, "")
	h := sha1.New()
	h.Write([]byte(data))
	hash := hex.EncodeToString(h.Sum(nil))
	
	if hash != signature {
		return fmt.Errorf("webhook signature verification failed")
	}

	return nil
}

// ParseInboundMessage 解析入站消息
func (w *WecomAdapter) ParseInboundMessage(ctx context.Context, body []byte, config map[string]interface{}) (*dto.UnifiedMessage, error) {
	// 解析XML数据
	var msg WecomMessage
	if err := xml.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	// 转换为统一消息格式
	unifiedMsg := &dto.UnifiedMessage{
		MessageID:         fmt.Sprintf("wecom_%d", msg.MsgID),
		PlatformMessageID: strconv.FormatInt(msg.MsgID, 10),
		SenderID:          msg.FromUserName,
		ReceiverID:        msg.ToUserName,
		ConversationID:    fmt.Sprintf("wecom_%s_%s", msg.FromUserName, msg.ToUserName),
		PlatformTimestamp: time.Unix(msg.CreateTime, 0),
		RawContent: map[string]interface{}{
			"agent_id": msg.AgentID,
			"msg_type": msg.MsgType,
		},
	}

	// 根据消息类型处理
	switch msg.MsgType {
	case "text":
		unifiedMsg.MessageType = entity.MessageTypeText
		unifiedMsg.Content = msg.Content
		unifiedMsg.SenderType = entity.SenderTypeUser
		unifiedMsg.ReceiverType = entity.ReceiverTypeBot

	case "image":
		unifiedMsg.MessageType = entity.MessageTypeImage
		unifiedMsg.MediaURL = msg.PicURL
		unifiedMsg.MediaType = entity.MediaTypeImage
		unifiedMsg.RawContent["media_id"] = msg.MediaID
		unifiedMsg.SenderType = entity.SenderTypeUser
		unifiedMsg.ReceiverType = entity.ReceiverTypeBot

	case "voice":
		unifiedMsg.MessageType = entity.MessageTypeAudio
		unifiedMsg.MediaURL = msg.MediaID
		unifiedMsg.MediaType = entity.MediaTypeVoice
		unifiedMsg.RawContent["format"] = msg.Format
		unifiedMsg.SenderType = entity.SenderTypeUser
		unifiedMsg.ReceiverType = entity.ReceiverTypeBot

	case "video":
		unifiedMsg.MessageType = entity.MessageTypeVideo
		unifiedMsg.MediaURL = msg.MediaID
		unifiedMsg.MediaType = entity.MediaTypeVideo
		unifiedMsg.RawContent["thumb_media_id"] = msg.ThumbMediaID
		unifiedMsg.SenderType = entity.SenderTypeUser
		unifiedMsg.ReceiverType = entity.ReceiverTypeBot

	case "location":
		unifiedMsg.MessageType = entity.MessageTypeLocation
		unifiedMsg.SenderType = entity.SenderTypeUser
		unifiedMsg.ReceiverType = entity.ReceiverTypeBot

		// 解析位置信息
		latitude, _ := strconv.ParseFloat(msg.LocationX, 64)
		longitude, _ := strconv.ParseFloat(msg.LocationY, 64)
		scale, _ := strconv.ParseFloat(msg.Scale, 64)

		unifiedMsg.LocationContent = &entity.LocationMessage{
			Latitude:  latitude,
			Longitude: longitude,
			Precision: scale,
			Address:   msg.Label,
		}

		// 设置位置信息到RawContent
		unifiedMsg.RawContent["location"] = map[string]interface{}{
			"latitude":  latitude,
			"longitude": longitude,
			"precision": scale,
			"address":   msg.Label,
		}

	case "link":
		unifiedMsg.MessageType = entity.MessageTypeLink
		unifiedMsg.Content = msg.Title
		unifiedMsg.SenderType = entity.SenderTypeUser
		unifiedMsg.ReceiverType = entity.ReceiverTypeBot
		unifiedMsg.RawContent["title"] = msg.Title
		unifiedMsg.RawContent["description"] = msg.Description
		unifiedMsg.RawContent["url"] = msg.URL

	case "event":
		unifiedMsg.MessageType = entity.MessageTypeEvent
		unifiedMsg.SenderType = entity.SenderTypeUser
		unifiedMsg.ReceiverType = entity.ReceiverTypeBot

		// 处理事件消息
		eventType := mapWecomEventType(msg.Event)
		unifiedMsg.EventContent = &entity.EventMessage{
			EventType: eventType,
			EventKey:  msg.EventKey,
			EventData: map[string]interface{}{
				"original_event": msg.Event,
			},
		}

		// 设置事件信息到RawContent
		unifiedMsg.RawContent["event"] = map[string]interface{}{
			"event_type": eventType,
			"event_key":  msg.EventKey,
		}

	default:
		return nil, fmt.Errorf("unsupported message type: %s", msg.MsgType)
	}

	return unifiedMsg, nil
}

// BuildWebhookPath 构建Webhook路径
func (w *WecomAdapter) BuildWebhookPath(channelID int64) string {
	return fmt.Sprintf("/webhook/wecom/%d", channelID)
}

// mapWecomEventType 映射企业微信事件类型到统一事件类型
func mapWecomEventType(wecomEvent string) string {
	switch wecomEvent {
	case "subscribe":
		return entity.EventTypeUserSubscribe
	case "unsubscribe":
		return entity.EventTypeUserUnsubscribe
	case "enter_agent":
		return entity.EventTypeUserEnter
	case "LOCATION":
		return entity.EventTypeLocationReport
	case "click":
		return entity.EventTypeMenuClick
	case "view":
		return entity.EventTypeMenuView
	case "scancode_push", "scancode_waitmsg":
		return entity.EventTypeScanQRCode
	case "location_select":
		return entity.EventTypeLocationSelect
	default:
		return entity.EventTypeCustom
	}
}

// 确保 WecomAdapter 实现了 WebhookProcessor 接口
var _ port.WebhookProcessor = (*WecomAdapter)(nil)
