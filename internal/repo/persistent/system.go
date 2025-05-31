package persistent

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/repo"
	"github.com/sivdead/OmniBotGo/pkg/database"
)

// SystemConfigRepo SystemConfig相关的数据访问层实现
type SystemConfigRepo struct {
	*BaseRepo
}

// NewSystemConfigRepo 创建SystemConfig Repository实例
func NewSystemConfigRepo(db database.CommonDB) repo.SystemConfigRepo {
	return &SystemConfigRepo{
		BaseRepo: NewBaseRepo(db),
	}
}

// Create 创建新的SystemConfig
func (r *SystemConfigRepo) Create(ctx context.Context, config *entity.SystemConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	if err := r.db.GetGORM().WithContext(ctx).Create(config).Error; err != nil {
		return r.handleError(err, "create system config")
	}
	return nil
}

// GetByKey 根据配置键获取SystemConfig
func (r *SystemConfigRepo) GetByKey(ctx context.Context, key string) (*entity.SystemConfig, error) {
	var config entity.SystemConfig
	err := r.db.GetGORM().WithContext(ctx).
		Where("config_key = ?", key).
		First(&config).Error
	if err != nil {
		return nil, r.handleError(err, "get system config by key")
	}
	return &config, nil
}

// GetByGroup 根据配置分组获取SystemConfig列表
func (r *SystemConfigRepo) GetByGroup(ctx context.Context, group string) ([]*entity.SystemConfig, error) {
	var configs []*entity.SystemConfig
	err := r.db.GetGORM().WithContext(ctx).
		Where("config_group = ?", group).
		Order("config_key ASC").
		Find(&configs).Error
	if err != nil {
		return nil, r.handleError(err, "get system configs by group")
	}
	return configs, nil
}

// GetAllByGroup 根据多个配置分组获取SystemConfig列表
func (r *SystemConfigRepo) GetAllByGroup(ctx context.Context, groups []string) ([]*entity.SystemConfig, error) {
	var configs []*entity.SystemConfig
	err := r.db.GetGORM().WithContext(ctx).
		Where("config_group IN ?", groups).
		Order("config_group ASC, config_key ASC").
		Find(&configs).Error
	if err != nil {
		return nil, r.handleError(err, "get system configs by groups")
	}
	return configs, nil
}

// GetUserEditableConfigs 获取用户可编辑的配置
func (r *SystemConfigRepo) GetUserEditableConfigs(ctx context.Context) ([]*entity.SystemConfig, error) {
	var configs []*entity.SystemConfig
	err := r.db.GetGORM().WithContext(ctx).
		Where("is_system = ?", false).
		Order("config_group ASC, config_key ASC").
		Find(&configs).Error
	if err != nil {
		return nil, r.handleError(err, "get user editable configs")
	}
	return configs, nil
}

// GetSystemConfigs 获取系统配置
func (r *SystemConfigRepo) GetSystemConfigs(ctx context.Context) ([]*entity.SystemConfig, error) {
	var configs []*entity.SystemConfig
	err := r.db.GetGORM().WithContext(ctx).
		Where("is_system = ?", true).
		Order("config_group ASC, config_key ASC").
		Find(&configs).Error
	if err != nil {
		return nil, r.handleError(err, "get system configs")
	}
	return configs, nil
}

// Update 更新SystemConfig
func (r *SystemConfigRepo) Update(ctx context.Context, config *entity.SystemConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	result := r.db.GetGORM().WithContext(ctx).Save(config)
	if result.Error != nil {
		return r.handleError(result.Error, "update system config")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateValue 更新配置值
func (r *SystemConfigRepo) UpdateValue(ctx context.Context, key, value string) error {
	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.SystemConfig{}).
		Where("config_key = ?", key).
		Update("config_value", value)

	if result.Error != nil {
		return r.handleError(result.Error, "update config value")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除SystemConfig（软删除）
func (r *SystemConfigRepo) Delete(ctx context.Context, id int64) error {
	return r.softDelete(ctx, &entity.SystemConfig{}, id)
}

// List 获取SystemConfig列表（分页）
func (r *SystemConfigRepo) List(ctx context.Context, params repo.ListParams) (*repo.PaginatedResult, error) {
	params = r.validateParams(params)

	var configs []*entity.SystemConfig
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.SystemConfig{}), params)

	result, err := r.paginate(ctx, query, params, &configs)
	if err != nil {
		return nil, r.handleError(err, "list system configs")
	}

	result.Items = configs
	return result, nil
}

// Exists 检查SystemConfig是否存在
func (r *SystemConfigRepo) Exists(ctx context.Context, key string) (bool, error) {
	return r.exists(ctx, &entity.SystemConfig{}, "config_key = ?", key)
}
