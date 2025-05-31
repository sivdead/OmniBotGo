package persistent

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/repo"
	"github.com/sivdead/OmniBotGo/pkg/database"
)

// MessageProcessorRepo MessageProcessor相关的数据访问层实现
type MessageProcessorRepo struct {
	*BaseRepo
}

// NewMessageProcessorRepo 创建MessageProcessor Repository实例
func NewMessageProcessorRepo(db database.CommonDB) repo.MessageProcessorRepo {
	return &MessageProcessorRepo{
		BaseRepo: NewBaseRepo(db),
	}
}

// Create 创建新的MessageProcessor
func (r *MessageProcessorRepo) Create(ctx context.Context, processor *entity.MessageProcessor) error {
	if err := processor.Validate(); err != nil {
		return fmt.Errorf("processor validation failed: %w", err)
	}

	if err := r.db.GetGORM().WithContext(ctx).Create(processor).Error; err != nil {
		return r.handleError(err, "create processor")
	}
	return nil
}

// GetByID 根据ID获取MessageProcessor
func (r *MessageProcessorRepo) GetByID(ctx context.Context, id int64) (*entity.MessageProcessor, error) {
	var processor entity.MessageProcessor
	err := r.db.GetGORM().WithContext(ctx).
		Preload("RoutingRules").
		Preload("APICallLogs").
		First(&processor, id).Error
	if err != nil {
		return nil, r.handleError(err, "get processor by id")
	}
	return &processor, nil
}

// GetByName 根据名称获取MessageProcessor
func (r *MessageProcessorRepo) GetByName(ctx context.Context, name string) (*entity.MessageProcessor, error) {
	var processor entity.MessageProcessor
	err := r.db.GetGORM().WithContext(ctx).
		Preload("RoutingRules").
		Where("processor_name = ?", name).
		First(&processor).Error
	if err != nil {
		return nil, r.handleError(err, "get processor by name")
	}
	return &processor, nil
}

// GetByType 根据类型获取MessageProcessor列表
func (r *MessageProcessorRepo) GetByType(ctx context.Context, processorType string) ([]*entity.MessageProcessor, error) {
	var processors []*entity.MessageProcessor
	err := r.db.GetGORM().WithContext(ctx).
		Where("processor_type = ?", processorType).
		Find(&processors).Error
	if err != nil {
		return nil, r.handleError(err, "get processors by type")
	}
	return processors, nil
}

// Update 更新MessageProcessor
func (r *MessageProcessorRepo) Update(ctx context.Context, processor *entity.MessageProcessor) error {
	if err := processor.Validate(); err != nil {
		return fmt.Errorf("processor validation failed: %w", err)
	}

	result := r.db.GetGORM().WithContext(ctx).Save(processor)
	if result.Error != nil {
		return r.handleError(result.Error, "update processor")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除MessageProcessor（软删除）
func (r *MessageProcessorRepo) Delete(ctx context.Context, id int64) error {
	return r.softDelete(ctx, &entity.MessageProcessor{}, id)
}

// List 获取MessageProcessor列表（分页）
func (r *MessageProcessorRepo) List(ctx context.Context, params repo.ListParams) (*repo.PaginatedResult, error) {
	params = r.validateParams(params)

	var processors []*entity.MessageProcessor
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.MessageProcessor{}), params)

	result, err := r.paginate(ctx, query, params, &processors)
	if err != nil {
		return nil, r.handleError(err, "list processors")
	}

	result.Items = processors
	return result, nil
}

// ListActive 获取所有激活状态的MessageProcessor
func (r *MessageProcessorRepo) ListActive(ctx context.Context) ([]*entity.MessageProcessor, error) {
	var processors []*entity.MessageProcessor
	err := r.db.GetGORM().WithContext(ctx).
		Where("status = ?", entity.StatusActive).
		Find(&processors).Error
	if err != nil {
		return nil, r.handleError(err, "list active processors")
	}
	return processors, nil
}

// ListByPriority 根据优先级获取MessageProcessor列表
func (r *MessageProcessorRepo) ListByPriority(ctx context.Context) ([]*entity.MessageProcessor, error) {
	var processors []*entity.MessageProcessor
	err := r.db.GetGORM().WithContext(ctx).
		Where("status = ?", entity.StatusActive).
		Order("priority ASC, id ASC").
		Find(&processors).Error
	if err != nil {
		return nil, r.handleError(err, "list processors by priority")
	}
	return processors, nil
}

// Exists 检查MessageProcessor是否存在
func (r *MessageProcessorRepo) Exists(ctx context.Context, id int64) (bool, error) {
	return r.exists(ctx, &entity.MessageProcessor{}, "id = ?", id)
}

// ExistsByName 检查指定名称的MessageProcessor是否存在
func (r *MessageProcessorRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	return r.exists(ctx, &entity.MessageProcessor{}, "processor_name = ?", name)
}

// MessageRoutingRuleRepo MessageRoutingRule相关的数据访问层实现
type MessageRoutingRuleRepo struct {
	*BaseRepo
}

// NewMessageRoutingRuleRepo 创建MessageRoutingRule Repository实例
func NewMessageRoutingRuleRepo(db database.CommonDB) repo.MessageRoutingRuleRepo {
	return &MessageRoutingRuleRepo{
		BaseRepo: NewBaseRepo(db),
	}
}

// Create 创建新的MessageRoutingRule
func (r *MessageRoutingRuleRepo) Create(ctx context.Context, rule *entity.MessageRoutingRule) error {
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("rule validation failed: %w", err)
	}

	if err := r.db.GetGORM().WithContext(ctx).Create(rule).Error; err != nil {
		return r.handleError(err, "create routing rule")
	}
	return nil
}

// GetByID 根据ID获取MessageRoutingRule
func (r *MessageRoutingRuleRepo) GetByID(ctx context.Context, id int64) (*entity.MessageRoutingRule, error) {
	var rule entity.MessageRoutingRule
	err := r.db.GetGORM().WithContext(ctx).
		Preload("MessageProcessor").
		First(&rule, id).Error
	if err != nil {
		return nil, r.handleError(err, "get routing rule by id")
	}
	return &rule, nil
}

// GetByProcessorID 根据处理器ID获取MessageRoutingRule列表
func (r *MessageRoutingRuleRepo) GetByProcessorID(ctx context.Context, processorID int64) ([]*entity.MessageRoutingRule, error) {
	var rules []*entity.MessageRoutingRule
	err := r.db.GetGORM().WithContext(ctx).
		Where("processor_id = ?", processorID).
		Order("priority ASC, id ASC").
		Find(&rules).Error
	if err != nil {
		return nil, r.handleError(err, "get routing rules by processor id")
	}
	return rules, nil
}

// GetActiveRules 获取所有激活的MessageRoutingRule
func (r *MessageRoutingRuleRepo) GetActiveRules(ctx context.Context) ([]*entity.MessageRoutingRule, error) {
	var rules []*entity.MessageRoutingRule
	err := r.db.GetGORM().WithContext(ctx).
		Preload("MessageProcessor").
		Where("status = ?", entity.StatusActive).
		Order("priority ASC, id ASC").
		Find(&rules).Error
	if err != nil {
		return nil, r.handleError(err, "get active routing rules")
	}
	return rules, nil
}

// GetMatchingRules 获取匹配指定条件的MessageRoutingRule列表
func (r *MessageRoutingRuleRepo) GetMatchingRules(ctx context.Context, botID int64, platformType string, channelID int64, messageType string) ([]*entity.MessageRoutingRule, error) {
	var rules []*entity.MessageRoutingRule

	// 获取所有激活的规则，在应用层进行匹配逻辑
	// 这里可以进行一些基础的数据库级过滤，复杂的匹配逻辑在业务层处理
	err := r.db.GetGORM().WithContext(ctx).
		Preload("MessageProcessor").
		Where("status = ?", entity.StatusActive).
		Order("priority ASC, id ASC").
		Find(&rules).Error
	if err != nil {
		return nil, r.handleError(err, "get matching routing rules")
	}

	return rules, nil
}

// GetFallbackRules 获取兜底规则
func (r *MessageRoutingRuleRepo) GetFallbackRules(ctx context.Context) ([]*entity.MessageRoutingRule, error) {
	var rules []*entity.MessageRoutingRule
	err := r.db.GetGORM().WithContext(ctx).
		Preload("MessageProcessor").
		Where("status = ? AND is_fallback = ?", entity.StatusActive, true).
		Order("priority ASC, id ASC").
		Find(&rules).Error
	if err != nil {
		return nil, r.handleError(err, "get fallback routing rules")
	}
	return rules, nil
}

// Update 更新MessageRoutingRule
func (r *MessageRoutingRuleRepo) Update(ctx context.Context, rule *entity.MessageRoutingRule) error {
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("rule validation failed: %w", err)
	}

	result := r.db.GetGORM().WithContext(ctx).Save(rule)
	if result.Error != nil {
		return r.handleError(result.Error, "update routing rule")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除MessageRoutingRule（软删除）
func (r *MessageRoutingRuleRepo) Delete(ctx context.Context, id int64) error {
	return r.softDelete(ctx, &entity.MessageRoutingRule{}, id)
}

// List 获取MessageRoutingRule列表（分页）
func (r *MessageRoutingRuleRepo) List(ctx context.Context, params repo.ListParams) (*repo.PaginatedResult, error) {
	params = r.validateParams(params)

	var rules []*entity.MessageRoutingRule
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.MessageRoutingRule{}), params)

	result, err := r.paginate(ctx, query, params, &rules)
	if err != nil {
		return nil, r.handleError(err, "list routing rules")
	}

	result.Items = rules
	return result, nil
}

// ListByPriority 根据优先级获取MessageRoutingRule列表
func (r *MessageRoutingRuleRepo) ListByPriority(ctx context.Context) ([]*entity.MessageRoutingRule, error) {
	var rules []*entity.MessageRoutingRule
	err := r.db.GetGORM().WithContext(ctx).
		Preload("MessageProcessor").
		Where("status = ?", entity.StatusActive).
		Order("priority ASC, id ASC").
		Find(&rules).Error
	if err != nil {
		return nil, r.handleError(err, "list routing rules by priority")
	}
	return rules, nil
}

// Exists 检查MessageRoutingRule是否存在
func (r *MessageRoutingRuleRepo) Exists(ctx context.Context, id int64) (bool, error) {
	return r.exists(ctx, &entity.MessageRoutingRule{}, "id = ?", id)
}
