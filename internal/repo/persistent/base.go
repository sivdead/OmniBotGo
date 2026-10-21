package persistent

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/database"
	"gorm.io/gorm"
)

const (
	filterMessageStatus = "message_status"
	filterPlatformType  = "platform_type"
	filterQueueName     = "queue_name"
	filterConfigKey     = "config_key"
)

var errUnknownFilter = fmt.Errorf("unknown query filter")

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

// buildQuery 构建查询条件 (E05: enums, time ranges, unknown filters error).
func (r *BaseRepo) buildQuery(db *gorm.DB, params port.ListParams) (*gorm.DB, error) {
	query := db

	for key, value := range params.Filters {
		if value == nil {
			continue
		}

		next, err := applyListFilter(query, key, value)
		if err != nil {
			return nil, err
		}

		query = next
	}

	return applyListOrder(query, params.OrderBy), nil
}

//nolint:gocyclo,cyclop,goconst,funlen // explicit allowlisted filter switch; short SQL identifiers
func applyListFilter(query *gorm.DB, key string, value interface{}) (*gorm.DB, error) {
	switch key {
	case "id":
		return query.Where("id = ?", value), nil
	case "status":
		return query.Where("status = ?", normalizeFilterValue(value)), nil
	case filterMessageStatus:
		return query.Where("message_status = ?", normalizeFilterValue(value)), nil
	case "direction":
		return query.Where("direction = ?", normalizeFilterValue(value)), nil
	case "channel_id":
		return query.Where("channel_id = ?", value), nil
	case "processor_id":
		return query.Where("processor_id = ?", value), nil
	case "conversation_id":
		return query.Where("conversation_id = ?", value), nil
	case filterPlatformType:
		return query.Where("platform_type = ?", value), nil
	case filterQueueName:
		return query.Where("queue_name = ?", value), nil
	case filterConfigKey:
		return query.Where("config_key = ?", value), nil
	case "config_group":
		return query.Where("config_group = ?", value), nil
	case "created_at_gte":
		return query.Where("created_at >= ?", value), nil
	case "created_at_lte":
		return query.Where("created_at <= ?", value), nil
	case "updated_at_gte":
		return query.Where("updated_at >= ?", value), nil
	case "updated_at_lte":
		return query.Where("updated_at <= ?", value), nil
	case "start_time":
		return query.Where("created_at >= ?", value), nil
	case "end_time":
		return query.Where("created_at <= ?", value), nil
	default:
		if strings.HasSuffix(key, "_like") {
			field := strings.TrimSuffix(key, "_like")
			if !isAllowedFilterField(field) {
				return nil, fmt.Errorf("%w: %s", errUnknownFilter, key)
			}

			str, ok := value.(string)
			if !ok || str == "" {
				return query, nil
			}

			return query.Where(fmt.Sprintf("%s LIKE ?", field), "%"+str+"%"), nil
		}

		return nil, fmt.Errorf("%w: %s", errUnknownFilter, key)
	}
}

func applyListOrder(query *gorm.DB, orderByRaw string) *gorm.DB {
	if orderByRaw == "" {
		return query.Order("id DESC")
	}

	const orderFieldName = "name"
	allowedFields := []string{"id", "created_at", "updated_at", orderFieldName, "priority", "status"}
	orderBy := strings.ToLower(orderByRaw)

	direction := "ASC"
	if strings.HasSuffix(orderBy, " desc") {
		direction = "DESC"
		orderBy = strings.TrimSuffix(orderBy, " desc")
	} else if strings.HasSuffix(orderBy, " asc") {
		orderBy = strings.TrimSuffix(orderBy, " asc")
	}

	for _, field := range allowedFields {
		if orderBy == field {
			return query.Order(fmt.Sprintf("%s %s", orderBy, direction))
		}
	}

	return query.Order("id DESC")
}

func isAllowedFilterField(field string) bool {
	switch field {
	case "id", "status", "name", "bot_name", "channel_name", "processor_name",
		"rule_name", filterConfigKey, "message_id", filterPlatformType, filterQueueName:
		return true
	default:
		return false
	}
}

//nolint:gocyclo,cyclop // enum type switch
func normalizeFilterValue(value interface{}) interface{} {
	switch v := value.(type) {
	case entity.MessageStatus:
		return int8(v)
	case *entity.MessageStatus:
		return enumPtrInt8(v)
	case entity.MessageDirection:
		return int8(v)
	case *entity.MessageDirection:
		return enumPtrInt8(v)
	case entity.Status:
		return int8(v)
	case *entity.Status:
		return enumPtrInt8(v)
	case entity.QueueStatus:
		return int8(v)
	case *entity.QueueStatus:
		return enumPtrInt8(v)
	case entity.ConnectionStatus:
		return int8(v)
	case *entity.ConnectionStatus:
		return enumPtrInt8(v)
	default:
		return value
	}
}

func enumPtrInt8[T ~int8](v *T) interface{} {
	if v == nil {
		return nil
	}

	return int8(*v)
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
func (r *BaseRepo) softDelete(ctx context.Context, model interface{}, id string) error {
	result := r.db.GetGORM().WithContext(ctx).Delete(model, "id = ?", id)
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
func (r *BaseRepo) validateParams(params port.ListParams) port.ListParams {
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

// Paginate 执行泛型分页查询并返回分页结果
func Paginate[T any](db *gorm.DB, ctx context.Context, query *gorm.DB, params port.ListParams, result *[]T) (*port.PaginatedResult[T], error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

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
