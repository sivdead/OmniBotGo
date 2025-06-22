package persistent

import (
	"context"
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"

	"github.com/sivdead/OmniBotGo/internal/repo"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/database"
)

// BaseRepo 基础Repository，提供通用的数据库操作方法
type BaseRepo struct {
	db database.CommonDB
}

// NewBaseRepo 创建基础Repository实例
func NewBaseRepo(db database.CommonDB) *BaseRepo {
	return &BaseRepo{db: db}
}

// WithTx 在事务中执行操作
func (r *BaseRepo) WithTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.GetGORM().WithContext(ctx).Transaction(fn)
}

// buildQuery 构建查询条件
func (r *BaseRepo) buildQuery(db *gorm.DB, params repo.ListParams) *gorm.DB {
	query := db

	// 应用过滤条件
	for key, value := range params.Filters {
		if value == nil {
			continue
		}

		switch key {
		case "id":
			query = query.Where("id = ?", value)
		case "status":
			query = query.Where("status = ?", value)
		case "created_at_gte":
			query = query.Where("created_at >= ?", value)
		case "created_at_lte":
			query = query.Where("created_at <= ?", value)
		case "updated_at_gte":
			query = query.Where("updated_at >= ?", value)
		case "updated_at_lte":
			query = query.Where("updated_at <= ?", value)
		default:
			// 对于其他字段，使用LIKE查询
			if str, ok := value.(string); ok && str != "" {
				if strings.Contains(key, "_like") {
					field := strings.TrimSuffix(key, "_like")
					query = query.Where(fmt.Sprintf("%s LIKE ?", field), "%"+str+"%")
				} else {
					query = query.Where(fmt.Sprintf("%s = ?", key), value)
				}
			}
		}
	}

	// 应用排序
	if params.OrderBy != "" {
		// 防止SQL注入，只允许特定字段排序
		allowedFields := []string{"id", "created_at", "updated_at", "name", "priority", "status"}
		orderBy := strings.ToLower(params.OrderBy)

		// 解析排序方向
		direction := "ASC"
		if strings.HasSuffix(orderBy, " desc") {
			direction = "DESC"
			orderBy = strings.TrimSuffix(orderBy, " desc")
		} else if strings.HasSuffix(orderBy, " asc") {
			orderBy = strings.TrimSuffix(orderBy, " asc")
		}

		// 验证字段名
		isAllowed := false
		for _, field := range allowedFields {
			if orderBy == field {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			query = query.Order(fmt.Sprintf("%s %s", orderBy, direction))
		} else {
			// 默认排序
			query = query.Order("id DESC")
		}
	} else {
		// 默认排序
		query = query.Order("id DESC")
	}

	return query
}

// PaginateTyped 执行泛型分页查询（独立函数）
func PaginateTyped[T any](db *gorm.DB, ctx context.Context, query *gorm.DB, params repo.ListParams, result *[]T) (*repo.PaginatedResult[T], error) {
	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// 计算分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	offset := (params.Page - 1) * params.PageSize
	totalPages := int(math.Ceil(float64(total) / float64(params.PageSize)))

	// 执行分页查询
	if err := query.Offset(offset).Limit(params.PageSize).Find(result).Error; err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}

	return &repo.PaginatedResult[T]{
		Items:      *result,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// paginate 执行分页查询（原始版本，保持向后兼容）
func (r *BaseRepo) paginate(ctx context.Context, query *gorm.DB, params repo.ListParams, result interface{}) (*repo.LegacyPaginatedResult, error) {
	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// 计算分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	offset := (params.Page - 1) * params.PageSize
	totalPages := int(math.Ceil(float64(total) / float64(params.PageSize)))

	// 执行分页查询
	if err := query.Offset(offset).Limit(params.PageSize).Find(result).Error; err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}

	return &repo.LegacyPaginatedResult{
		Items:      result,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// exists 检查记录是否存在
func (r *BaseRepo) exists(ctx context.Context, model interface{}, where string, args ...interface{}) (bool, error) {
	var count int64
	err := r.db.GetGORM().WithContext(ctx).Model(model).Where(where, args...).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return count > 0, nil
}

// softDelete 软删除记录
func (r *BaseRepo) softDelete(ctx context.Context, model interface{}, id int64) error {
	result := r.db.GetGORM().WithContext(ctx).Delete(model, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// hardDelete 硬删除记录
func (r *BaseRepo) hardDelete(ctx context.Context, model interface{}, where string, args ...interface{}) error {
	result := r.db.GetGORM().WithContext(ctx).Unscoped().Where(where, args...).Delete(model)
	if result.Error != nil {
		return fmt.Errorf("failed to hard delete record: %w", result.Error)
	}
	return nil
}

// validateParams 验证查询参数
func (r *BaseRepo) validateParams(params repo.ListParams) repo.ListParams {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	if params.Filters == nil {
		params.Filters = make(map[string]interface{})
	}
	return params
}

// isNotFound 检查错误是否为记录不存在
func (r *BaseRepo) isNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

// handleError 统一处理数据库错误
func (r *BaseRepo) handleError(err error, operation string) error {
	if err == nil {
		return nil
	}

	if r.isNotFound(err) {
		return err // 直接返回，上层处理
	}

	return fmt.Errorf("%s failed: %w", operation, err)
}

// convertToInternalParams 转换port.ListParams到repo.ListParams
func convertToInternalParams(params port.ListParams) repo.ListParams {
	return repo.ListParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		OrderBy:  params.OrderBy,
		Filters:  params.Filters,
	}
}

// PaginateTypedForPort 执行泛型分页查询并返回port.PaginatedResult
func PaginateTypedForPort[T any](db *gorm.DB, ctx context.Context, query *gorm.DB, params repo.ListParams, result *[]T) (*port.PaginatedResult[T], error) {
	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// 计算分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	offset := (params.Page - 1) * params.PageSize
	totalPages := int(math.Ceil(float64(total) / float64(params.PageSize)))

	// 执行分页查询
	if err := query.Offset(offset).Limit(params.PageSize).Find(result).Error; err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}

	return &port.PaginatedResult[T]{
		Items:      *result,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}
