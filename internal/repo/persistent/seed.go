package persistent

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/pkg/database"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// DataSeeder 数据初始化器
type DataSeeder struct {
	db     database.CommonDB
	logger logger.Interface
}

// NewDataSeeder 创建数据初始化器
func NewDataSeeder(db database.CommonDB, logger logger.Interface) *DataSeeder {
	return &DataSeeder{
		db:     db,
		logger: logger,
	}
}

// SeedAll 初始化所有基础数据
func (s *DataSeeder) SeedAll(ctx context.Context) error {
	s.logger.Info("开始初始化数据库种子数据...")

	if err := s.SeedBots(ctx); err != nil {
		return fmt.Errorf("初始化Bot数据失败: %w", err)
	}

	if err := s.SeedSystemConfigs(ctx); err != nil {
		return fmt.Errorf("初始化系统配置失败: %w", err)
	}

	if err := s.SeedMessageProcessors(ctx); err != nil {
		return fmt.Errorf("初始化消息处理器失败: %w", err)
	}

	s.logger.Info("数据库种子数据初始化完成")
	return nil
}

// SeedBots 初始化默认Bot数据
func (s *DataSeeder) SeedBots(ctx context.Context) error {
	bots := []*entity.Bot{
		{
			BotName:     "默认客服Bot",
			BotType:     "standard",
			Description: "默认的客服机器人，处理常见问答",
			Config:      entity.NewJSONField(`{"auto_reply": true, "fallback_message": "您好，我是智能客服，请问有什么可以帮助您的？"}`),
			Status:      entity.StatusActive,
			CreatedBy:   "system",
		},
		{
			BotName:     "AI助手Bot",
			BotType:     "ai",
			Description: "基于AI的智能助手，支持复杂对话",
			Config:      entity.NewJSONField(`{"model": "gpt-3.5-turbo", "temperature": 0.7, "max_tokens": 1000}`),
			Status:      entity.StatusActive,
			CreatedBy:   "system",
		},
	}

	for _, bot := range bots {
		// 检查是否已存在
		var count int64
		err := s.db.GetGORM().WithContext(ctx).Model(&entity.Bot{}).
			Where("bot_name = ?", bot.BotName).Count(&count).Error
		if err != nil {
			return fmt.Errorf("检查Bot是否存在失败: %w", err)
		}

		if count == 0 {
			err = s.db.GetGORM().WithContext(ctx).Create(bot).Error
			if err != nil {
				return fmt.Errorf("创建Bot失败: %w", err)
			}
			s.logger.Info("已创建Bot: %s", bot.BotName)
		} else {
			s.logger.Info("Bot已存在，跳过: %s", bot.BotName)
		}
	}

	return nil
}

// SeedSystemConfigs 初始化系统配置
func (s *DataSeeder) SeedSystemConfigs(ctx context.Context) error {
	configs := []*entity.SystemConfig{
		{
			ConfigKey:   "server.port",
			ConfigValue: "8080",
			ConfigType:  "int",
			ConfigGroup: "server",
			Description: "服务器端口",
			IsSystem:    true,
		},
		{
			ConfigKey:   "server.host",
			ConfigValue: "0.0.0.0",
			ConfigType:  "string",
			ConfigGroup: "server",
			Description: "服务器绑定地址",
			IsSystem:    true,
		},
		{
			ConfigKey:   "log.level",
			ConfigValue: "info",
			ConfigType:  "string",
			ConfigGroup: "logging",
			Description: "日志级别",
			IsSystem:    false,
		},
		{
			ConfigKey:   "message.max_size",
			ConfigValue: "10485760",
			ConfigType:  "int",
			ConfigGroup: "message",
			Description: "最大消息大小(字节)",
			IsSystem:    false,
		},
		{
			ConfigKey:   "webhook.timeout",
			ConfigValue: "30",
			ConfigType:  "int",
			ConfigGroup: "webhook",
			Description: "Webhook超时时间(秒)",
			IsSystem:    false,
		},
		{
			ConfigKey:   "token.default_expire",
			ConfigValue: "7200",
			ConfigType:  "int",
			ConfigGroup: "auth",
			Description: "默认Token过期时间(秒)",
			IsSystem:    false,
		},
		{
			ConfigKey:   "message.retention_days",
			ConfigValue: "30",
			ConfigType:  "int",
			ConfigGroup: "message",
			Description: "消息保留天数",
			IsSystem:    false,
		},
		{
			ConfigKey:   "queue.max_size",
			ConfigValue: "10000",
			ConfigType:  "int",
			ConfigGroup: "queue",
			Description: "队列最大长度",
			IsSystem:    false,
		},
		{
			ConfigKey:   "webhook.base_url",
			ConfigValue: "http://localhost:8080",
			ConfigType:  "string",
			ConfigGroup: "webhook",
			Description: "Webhook基础URL",
			IsSystem:    false,
		},
		{
			ConfigKey:   "channel.webhook_path_template",
			ConfigValue: "/webhook/channel-{id}",
			ConfigType:  "string",
			ConfigGroup: "channel",
			Description: "通道Webhook路径模板",
			IsSystem:    true,
		},
	}

	for _, config := range configs {
		// 检查是否已存在
		var count int64
		err := s.db.GetGORM().WithContext(ctx).Model(&entity.SystemConfig{}).
			Where("config_key = ?", config.ConfigKey).Count(&count).Error
		if err != nil {
			return fmt.Errorf("检查系统配置是否存在失败: %w", err)
		}

		if count == 0 {
			err = s.db.GetGORM().WithContext(ctx).Create(config).Error
			if err != nil {
				return fmt.Errorf("创建系统配置失败: %w", err)
			}
			s.logger.Info("已创建系统配置: %s", config.ConfigKey)
		} else {
			s.logger.Info("系统配置已存在，跳过: %s", config.ConfigKey)
		}
	}

	return nil
}

// SeedMessageProcessors 初始化消息处理器
func (s *DataSeeder) SeedMessageProcessors(ctx context.Context) error {
	processors := []*entity.MessageProcessor{
		{
			ProcessorName: "OpenAI聊天处理器",
			ProcessorType: "openai_chat",
			Config:        entity.NewJSONField(`{"api_key": "sk-xxxxx", "model": "gpt-3.5-turbo", "temperature": 0.7, "base_url": "https://api.openai.com"}`),
			Priority:      100,
			Status:        entity.StatusActive,
		},
		{
			ProcessorName: "客服工单系统",
			ProcessorType: "webhook_forwarder",
			Config:        entity.NewJSONField(`{"webhook_url": "https://crm.company.com/api/webhook", "api_key": "crm_api_key", "timeout": 30}`),
			Priority:      200,
			Status:        entity.StatusActive,
		},
		{
			ProcessorName: "订单查询处理器",
			ProcessorType: "order_query",
			Config:        entity.NewJSONField(`{"api_base_url": "https://order.company.com/api", "api_key": "order_api_key"}`),
			Priority:      300,
			Status:        entity.StatusActive,
		},
		{
			ProcessorName: "邮件通知处理器",
			ProcessorType: "email_notifier",
			Config:        entity.NewJSONField(`{"smtp_host": "mail.company.com", "smtp_user": "bot@company.com", "smtp_pass": "xxxxx"}`),
			Priority:      400,
			Status:        entity.StatusActive,
		},
	}

	for _, processor := range processors {
		// 检查是否已存在
		var count int64
		err := s.db.GetGORM().WithContext(ctx).Model(&entity.MessageProcessor{}).
			Where("processor_name = ?", processor.ProcessorName).Count(&count).Error
		if err != nil {
			return fmt.Errorf("检查消息处理器是否存在失败: %w", err)
		}

		if count == 0 {
			err = s.db.GetGORM().WithContext(ctx).Create(processor).Error
			if err != nil {
				return fmt.Errorf("创建消息处理器失败: %w", err)
			}
			s.logger.Info("已创建消息处理器: %s", processor.ProcessorName)
		} else {
			s.logger.Info("消息处理器已存在，跳过: %s", processor.ProcessorName)
		}
	}

	return nil
}

// ClearAll 清空所有数据（仅用于测试环境）
func (s *DataSeeder) ClearAll(ctx context.Context) error {
	s.logger.Warn("开始清空所有数据...")

	// 按照依赖关系逆序删除
	tables := []interface{}{
		&entity.APICallLog{},
		&entity.ConnectionLog{},
		&entity.MessageQueue{},
		&entity.MessageRoutingRule{},
		&entity.MessageProcessor{},
		&entity.SystemConfig{},
		&entity.Message{},
		&entity.Channel{},
		&entity.Bot{},
	}

	return s.db.GetGORM().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, table := range tables {
			// 使用Unscoped进行硬删除
			if err := tx.Unscoped().Where("1 = 1").Delete(table).Error; err != nil {
				return fmt.Errorf("删除表数据失败: %w", err)
			}
		}
		s.logger.Warn("所有数据已清空")
		return nil
	})
}

// ResetAutoIncrement 重置自增ID（仅用于测试环境）
func (s *DataSeeder) ResetAutoIncrement(ctx context.Context) error {
	s.logger.Warn("重置自增ID...")

	tables := []string{
		"bots", "channels", "messages", "message_processors",
		"message_routing_rules", "system_configs", "message_queue",
		"connection_logs", "api_call_logs",
	}

	return s.db.GetGORM().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, table := range tables {
			sql := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = 1", table)
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("重置表 %s 自增ID失败: %w", table, err)
			}
		}
		s.logger.Warn("所有表自增ID已重置")
		return nil
	})
}
