package usecase

import (
	"context"
	"fmt"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// systemConfigUC 系统配置用例实现
type systemConfigUC struct {
	repo   port.SystemConfigRepository
	logger logger.Interface
}

// NewSystemConfigUseCase 创建系统配置用例
func NewSystemConfigUseCase(
	repo port.SystemConfigRepository,
	logger logger.Interface,
) SystemConfigUseCase {
	return &systemConfigUC{
		repo:   repo,
		logger: logger,
	}
}

// CreateSystemConfig 创建系统配置
func (uc *systemConfigUC) CreateSystemConfig(ctx context.Context, req CreateSystemConfigRequest) (*entity.SystemConfig, error) {
	// 检查配置项是否已存在
	exists, err := uc.repo.Exists(ctx, req.Key)
	if err != nil {
		uc.logger.Error("failed to check system config existence", "error", err, "key", req.Key)
		return nil, fmt.Errorf("failed to check system config existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("system config with key %s already exists", req.Key)
	}

	// 创建配置实体
	config := &entity.SystemConfig{
		ConfigKey:   req.Key,
		ConfigValue: req.Value,
		ConfigType:  req.Type,
		ConfigGroup: req.Group,
		Description: req.Description,
		IsEncrypted: req.IsEncrypted,
		IsSystem:    req.IsSystem,
	}

	// 保存到数据库
	if err := uc.repo.Create(ctx, config); err != nil {
		uc.logger.Error("failed to create system config", "error", err, "key", req.Key)
		return nil, fmt.Errorf("failed to create system config: %w", err)
	}

	uc.logger.Info("system config created", "key", config.ConfigKey, "group", config.ConfigGroup)
	return config, nil
}

// GetSystemConfig 获取系统配置
func (uc *systemConfigUC) GetSystemConfig(ctx context.Context, key string) (*entity.SystemConfig, error) {
	config, err := uc.repo.GetByKey(ctx, key)
	if err != nil {
		uc.logger.Error("failed to get system config", "error", err, "key", key)
		return nil, fmt.Errorf("failed to get system config: %w", err)
	}
	return config, nil
}

// UpdateSystemConfig 更新系统配置
func (uc *systemConfigUC) UpdateSystemConfig(ctx context.Context, req UpdateSystemConfigRequest) (*entity.SystemConfig, error) {
	// 获取现有配置
	config, err := uc.repo.GetByKey(ctx, req.Key)
	if err != nil {
		uc.logger.Error("failed to get system config for update", "error", err, "key", req.Key)
		return nil, fmt.Errorf("failed to get system config: %w", err)
	}

	// 系统配置不允许修改某些字段
	if config.IsSystem && (req.Type != nil || req.IsEncrypted != nil) {
		return nil, fmt.Errorf("cannot modify type or encryption setting of system config")
	}

	// 更新字段
	if req.Value != nil {
		config.ConfigValue = *req.Value
	}
	if req.Type != nil && !config.IsSystem {
		config.ConfigType = *req.Type
	}
	if req.Group != nil {
		config.ConfigGroup = *req.Group
	}
	if req.Description != nil {
		config.Description = *req.Description
	}
	if req.IsEncrypted != nil && !config.IsSystem {
		config.IsEncrypted = *req.IsEncrypted
	}

	// 保存更新
	if err := uc.repo.Update(ctx, config); err != nil {
		uc.logger.Error("failed to update system config", "error", err, "key", req.Key)
		return nil, fmt.Errorf("failed to update system config: %w", err)
	}

	uc.logger.Info("system config updated", "key", config.ConfigKey)
	return config, nil
}

// DeleteSystemConfig 删除系统配置
func (uc *systemConfigUC) DeleteSystemConfig(ctx context.Context, key string) error {
	// 获取配置检查是否是系统配置
	config, err := uc.repo.GetByKey(ctx, key)
	if err != nil {
		uc.logger.Error("failed to get system config for deletion", "error", err, "key", key)
		return fmt.Errorf("failed to get system config: %w", err)
	}

	if config.IsSystem {
		return fmt.Errorf("cannot delete system config")
	}

	// 删除配置
	if err := uc.repo.Delete(ctx, config.ID); err != nil {
		uc.logger.Error("failed to delete system config", "error", err, "key", key)
		return fmt.Errorf("failed to delete system config: %w", err)
	}

	uc.logger.Info("system config deleted", "key", key)
	return nil
}

// ListSystemConfigs 获取系统配置列表
func (uc *systemConfigUC) ListSystemConfigs(ctx context.Context, params ListSystemConfigsParams) (*SystemConfigListResult, error) {
	// 构建查询参数
	filters := make(map[string]interface{})
	if params.Group != nil {
		filters["group"] = *params.Group
	}
	if params.Key != nil {
		filters["key"] = *params.Key
	}
	if params.IsSystem != nil {
		filters["is_system"] = *params.IsSystem
	}

	listParams := port.ListParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		Filters:  filters,
		OrderBy:  "group, key",
	}

	// 查询数据
	result, err := uc.repo.List(ctx, listParams)
	if err != nil {
		uc.logger.Error("failed to list system configs", "error", err)
		return nil, fmt.Errorf("failed to list system configs: %w", err)
	}

	// 转换结果
	items := make([]entity.SystemConfig, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &SystemConfigListResult{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}, nil
}

// GetSystemConfigsByGroup 根据组获取系统配置
func (uc *systemConfigUC) GetSystemConfigsByGroup(ctx context.Context, group string) ([]*entity.SystemConfig, error) {
	configs, err := uc.repo.GetByGroup(ctx, group)
	if err != nil {
		uc.logger.Error("failed to get system configs by group", "error", err, "group", group)
		return nil, fmt.Errorf("failed to get system configs by group: %w", err)
	}
	return configs, nil
}
