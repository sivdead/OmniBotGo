package persistent

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/mysql"
)

// TranslationRepo -.
type TranslationRepo struct {
	*mysql.MySQL
}

// New -.
func New(db *mysql.MySQL) *TranslationRepo {
	return &TranslationRepo{db}
}

// GetHistory - 使用GORM进行简单查询
func (r *TranslationRepo) GetHistory(ctx context.Context) ([]entity.Translation, error) {
	var entities []entity.Translation

	err := r.DB.WithContext(ctx).Find(&entities).Error
	if err != nil {
		return nil, fmt.Errorf("TranslationRepo - GetHistory - r.DB.Find: %w", err)
	}

	return entities, nil
}

// Store - 使用GORM进行简单的创建操作
func (r *TranslationRepo) Store(ctx context.Context, t entity.Translation) error {
	err := r.DB.WithContext(ctx).Create(&t).Error
	if err != nil {
		return fmt.Errorf("TranslationRepo - Store - r.DB.Create: %w", err)
	}

	return nil
}

// GetHistoryWithComplexConditions - 使用Squirrel进行复杂查询
// 示例：根据多个动态条件查询翻译历史，包含分页和排序
func (r *TranslationRepo) GetHistoryWithComplexConditions(ctx context.Context, conditions map[string]interface{}, limit, offset int, orderBy string) ([]entity.Translation, error) {
	query := r.Builder.
		Select("id", "source", "destination", "original", "translation", "created_at", "updated_at").
		From("translations")

	// 动态添加查询条件
	for key, value := range conditions {
		switch key {
		case "source":
			query = query.Where("source = ?", value)
		case "destination":
			query = query.Where("destination = ?", value)
		case "original_like":
			query = query.Where("original LIKE ?", "%"+value.(string)+"%")
		case "created_after":
			query = query.Where("created_at > ?", value)
		case "created_before":
			query = query.Where("created_at < ?", value)
		}
	}

	// 添加排序和分页
	if orderBy != "" {
		query = query.OrderBy(orderBy)
	} else {
		query = query.OrderBy("created_at DESC")
	}

	if limit > 0 {
		query = query.Limit(uint64(limit))
	}

	if offset > 0 {
		query = query.Offset(uint64(offset))
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("TranslationRepo - GetHistoryWithComplexConditions - query.ToSql: %w", err)
	}

	rows, err := r.SqlDB.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("TranslationRepo - GetHistoryWithComplexConditions - r.SqlDB.QueryContext: %w", err)
	}
	defer rows.Close()

	var entities []entity.Translation
	for rows.Next() {
		var e entity.Translation
		err := rows.Scan(&e.ID, &e.Source, &e.Destination, &e.Original, &e.Translation, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("TranslationRepo - GetHistoryWithComplexConditions - rows.Scan: %w", err)
		}
		entities = append(entities, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("TranslationRepo - GetHistoryWithComplexConditions - rows.Err: %w", err)
	}

	return entities, nil
}

// GetTranslationStatistics - 使用Squirrel进行统计查询
// 示例：获取翻译统计信息，按源语言和目标语言分组
func (r *TranslationRepo) GetTranslationStatistics(ctx context.Context) ([]map[string]interface{}, error) {
	sql, args, err := r.Builder.
		Select(
			"source",
			"destination", 
			"COUNT(*) as total_count",
			"COUNT(DISTINCT original) as unique_originals",
			"AVG(LENGTH(original)) as avg_original_length",
			"MAX(created_at) as last_translation_time",
		).
		From("translations").
		GroupBy("source", "destination").
		OrderBy("total_count DESC").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("TranslationRepo - GetTranslationStatistics - query.ToSql: %w", err)
	}

	rows, err := r.SqlDB.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("TranslationRepo - GetTranslationStatistics - r.SqlDB.QueryContext: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var source, destination string
		var totalCount, uniqueOriginals int
		var avgOriginalLength float64
		var lastTranslationTime interface{}

		err := rows.Scan(&source, &destination, &totalCount, &uniqueOriginals, &avgOriginalLength, &lastTranslationTime)
		if err != nil {
			return nil, fmt.Errorf("TranslationRepo - GetTranslationStatistics - rows.Scan: %w", err)
		}

		result := map[string]interface{}{
			"source":                   source,
			"destination":              destination,
			"total_count":              totalCount,
			"unique_originals":         uniqueOriginals,
			"avg_original_length":      avgOriginalLength,
			"last_translation_time":    lastTranslationTime,
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("TranslationRepo - GetTranslationStatistics - rows.Err: %w", err)
	}

	return results, nil
}

// BatchInsertTranslations - 使用GORM进行批量插入
func (r *TranslationRepo) BatchInsertTranslations(ctx context.Context, translations []entity.Translation) error {
	err := r.DB.WithContext(ctx).CreateInBatches(translations, 100).Error
	if err != nil {
		return fmt.Errorf("TranslationRepo - BatchInsertTranslations - r.DB.CreateInBatches: %w", err)
	}

	return nil
}

// UpdateTranslationsWithCondition - 使用Squirrel进行条件更新
func (r *TranslationRepo) UpdateTranslationsWithCondition(ctx context.Context, updateFields map[string]interface{}, whereConditions map[string]interface{}) error {
	query := r.Builder.Update("translations")

	// 添加更新字段
	for field, value := range updateFields {
		query = query.Set(field, value)
	}

	// 添加WHERE条件
	for field, value := range whereConditions {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("TranslationRepo - UpdateTranslationsWithCondition - query.ToSql: %w", err)
	}

	_, err = r.SqlDB.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("TranslationRepo - UpdateTranslationsWithCondition - r.SqlDB.ExecContext: %w", err)
	}

	return nil
} 