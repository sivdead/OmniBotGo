package usecase

import (
	"context"
	"fmt"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// processorUC 处理器用例实现
type processorUC struct {
	processorRepo   port.MessageProcessorRepository
	routingRuleRepo port.MessageRoutingRuleRepository
	logger          logger.Interface
}

// NewProcessorUseCase 创建处理器用例
func NewProcessorUseCase(
	processorRepo port.MessageProcessorRepository,
	routingRuleRepo port.MessageRoutingRuleRepository,
	logger logger.Interface,
) ProcessorUseCase {
	return &processorUC{
		processorRepo:   processorRepo,
		routingRuleRepo: routingRuleRepo,
		logger:          logger,
	}
}

// CreateProcessor 创建处理器
func (uc *processorUC) CreateProcessor(ctx context.Context, req CreateProcessorRequest) (*entity.MessageProcessor, error) {
	// 检查名称是否已存在
	exists, err := uc.processorRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		uc.logger.Error("failed to check processor existence", "error", err, "name", req.Name)
		return nil, fmt.Errorf("failed to check processor existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("processor with name %s already exists", req.Name)
	}

	// 创建处理器实体
	processor := &entity.MessageProcessor{
		ProcessorName: req.Name,
		ProcessorType: req.Type,
		Config:        req.Config,
		Priority:      req.Priority,
		Status:        entity.StatusActive,
	}

	// 保存到数据库
	if err := uc.processorRepo.Create(ctx, processor); err != nil {
		uc.logger.Error("failed to create processor", "error", err, "name", req.Name)
		return nil, fmt.Errorf("failed to create processor: %w", err)
	}

	uc.logger.Info("processor created", "name", processor.ProcessorName, "type", processor.ProcessorType)
	return processor, nil
}

// ListProcessors 获取处理器列表
func (uc *processorUC) ListProcessors(ctx context.Context, params ListProcessorsParams) (*ProcessorListResult, error) {
	// 构建查询参数
	filters := make(map[string]interface{})
	if params.Type != nil {
		filters["processor_type"] = *params.Type
	}
	if params.Status != nil {
		filters["status"] = *params.Status
	}
	if params.IsEnabled != nil {
		filters["status"] = entity.StatusActive // IsEnabled映射到Status
	}

	listParams := port.ListParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		Filters:  filters,
		OrderBy:  "priority DESC, created_at DESC",
	}

	// 查询数据
	result, err := uc.processorRepo.List(ctx, listParams)
	if err != nil {
		uc.logger.Error("failed to list processors", "error", err)
		return nil, fmt.Errorf("failed to list processors: %w", err)
	}

	// 转换结果
	items := make([]entity.MessageProcessor, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &ProcessorListResult{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}, nil
}

// GetProcessor 获取处理器详情
func (uc *processorUC) GetProcessor(ctx context.Context, id string) (*entity.MessageProcessor, error) {
	processor, err := uc.processorRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("failed to get processor", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get processor: %w", err)
	}
	return processor, nil
}

// UpdateProcessor 更新处理器
func (uc *processorUC) UpdateProcessor(ctx context.Context, req UpdateProcessorRequest) (*entity.MessageProcessor, error) {
	// 获取现有处理器
	processor, err := uc.processorRepo.GetByID(ctx, req.ID)
	if err != nil {
		uc.logger.Error("failed to get processor for update", "error", err, "id", req.ID)
		return nil, fmt.Errorf("failed to get processor: %w", err)
	}

	// 如果要更新名称，检查新名称是否已存在
	if req.Name != nil && *req.Name != processor.ProcessorName {
		exists, err := uc.processorRepo.ExistsByName(ctx, *req.Name)
		if err != nil {
			uc.logger.Error("failed to check processor name existence", "error", err, "name", *req.Name)
			return nil, fmt.Errorf("failed to check processor name existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("processor with name %s already exists", *req.Name)
		}
		processor.ProcessorName = *req.Name
	}

	// 更新其他字段
	if req.Config != nil {
		processor.Config = req.Config
	}
	if req.Priority != nil {
		processor.Priority = *req.Priority
	}
	if req.IsEnabled != nil {
		if *req.IsEnabled {
			processor.Status = entity.StatusActive
		} else {
			processor.Status = entity.StatusInactive
		}
	}
	if req.Status != nil {
		processor.Status = *req.Status
	}

	// 保存更新
	if err := uc.processorRepo.Update(ctx, processor); err != nil {
		uc.logger.Error("failed to update processor", "error", err, "id", req.ID)
		return nil, fmt.Errorf("failed to update processor: %w", err)
	}

	uc.logger.Info("processor updated", "id", processor.ID, "name", processor.ProcessorName)
	return processor, nil
}

// DeleteProcessor 删除处理器
func (uc *processorUC) DeleteProcessor(ctx context.Context, id string) error {
	// 检查是否有关联的路由规则
	rules, err := uc.routingRuleRepo.GetByProcessorID(ctx, id)
	if err != nil {
		uc.logger.Error("failed to check routing rules", "error", err, "processorID", id)
		return fmt.Errorf("failed to check routing rules: %w", err)
	}
	if len(rules) > 0 {
		return fmt.Errorf("cannot delete processor with existing routing rules")
	}

	// 删除处理器
	if err := uc.processorRepo.Delete(ctx, id); err != nil {
		uc.logger.Error("failed to delete processor", "error", err, "id", id)
		return fmt.Errorf("failed to delete processor: %w", err)
	}

	uc.logger.Info("processor deleted", "id", id)
	return nil
}

// UpdateProcessorStatus 更新处理器状态
func (uc *processorUC) UpdateProcessorStatus(ctx context.Context, req UpdateProcessorStatusRequest) error {
	// 获取处理器
	processor, err := uc.processorRepo.GetByID(ctx, req.ID)
	if err != nil {
		uc.logger.Error("failed to get processor for status update", "error", err, "id", req.ID)
		return fmt.Errorf("failed to get processor: %w", err)
	}

	// 更新状态
	processor.Status = req.Status
	if err := uc.processorRepo.Update(ctx, processor); err != nil {
		uc.logger.Error("failed to update processor status", "error", err, "id", req.ID)
		return fmt.Errorf("failed to update processor status: %w", err)
	}

	uc.logger.Info("processor status updated", "id", req.ID, "status", req.Status)
	return nil
}

// CreateRoutingRule 创建路由规则
func (uc *processorUC) CreateRoutingRule(ctx context.Context, req CreateRoutingRuleRequest) (*entity.MessageRoutingRule, error) {
	// 验证处理器是否存在
	exists, err := uc.processorRepo.Exists(ctx, req.ProcessorID)
	if err != nil {
		uc.logger.Error("failed to check processor existence", "error", err, "processorID", req.ProcessorID)
		return nil, fmt.Errorf("failed to check processor existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("processor with ID %d not found", req.ProcessorID)
	}

	// 创建路由规则实体
	rule := &entity.MessageRoutingRule{
		ProcessorID: req.ProcessorID,
		RuleName:    req.RuleName,
		Priority:    req.Priority,
		IsFallback:  req.IsFallback,
		Status:      entity.StatusActive,
	}

	// 设置平台类型
	if req.PlatformType != "" {
		rule.PlatformTypes = entity.JSONField{
			"types": []string{req.PlatformType},
		}
	}

	// 设置消息类型
	if req.MessageType != "" {
		rule.MessageTypes = entity.JSONField{
			"types": []string{req.MessageType},
		}
	}

	// 设置条件
	if req.Conditions != nil {
		// 将条件存储在适当的字段中
		if patterns, ok := req.Conditions["content_patterns"]; ok {
			rule.ContentPatterns = entity.JSONField{"patterns": patterns}
		}
		if patterns, ok := req.Conditions["sender_patterns"]; ok {
			rule.SenderPatterns = entity.JSONField{"patterns": patterns}
		}
	}

	// 保存到数据库
	if err := uc.routingRuleRepo.Create(ctx, rule); err != nil {
		uc.logger.Error("failed to create routing rule", "error", err, "name", req.RuleName)
		return nil, fmt.Errorf("failed to create routing rule: %w", err)
	}

	uc.logger.Info("routing rule created", "name", rule.RuleName, "processorID", rule.ProcessorID)
	return rule, nil
}

// ListRoutingRules 获取路由规则列表
func (uc *processorUC) ListRoutingRules(ctx context.Context, processorID string, params ListRoutingRulesParams) ([]*entity.MessageRoutingRule, error) {
	// 如果指定了处理器ID，直接使用专门的方法
	if processorID != "" {
		rules, err := uc.routingRuleRepo.GetByProcessorID(ctx, processorID)
		if err != nil {
			uc.logger.Error("failed to get routing rules by processor", "error", err, "processorID", processorID)
			return nil, fmt.Errorf("failed to get routing rules: %w", err)
		}
		return rules, nil
	}

	// 否则使用通用列表方法
	filters := make(map[string]interface{})
	// 注意：由于PlatformTypes和MessageTypes是JSON字段，过滤可能需要特殊处理
	if params.Status != nil {
		filters["status"] = *params.Status
	}
	if params.IsFallback != nil {
		filters["is_fallback"] = *params.IsFallback
	}

	listParams := port.ListParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		Filters:  filters,
		OrderBy:  "priority DESC, created_at DESC",
	}

	result, err := uc.routingRuleRepo.List(ctx, listParams)
	if err != nil {
		uc.logger.Error("failed to list routing rules", "error", err)
		return nil, fmt.Errorf("failed to list routing rules: %w", err)
	}

	return result.Items, nil
}

// UpdateRoutingRule 更新路由规则
func (uc *processorUC) UpdateRoutingRule(ctx context.Context, req UpdateRoutingRuleRequest) (*entity.MessageRoutingRule, error) {
	// 获取现有规则
	rule, err := uc.routingRuleRepo.GetByID(ctx, req.ID)
	if err != nil {
		uc.logger.Error("failed to get routing rule for update", "error", err, "id", req.ID)
		return nil, fmt.Errorf("failed to get routing rule: %w", err)
	}

	// 如果要更新处理器ID，验证新处理器是否存在
	if req.ProcessorID != nil && *req.ProcessorID != rule.ProcessorID {
		exists, err := uc.processorRepo.Exists(ctx, *req.ProcessorID)
		if err != nil {
			uc.logger.Error("failed to check processor existence", "error", err, "processorID", *req.ProcessorID)
			return nil, fmt.Errorf("failed to check processor existence: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("processor with ID %s not found", *req.ProcessorID)
		}
		rule.ProcessorID = *req.ProcessorID
	}

	// 更新其他字段
	if req.RuleName != nil {
		rule.RuleName = *req.RuleName
	}

	// 更新平台类型
	if req.PlatformType != nil {
		if *req.PlatformType == "" {
			rule.PlatformTypes = nil
		} else {
			rule.PlatformTypes = entity.JSONField{
				"types": []string{*req.PlatformType},
			}
		}
	}

	// 更新消息类型
	if req.MessageType != nil {
		if *req.MessageType == "" {
			rule.MessageTypes = nil
		} else {
			rule.MessageTypes = entity.JSONField{
				"types": []string{*req.MessageType},
			}
		}
	}

	// 更新条件
	if req.Conditions != nil {
		if patterns, ok := req.Conditions["content_patterns"]; ok {
			rule.ContentPatterns = entity.JSONField{"patterns": patterns}
		}
		if patterns, ok := req.Conditions["sender_patterns"]; ok {
			rule.SenderPatterns = entity.JSONField{"patterns": patterns}
		}
	}

	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if req.IsFallback != nil {
		rule.IsFallback = *req.IsFallback
	}
	if req.Status != nil {
		rule.Status = *req.Status
	}

	// 保存更新
	if err := uc.routingRuleRepo.Update(ctx, rule); err != nil {
		uc.logger.Error("failed to update routing rule", "error", err, "id", req.ID)
		return nil, fmt.Errorf("failed to update routing rule: %w", err)
	}

	uc.logger.Info("routing rule updated", "id", rule.ID, "name", rule.RuleName)
	return rule, nil
}

// DeleteRoutingRule 删除路由规则
func (uc *processorUC) DeleteRoutingRule(ctx context.Context, ruleID string) error {
	// 删除规则
	if err := uc.routingRuleRepo.Delete(ctx, ruleID); err != nil {
		uc.logger.Error("failed to delete routing rule", "error", err, "id", ruleID)
		return fmt.Errorf("failed to delete routing rule: %w", err)
	}

	uc.logger.Info("routing rule deleted", "id", ruleID)
	return nil
}
