package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestExampleUsage 展示如何使用Base实体类和相关类型
func TestExampleUsage(t *testing.T) {
	// 1. 创建Bot实例，展示JSONField和Status的使用
	bot := Bot{
		BotName:     "智能客服机器人",
		BotType:     "chatbot",
		Description: "企业智能客服机器人",
		Config: JSONField{
			"max_tokens":    1000,
			"temperature":   0.7,
			"model":         "gpt-3.5-turbo",
			"enable_memory": true,
		},
		Status:    StatusActive,
		CreatedBy: "admin",
	}

	// 操作JSON配置
	bot.Config.Set("api_version", "v1.0")
	bot.Config.Set("timeout", 30)

	// 读取配置
	maxTokens := bot.Config.GetInt("max_tokens")
	model := bot.Config.GetString("model")
	enableMemory := bot.Config.GetBool("enable_memory")
	timeout := bot.Config.GetInt("timeout")

	assert.Equal(t, 1000, maxTokens)
	assert.Equal(t, "gpt-3.5-turbo", model)
	assert.True(t, enableMemory)
	assert.Equal(t, 30, timeout)

	// 检查状态
	assert.True(t, bot.Status.IsActive())
	assert.Equal(t, "active", bot.Status.String())

	// 2. 创建Channel实例，展示枚举类型使用
	channel := Channel{
		BotID:        1,
		PlatformType: "wecom",
		ChannelName:  "企业微信通道",
		WebhookPath:  "/webhook/wecom/123",
		Config: JSONField{
			"corp_id":     "ww123456789",
			"agent_id":    "1000001",
			"secret":      "encrypted_secret",
			"verify_url":  true,
			"webhook_url": "https://api.example.com/webhook",
		},
		ConnectionStatus: ConnectionStatusConnected,
		Status:           StatusActive,
	}

	// 检查连接状态
	assert.Equal(t, "connected", channel.ConnectionStatus.String())
	assert.Equal(t, "wecom", channel.PlatformType)

	// 3. 创建Message实例，展示消息相关枚举
	message := Message{
		MessageID:         "msg_20231201_001",
		ChannelID:         1,
		PlatformMessageID: "wecom_msg_123456",
		Direction:         MessageDirectionInbound,
		MessageType:       "text",
		ContentType:       "text/plain",
		SenderID:          "user_001",
		SenderName:        "张三",
		SenderType:        "user",
		Content:           "你好，请问如何使用这个系统？",
		RawContent: JSONField{
			"msgtype": "text",
			"text": map[string]interface{}{
				"content": "你好，请问如何使用这个系统？",
			},
			"from": map[string]interface{}{
				"userid": "user_001",
				"name":   "张三",
			},
		},
		MessageStatus:     MessageStatusProcessed,
		ConversationID:    "conv_001",
		PlatformTimestamp: time.Now(),
	}

	// 检查消息状态和方向
	assert.Equal(t, "inbound", message.Direction.String())
	assert.Equal(t, "processed", message.MessageStatus.String())

	// 从原始内容中读取信息
	msgType := message.RawContent.GetString("msgtype")
	assert.Equal(t, "text", msgType)

	// 4. 创建队列实例，展示队列状态
	queue := MessageQueue{
		QueueName: "message_processing",
		MessageID: message.MessageID,
		Payload: JSONField{
			"message_id": message.MessageID,
			"action":     "process_inbound",
			"priority":   "normal",
			"metadata": map[string]interface{}{
				"retry_count": 0,
				"created_by":  "system",
			},
		},
		Priority:    100,
		MaxRetries:  3,
		Status:      QueueStatusPending,
		ScheduledAt: time.Now(),
	}

	assert.Equal(t, "pending", queue.Status.String())
	action := queue.Payload.GetString("action")
	assert.Equal(t, "process_inbound", action)

	// 5. 创建路由规则，展示路由类型
	rule := MessageRoutingRule{
		RuleName:        "客服路由规则",
		RuleDescription: "将用户消息路由到客服处理器",
		PlatformTypes: JSONField{
			"platforms": []string{"wecom", "dingtalk"},
		},
		MessageTypes: JSONField{
			"types": []string{"text", "image"},
		},
		ProcessorID:    1,
		RouteType:      RouteTypeDirect,
		Priority:       100,
		IsFallback:     false,
		ConditionLogic: "AND",
		Status:         StatusActive,
	}

	assert.Equal(t, "direct", rule.RouteType.String())
	assert.False(t, rule.IsFallback)

	// 6. 演示Base实体的通用字段
	now := time.Now()
	bot.BaseEntity.CreatedAt = now
	bot.BaseEntity.UpdatedAt = now

	assert.False(t, bot.CreatedAt.IsZero())
	assert.False(t, bot.UpdatedAt.IsZero())
	assert.True(t, bot.DeletedAt.Time.IsZero()) // 软删除字段初始为空
}

// TestJSONFieldAdvancedUsage 展示JSONField的高级用法
func TestJSONFieldAdvancedUsage(t *testing.T) {
	// 创建复杂的JSON配置
	config := JSONField{
		"database": map[string]interface{}{
			"host":     "localhost",
			"port":     3306,
			"username": "root",
			"ssl":      true,
		},
		"redis": map[string]interface{}{
			"host": "localhost",
			"port": 6379,
			"db":   0,
		},
		"features": []interface{}{
			"auto_reply",
			"sentiment_analysis",
			"keyword_detection",
		},
		"limits": map[string]interface{}{
			"max_messages_per_minute": 100,
			"max_file_size_mb":        10,
		},
	}

	// 测试JSON序列化
	jsonStr := config.String()
	assert.Contains(t, jsonStr, "database")
	assert.Contains(t, jsonStr, "redis")

	// 测试类型转换
	dbConfig := config.Get("database")
	assert.NotNil(t, dbConfig)

	features := config.Get("features")
	assert.NotNil(t, features)

	// 测试嵌套访问（需要类型断言）
	if dbMap, ok := dbConfig.(map[string]interface{}); ok {
		host := dbMap["host"].(string)
		// 处理不同的数字类型
		var port int
		switch v := dbMap["port"].(type) {
		case int:
			port = v
		case float64:
			port = int(v)
		}
		ssl := dbMap["ssl"].(bool)

		assert.Equal(t, "localhost", host)
		assert.Equal(t, 3306, port)
		assert.True(t, ssl)
	}
}

// TestEntityRelationships 展示实体关系的使用
func TestEntityRelationships(t *testing.T) {
	// 模拟有关联关系的数据结构
	bot := Bot{
		BaseEntity: BaseEntity{ID: 1},
		BotName:    "测试机器人",
		Status:     StatusActive,
	}

	channel := Channel{
		BaseEntity:       BaseEntity{ID: 1},
		BotID:            bot.ID,
		PlatformType:     "wecom",
		ChannelName:      "测试通道",
		ConnectionStatus: ConnectionStatusConnected,
		Status:           StatusActive,
		Bot:              &bot, // 关联关系
	}

	message := Message{
		BaseEntity:    BaseEntity{ID: 1},
		MessageID:     "msg_001",
		ChannelID:     channel.ID,
		Direction:     MessageDirectionInbound,
		MessageType:   "text",
		Content:       "测试消息",
		MessageStatus: MessageStatusProcessed,
		Channel:       &channel, // 关联关系
	}

	// 验证关联关系
	assert.Equal(t, bot.ID, channel.BotID)
	assert.Equal(t, channel.ID, message.ChannelID)
	assert.Equal(t, "测试机器人", message.Channel.Bot.BotName)
}

// TestStatusTransitions 展示状态转换的使用场景
func TestStatusTransitions(t *testing.T) {
	// 模拟消息状态转换
	message := Message{
		MessageStatus: MessageStatusPending,
	}

	// 开始处理
	message.MessageStatus = MessageStatusProcessing
	assert.Equal(t, "processing", message.MessageStatus.String())

	// 处理完成
	message.MessageStatus = MessageStatusProcessed
	assert.Equal(t, "processed", message.MessageStatus.String())

	// 发送成功
	message.MessageStatus = MessageStatusSent
	assert.Equal(t, "sent", message.MessageStatus.String())

	// 模拟队列状态转换
	queue := MessageQueue{
		Status: QueueStatusPending,
	}

	// 开始执行
	queue.Status = QueueStatusRunning
	assert.Equal(t, "running", queue.Status.String())

	// 执行完成
	queue.Status = QueueStatusCompleted
	assert.Equal(t, "completed", queue.Status.String())
}
