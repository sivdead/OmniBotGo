package usecase

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// RoutingUseCase 消息路由业务逻辑接口
type RoutingUseCase interface {
	// RouteMessage 根据路由规则处理消息
	RouteMessage(ctx context.Context, message *entity.Message) ([]*entity.MessageProcessor, error)
	// GetMatchingRules 获取匹配的路由规则
	GetMatchingRules(ctx context.Context, message *entity.Message) ([]*entity.MessageRoutingRule, error)
	// CreateRoutingRule 创建路由规则
	CreateRoutingRule(ctx context.Context, rule *entity.MessageRoutingRule) error
	// UpdateRoutingRule 更新路由规则
	UpdateRoutingRule(ctx context.Context, rule *entity.MessageRoutingRule) error
	// DeleteRoutingRule 删除路由规则
	DeleteRoutingRule(ctx context.Context, ruleID string) error
	// GetRoutingRules 获取路由规则列表
	GetRoutingRules(ctx context.Context, filters RoutingRuleFilters) ([]*entity.MessageRoutingRule, error)
}

// RoutingRuleFilters 路由规则查询过滤器
type RoutingRuleFilters struct {
	ProcessorID  *int64         `json:"processor_id,omitempty"`
	PlatformType *string        `json:"platform_type,omitempty"`
	MessageType  *string        `json:"message_type,omitempty"`
	Status       *entity.Status `json:"status,omitempty"`
	IsFallback   *bool          `json:"is_fallback,omitempty"`
	Page         int            `json:"page"`
	PageSize     int            `json:"page_size"`
}

// routingUseCase 路由业务逻辑实现
type routingUseCase struct {
	routingRuleRepo      port.MessageRoutingRuleRepository
	messageProcessorRepo port.MessageProcessorRepository
	channelRepo          port.ChannelRepository
	logger               logger.Interface
}

// NewRoutingUseCase 创建路由业务逻辑实例
func NewRoutingUseCase(
	routingRuleRepo port.MessageRoutingRuleRepository,
	messageProcessorRepo port.MessageProcessorRepository,
	channelRepo port.ChannelRepository,
	logger logger.Interface,
) RoutingUseCase {
	return &routingUseCase{
		routingRuleRepo:      routingRuleRepo,
		messageProcessorRepo: messageProcessorRepo,
		channelRepo:          channelRepo,
		logger:               logger,
	}
}

// RouteMessage 根据路由规则处理消息
func (uc *routingUseCase) RouteMessage(ctx context.Context, message *entity.Message) ([]*entity.MessageProcessor, error) {
	uc.logger.Info("开始路由消息", "message_id", message.MessageID, "channel_id", message.ChannelID)

	// 获取匹配的路由规则
	rules, err := uc.GetMatchingRules(ctx, message)
	if err != nil {
		uc.logger.Error("获取匹配路由规则失败", "error", err)
		return nil, fmt.Errorf("获取匹配路由规则失败: %w", err)
	}

	if len(rules) == 0 {
		uc.logger.Warn("没有找到匹配的路由规则", "message_id", message.MessageID)
		return nil, fmt.Errorf("没有找到匹配的路由规则")
	}

	// 根据优先级排序规则
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	// 获取对应的处理器
	processors := make([]*entity.MessageProcessor, 0, len(rules))
	for _, rule := range rules {
		processor, err := uc.messageProcessorRepo.GetByID(ctx, rule.ProcessorID)
		if err != nil {
			uc.logger.Error("获取处理器失败", "processor_id", rule.ProcessorID, "error", err)
			continue
		}

		if !processor.IsActive() {
			uc.logger.Warn("处理器未激活", "processor_id", rule.ProcessorID)
			continue
		}

		processors = append(processors, processor)

		// 如果不是广播模式，只取第一个匹配的处理器
		if rule.RouteType != entity.RouteTypeBroadcast {
			break
		}
	}

	uc.logger.Info("消息路由完成", "message_id", message.MessageID, "processor_count", len(processors))
	return processors, nil
}

// GetMatchingRules 获取匹配的路由规则
func (uc *routingUseCase) GetMatchingRules(ctx context.Context, message *entity.Message) ([]*entity.MessageRoutingRule, error) {
	// 获取通道信息
	channel, err := uc.channelRepo.GetByID(ctx, message.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("获取通道信息失败: %w", err)
	}

	// 获取所有激活的路由规则
	activeStatus := entity.StatusActive
	filters := RoutingRuleFilters{
		Status: &activeStatus,
	}
	allRules, err := uc.GetRoutingRules(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("获取路由规则失败: %w", err)
	}

	// 过滤匹配的规则
	var matchingRules []*entity.MessageRoutingRule
	var fallbackRules []*entity.MessageRoutingRule

	for _, rule := range allRules {
		if uc.matchesRule(rule, message, channel) {
			if rule.IsFallback {
				fallbackRules = append(fallbackRules, rule)
			} else {
				matchingRules = append(matchingRules, rule)
			}
		}
	}

	// 如果没有匹配的常规规则，使用兜底规则
	if len(matchingRules) == 0 && len(fallbackRules) > 0 {
		matchingRules = fallbackRules
	}

	return matchingRules, nil
}

// matchesRule 检查消息是否匹配路由规则
func (uc *routingUseCase) matchesRule(rule *entity.MessageRoutingRule, message *entity.Message, channel *entity.Channel) bool {
	// 检查Bot ID匹配
	if !rule.MatchesBotID(channel.BotID) {
		return false
	}

	// 检查平台类型匹配
	if !rule.MatchesPlatformType(channel.PlatformType) {
		return false
	}

	// 检查通道ID匹配
	if !rule.MatchesChannelID(message.ChannelID) {
		return false
	}

	// 检查消息类型匹配
	if !rule.MatchesMessageType(message.MessageType) {
		return false
	}

	// 检查发送者模式匹配
	if !uc.matchesSenderPatterns(rule, message) {
		return false
	}

	// 检查内容模式匹配
	if !uc.matchesContentPatterns(rule, message) {
		return false
	}

	return true
}

// matchesSenderPatterns 检查发送者模式匹配
func (uc *routingUseCase) matchesSenderPatterns(rule *entity.MessageRoutingRule, message *entity.Message) bool {
	if rule.SenderPatterns == nil {
		return true
	}

	patterns, ok := rule.SenderPatterns["patterns"]
	if !ok {
		return true
	}

	if patternList, ok := patterns.([]interface{}); ok {
		for _, pattern := range patternList {
			if patternStr, ok := pattern.(string); ok {
				// 支持正则表达式匹配
				if matched, _ := regexp.MatchString(patternStr, message.SenderID); matched {
					return true
				}
				// 支持简单字符串匹配
				if strings.Contains(message.SenderID, patternStr) {
					return true
				}
			}
		}
		return false // 如果有模式但都不匹配，返回false
	}

	return true
}

// matchesContentPatterns 检查内容模式匹配
func (uc *routingUseCase) matchesContentPatterns(rule *entity.MessageRoutingRule, message *entity.Message) bool {
	if rule.ContentPatterns == nil {
		return true
	}

	patterns, ok := rule.ContentPatterns["patterns"]
	if !ok {
		return true
	}

	logic, _ := rule.ContentPatterns["logic"].(string)
	if logic == "" {
		logic = "OR" // 默认为OR逻辑
	}

	if patternList, ok := patterns.([]interface{}); ok {
		matchCount := 0
		for _, pattern := range patternList {
			if patternStr, ok := pattern.(string); ok {
				matched := false
				// 支持正则表达式匹配
				if matched, _ = regexp.MatchString(patternStr, message.Content); matched {
					matchCount++
				} else if strings.Contains(message.Content, patternStr) {
					// 支持简单字符串匹配
					matched = true
					matchCount++
				}

				// OR逻辑：只要有一个匹配就返回true
				if logic == "OR" && matched {
					return true
				}
			}
		}

		// AND逻辑：所有模式都要匹配
		if logic == "AND" {
			return matchCount == len(patternList)
		}

		// OR逻辑但没有匹配到任何模式
		return false
	}

	return true
}

// CreateRoutingRule 创建路由规则
func (uc *routingUseCase) CreateRoutingRule(ctx context.Context, rule *entity.MessageRoutingRule) error {
	uc.logger.Info("创建路由规则", "rule_name", rule.RuleName)

	// 验证数据
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("路由规则验证失败: %w", err)
	}

	// 检查处理器是否存在
	processor, err := uc.messageProcessorRepo.GetByID(ctx, rule.ProcessorID)
	if err != nil {
		return fmt.Errorf("处理器不存在: %w", err)
	}

	if !processor.IsActive() {
		return fmt.Errorf("处理器未激活")
	}

	// 创建路由规则
	if err := uc.routingRuleRepo.Create(ctx, rule); err != nil {
		return fmt.Errorf("创建路由规则失败: %w", err)
	}

	uc.logger.Info("路由规则创建成功", "rule_id", rule.ID)
	return nil
}

// UpdateRoutingRule 更新路由规则
func (uc *routingUseCase) UpdateRoutingRule(ctx context.Context, rule *entity.MessageRoutingRule) error {
	uc.logger.Info("更新路由规则", "rule_id", rule.ID)

	// 验证数据
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("路由规则验证失败: %w", err)
	}

	// 检查规则是否存在
	existingRule, err := uc.routingRuleRepo.GetByID(ctx, rule.ID)
	if err != nil {
		return fmt.Errorf("路由规则不存在: %w", err)
	}

	// 检查处理器是否存在
	processor, err := uc.messageProcessorRepo.GetByID(ctx, rule.ProcessorID)
	if err != nil {
		return fmt.Errorf("处理器不存在: %w", err)
	}

	if !processor.IsActive() {
		return fmt.Errorf("处理器未激活")
	}

	// 保留一些不应更改的字段
	rule.CreatedAt = existingRule.CreatedAt

	// 更新路由规则
	if err := uc.routingRuleRepo.Update(ctx, rule); err != nil {
		return fmt.Errorf("更新路由规则失败: %w", err)
	}

	uc.logger.Info("路由规则更新成功", "rule_id", rule.ID)
	return nil
}

// DeleteRoutingRule 删除路由规则
func (uc *routingUseCase) DeleteRoutingRule(ctx context.Context, ruleID string) error {
	uc.logger.Info("删除路由规则", "rule_id", ruleID)

	// 检查规则是否存在
	_, err := uc.routingRuleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("路由规则不存在: %w", err)
	}

	// 软删除路由规则
	if err := uc.routingRuleRepo.Delete(ctx, ruleID); err != nil {
		return fmt.Errorf("删除路由规则失败: %w", err)
	}

	uc.logger.Info("路由规则删除成功", "rule_id", ruleID)
	return nil
}

// GetRoutingRules 获取路由规则列表
func (uc *routingUseCase) GetRoutingRules(ctx context.Context, filters RoutingRuleFilters) ([]*entity.MessageRoutingRule, error) {
	// 构建查询条件
	conditions := make(map[string]interface{})

	if filters.ProcessorID != nil {
		conditions["processor_id"] = *filters.ProcessorID
	}

	if filters.Status != nil {
		conditions["status"] = *filters.Status
	}

	if filters.IsFallback != nil {
		conditions["is_fallback"] = *filters.IsFallback
	}

	// 设置分页参数
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	}

	// 构建查询参数
	listParams := port.ListParams{
		Page:     filters.Page,
		PageSize: filters.PageSize,
		Filters:  conditions,
		OrderBy:  "priority ASC, created_at DESC",
	}

	// 查询路由规则
	result, err := uc.routingRuleRepo.List(ctx, listParams)
	if err != nil {
		return nil, fmt.Errorf("查询路由规则失败: %w", err)
	}

	return result.Items, nil
}
