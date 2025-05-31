// Package entity defines main entities for business logic (services), data base mapping and
// HTTP response objects if suitable. Each logic group entities in own file.
package entity

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"gorm.io/gorm"
)

// BaseEntity 包含所有数据库表的通用字段
type BaseEntity struct {
	ID        int64          `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index;comment:软删除时间"`
}

// JSONField 通用JSON字段类型，用于存储JSON格式的配置和数据
type JSONField map[string]interface{}

// Scan 实现 sql.Scanner 接口，用于从数据库读取JSON数据
func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONField", value)
	}
}

// Value 实现 driver.Valuer 接口，用于向数据库写入JSON数据
func (j JSONField) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// String 返回JSON字符串表示
func (j JSONField) String() string {
	data, _ := json.Marshal(j)
	return string(data)
}

// Get 获取JSON字段中的值
func (j JSONField) Get(key string) interface{} {
	if j == nil {
		return nil
	}
	return j[key]
}

// Set 设置JSON字段中的值
func (j JSONField) Set(key string, value interface{}) {
	if j == nil {
		j = make(map[string]interface{})
	}
	j[key] = value
}

// GetString 获取字符串类型的值
func (j JSONField) GetString(key string) string {
	if val := j.Get(key); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt 获取整数类型的值
func (j JSONField) GetInt(key string) int {
	if val := j.Get(key); val != nil {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}

// GetBool 获取布尔类型的值
func (j JSONField) GetBool(key string) bool {
	if val := j.Get(key); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// ValidationError 验证错误类型
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 实现error接口
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError 创建新的验证错误
func NewValidationError(field, message string) error {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewJSONField 从JSON字符串创建JSONField
func NewJSONField(jsonStr string) JSONField {
	var field JSONField
	if err := json.Unmarshal([]byte(jsonStr), &field); err != nil {
		// 如果解析失败，返回空的JSONField
		return make(JSONField)
	}
	return field
}
